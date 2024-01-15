package producer

import (
	"alert_system/alert_initiator/internal/queue"
	"encoding/json"
	"log"
)

func PublishIntoQueues(kind string, payload interface{}) error {
	queueObj, err := queue.GetOpenQueue(kind)
	if err != nil {
		log.Printf("error: %v while connecting to queue: %s\n", err, kind)
		return err
	}
	dataBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("error: %v while marshaling the payload in queue: %s\n", err, kind)
		return err
	}
	return queueObj.Queue.PublishBytes(dataBytes)
}
