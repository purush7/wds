package queue

import (
	"alert_system/constants"
	"time"
)

type QueueConfig struct {
	Name          string
	NumConsumers  int
	PrefetchLimit int64
	PollDuration  time.Duration
	IsActive      bool

	// BatchingEnabled bool
	// BatchSize       int64
	// BatchTimeout    time.Duration
}

func GetNewQueueConfig(name string, numConsumers int, prefetchLimit int64, pollDuration time.Duration, isActive bool) QueueConfig {
	return QueueConfig{
		Name:          name,
		NumConsumers:  numConsumers,
		PrefetchLimit: prefetchLimit,
		PollDuration:  pollDuration,
		IsActive:      isActive,
	}
}

func GetAllQueueWithConfig() (queues []QueueConfig) {

	//This Queue contains the payload of file_path
	batcherQueue := GetNewQueueConfig(constants.BatcherQueue, constants.NumConsumers, constants.PrefetchLimit, constants.PollDuration, true)
	queues = append(queues, batcherQueue)

	//This Queue contains the topics of user_id, alert_message
	notifierQueue := GetNewQueueConfig(constants.NotifierQueue, constants.NumConsumers, constants.PrefetchLimit, constants.PollDuration, true)
	queues = append(queues, notifierQueue)

	return
}
