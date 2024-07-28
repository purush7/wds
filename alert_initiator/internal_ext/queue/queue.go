package queue

import (
	"log"
	"math"

	"github.com/adjust/rmq/v5"
)

type OpenQueue struct {
	Name  string
	Queue rmq.Queue
}

func GetOpenQueue(name string) (openQueue OpenQueue, err error) {
	openQueue.Name = name
	connection := GetConnection()
	queue, err := connection.OpenQueue(name)
	if err != nil {
		log.Printf("failed to open queue (%v): %s", name, err)
		return
	}
	openQueue.Queue = queue

	return
}

// Publish - Producer for the queue
func (op *OpenQueue) Publish(payload string) {
	if publishErr := op.Queue.Publish(payload); publishErr != nil {
		log.Printf("failed to publish in queue (%v): %s", op.Name, publishErr)
	}
}

func (op *OpenQueue) GetRejectedCount() int64 {
	rejected, err := op.Queue.ReturnRejected(math.MaxInt64)
	if err != nil {
		log.Printf("failed to return rejected for queue (%v): %s", op.Name, err)
		return 0
	}
	return rejected
}

func (op *OpenQueue) GetUnAckedCount() int64 {
	unacked, err := op.Queue.ReturnUnacked(math.MaxInt64)
	if err != nil {
		log.Printf("failed to return UnAcked for queue (%v): %s", op.Name, err)
		return 0
	}
	return unacked
}

func (op *OpenQueue) Purge() int64 {
	count, err := op.Queue.PurgeReady()
	if err != nil {
		log.Printf("failed to purge for queue (%v): %s", op.Name, err)
		return 0
	}
	return count
}

func CleanQueues() {
	connection := GetConnection()
	cleaner := rmq.NewCleaner(connection)
	returned, err := cleaner.Clean()
	if err != nil {
		log.Printf("failed to clean: %s", err)
		return
	}
	log.Printf("cleaned %d", returned)
}
