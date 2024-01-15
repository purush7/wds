package monitor

import (
	"alert_system/alert_initiator/internal/entities"
	"alert_system/alert_initiator/internal/model"
	"alert_system/alert_initiator/internal/queue"
	"alert_system/alert_initiator/internal/services/producer"
	"alert_system/constants"
	"alert_system/infra/redis"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/adjust/rmq/v5"
	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
)

func MontiorService() {
	go MontiorQueueService()
	go MontiorDbService()
	MontiorFileService2()
}

func MontiorQueueService() {
	queueConfigs := queue.GetAllQueueWithConfig()
	queueNames := make([]string, len(queueConfigs))
	rmqConn := queue.GetConnection()
	for index := range queueConfigs {
		queueNames[index] = queueConfigs[index].Name
	}
	for {
		tooManyRequests, err := redis.Get(constants.TooManyRequests)
		if err != nil && err != redis.Nil {
			//alert
			log.Println(err)
			return
		}

		if tooManyRequests == constants.TooManyRequestsTrue {
			time.Sleep(time.Second)
			continue
		}

		stats, err := rmq.CollectStats(queueNames, rmqConn)
		if err != nil {
			//alert
			log.Printf("error in monitor queue service while getting the queue stats %s\n", err)
			return
		}
		for queueName, queueStat := range stats.QueueStats {
			// send alerts if the ready count or reject count cross the size
			if queueStat.ReadyCount > constants.QueueSizeThershold {
				log.Printf("size of queue: %s crossed the thershold, current: %d and threshold: %d", queueName, queueStat.ReadyCount, constants.QueueSizeThershold)
			}
			if queueStat.RejectedCount > 0 {
				log.Printf("rejected size of queue: %s is %d", queueName, queueStat.RejectedCount)
				rmqQueue, err := queue.GetOpenQueue(queueName)
				if err != nil {
					log.Println(err)
				}
				n, err := rmqQueue.Queue.ReturnRejected(constants.BatchLimit)
				if err != nil {
					log.Println(err)
				}
				log.Printf("%d tasks are made ready from rejected state of queue: %s\n", n, queueName)
			}
		}
		//sleep for a second
		time.Sleep(time.Second)
	}
}

func MontiorDbService() {
	for {
		for {
			tooManyRequests, err := redis.Get(constants.TooManyRequests)
			if err != nil && err != redis.Nil {
				//alert
				log.Println("error in monitor db service:", err)
				return
			}

			if tooManyRequests != constants.TooManyRequestsTrue {
				getPayloadFromDb()
			}
			time.Sleep(time.Second)
		}
	}
}

func getPayloadFromDb() {
	//get from db
	payloadBatch := model.GetBatchFromDb()
	err := redis.Set(constants.Publishing, constants.PublishingTrue, time.Second)
	if err != nil {
		log.Println("err while locking ", err)
	}
	for index := range payloadBatch {
		tooManyRequests, err := redis.Get(constants.TooManyRequests)
		if err != nil && err != redis.Nil {
			//alert
			log.Println(err)
			return
		}

		if tooManyRequests == constants.TooManyRequestsTrue {
			break
		}
		err = producer.PublishIntoQueues(constants.BatcherQueue, entities.BatcherPayload{
			FilePath: payloadBatch[index].FilePath,
			Id:       payloadBatch[index].Id,
		})
		if err != nil {
			//alert
			log.Printf(" error while triggering notification system of filepath: %s, :%v\n", payloadBatch[index].FilePath, err)
			return
		}
		model.Delete(payloadBatch[index].Id)
	}

	err = redis.Delete(constants.Publishing)
	if err != nil {
		log.Println("err while deleting the lock ", err)
	}
	//sleep as can't allocate resource always to this service
	time.Sleep(time.Second)
}

func MontiorFileService() {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Specify the folder to watch
	folderPath := "/tmp"
	err = watcher.Add(folderPath)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("New file:", event.Name)
				err = producer.PublishIntoQueues(constants.BatcherQueue, entities.BatcherPayload{
					FilePath: event.Name,
					Id:       uuid.New().String(),
				})
				if err != nil {
					log.Printf(" error while triggering notification system of filepath: %s, :%v\n", event.Name, err)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Error:", err)
		}
	}
}

func MontiorFileService2() {
	// Start file system event monitoring
	go monitorFolder()

	// Start checking if the files are open periodically
	go checkFileOpenStatus()

	// Start simulating pushing to a queue
	go simulateQueueProcessing()
}

func scanDirectory(directoryPath string) error {
	// Open the directory
	dir, err := os.Open(directoryPath)
	if err != nil {
		return err
	}
	defer dir.Close()

	// Read the files in the directory
	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	// Process each file
	for _, fileInfo := range fileInfos {
		if fileInfo.Mode().IsRegular() {
			// Check if the file is new or has been modified recently
			if time.Since(fileInfo.ModTime()) < 5*time.Second {
				filePath := filepath.Join(directoryPath, fileInfo.Name())
				log.Printf("New or modified file detected: %s\n", filePath)

				// Process the new file (you can add your logic here)

				// Note: You may want to avoid processing the same file multiple times.
				// You could keep track of processed files in a data structure and skip
				// files that have already been processed.
			}
		}
	}

	return nil
}
