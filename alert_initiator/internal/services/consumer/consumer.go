package consumer

import (
	"alert_system/alert_initiator/internal/queue"
	"alert_system/constants"
	"fmt"
	"log"

	"github.com/adjust/rmq/v5"
)

func getConsumer(num int, kind string) rmq.Consumer {
	switch kind {
	case constants.BatcherQueue:
		return Batcher(num, kind)
	case constants.NotifierQueue:
		return Notifier(num, kind)
	}

	return nil
}

func StartConsumers() (errors []error) {

	// Queues
	queueConfigs := queue.GetAllQueueWithConfig()

	for _, config := range queueConfigs {
		if !config.IsActive {
			log.Printf("Inactive queue %v, not initialized.", config.Name)
			continue
		}
		connection := queue.GetConnection()
		queue, OpenQueueErr := connection.OpenQueue(config.Name)
		if OpenQueueErr != nil {
			log.Printf("Unable to open queue (%v), error: %v", config.Name, OpenQueueErr)
			errors = append(errors, OpenQueueErr)
			continue
		}

		if startErr := queue.StartConsuming(config.PrefetchLimit, config.PollDuration); startErr != nil {
			log.Printf("Unable to start consumer for queue (%v), error: %v", config.Name, startErr)
			errors = append(errors, startErr)
			continue
		}

		for i := 0; i < config.NumConsumers; i++ {
			consumerName := fmt.Sprintf(constants.ConsumerNameString, config.Name, i)
			newConsumer := getConsumer(i, config.Name)
			if newConsumer == nil {
				log.Printf("Unable to get a consumer for queue (%v)", config.Name)
				continue
			}
			if _, addConsumerErr := queue.AddConsumer(consumerName, newConsumer); addConsumerErr != nil {
				log.Printf("Unable to add consumer for queue (%v), error: %v", config.Name, addConsumerErr)
				errors = append(errors, addConsumerErr)
				break
			}

		}

	}

	return
}
