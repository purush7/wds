package consumer

import (
	"alert_system/alert_initiator/internal/services/worker"
	"fmt"
	"log"
	"time"

	"github.com/adjust/rmq/v5"
)

type notifierConsumer struct {
	name      string
	queueName string
	count     int
	before    time.Time
}

func Notifier(tag int, queueName string) *notifierConsumer {
	return &notifierConsumer{
		name:      fmt.Sprintf("consumer_%d", tag),
		queueName: queueName,
		count:     0,
		before:    time.Now(),
	}
}

func (consumer *notifierConsumer) Consume(delivery rmq.Delivery) {
	payload := delivery.Payload()
	log.Printf("start consumer for queue: %v with payload: %s", consumer.queueName, payload)
	// handler
	handlerErr := worker.ProcessNotifierTopic([]byte(payload))
	if handlerErr != nil {
		if err := delivery.Reject(); err != nil {
			log.Printf("failed to reject %s: %s which got handle error: %s", payload, err, handlerErr)
		} else {
			log.Printf("rejected %s because of error: %s", payload, handlerErr)
		}
	} else { // reject one per batch
		if err := delivery.Ack(); err != nil {
			log.Printf("failed to ack %s: %s", payload, err)
		} else {
			log.Printf("acked %s", payload)
		}
	}
}
