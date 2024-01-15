package monitor

import (
	"alert_system/alert_initiator/internal/entities"
	"alert_system/alert_initiator/internal/services/producer"
	"alert_system/constants"
	"log"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
)

const monitoredFolder = "/tmp"        // Specify the path to the monitored folder
const checkInterval = 1 * time.Second // Adjust the interval for checking if the file is open

var fileEventChan = make(chan string, 10) // Channel to store file creation events
var queueChan = make(chan string, 10)     // Channel to simulate a queue

var mutex sync.Mutex // Mutex to synchronize access to the fileEventChan

func monitorFolder() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Add the folder to the watcher
	if err := watcher.Add(monitoredFolder); err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// Process only file creation events
			if event.Op&fsnotify.Create == fsnotify.Create {
				// Store the absolute path of the file in the channel
				absPath, err := filepath.Abs(event.Name)
				if err != nil {
					log.Printf("Error getting absolute path: %v\n", err)
					continue
				}

				mutex.Lock()
				fileEventChan <- absPath
				mutex.Unlock()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Error:", err)
		}
	}
}

func checkFileOpenStatus() {
	for {
		// Wait for the specified interval
		time.Sleep(checkInterval)

		// Check the files in the channel
		mutex.Lock()
		for len(fileEventChan) > 0 {
			filePath := <-fileEventChan

			// Check if the file is open using lsof
			if isFileOpen(filePath) {
				log.Printf("File %s is currently open.\n", filePath)
				fileEventChan <- filePath
			} else {
				// File is not open, push the path to the queue channel
				queueChan <- filePath
				log.Printf("File %s is not open. Added to the queue.\n", filePath)
			}
		}
		mutex.Unlock()
	}
}

func simulateQueueProcessing() {
	for {
		// Wait for the specified interval
		time.Sleep(checkInterval)

		// Check the queue channel
		for len(queueChan) > 0 {
			filePath := <-queueChan
			err := producer.PublishIntoQueues(constants.BatcherQueue, entities.BatcherPayload{
				FilePath: filePath,
				Id:       uuid.New().String(),
			})
			if err != nil {
				log.Printf(" error while triggering notification system of filepath: %s, :%v\n", filePath, err)
			} else {
				log.Printf("pushed the filepath: %s \n", filePath)
			}
		}
	}
}

func isFileOpen(filePath string) bool {
	// Attempt to open the file with read-only access
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0)
	if err != nil {
		if os.IsPermission(err) {
			// Permission error means the file is open
			return true
		}
	} else {
		// File opened successfully, close it
		file.Close()
		return false
	}

	// Use fcntl syscall to get file status flags
	fd := int(file.Fd())
	flags, _, err := syscall.Syscall(syscall.SYS_FCNTL, uintptr(fd), syscall.F_GETFL, 0)
	if err != nil {
		log.Printf("Error getting file status flags: %v\n", err)
		return true // Assume file is open on error
	}

	// Check if the file is open
	return flags&syscall.O_ACCMODE != syscall.O_WRONLY
}
