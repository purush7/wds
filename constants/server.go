package constants

import "time"

const (
	SERVER_MODE_HTTP_SERVER string = "httpserver"
	SERVER_MODE_WORKER      string = "worker"
)

const (
	WORKER_DEFAULT_COUNT int64 = 1
)

// Consumer
const (
	PrefetchLimit = 4
	PollDuration  = 10 * time.Millisecond
	NumConsumers  = 1
	// BatchSize     = 10
	// BatchTimeout  = time.Second

	// ConsumeDuration = time.Second

	ConsumerNameString = "%v_consumer_%d"
)

// Queue, Names
const (
	BatcherQueue       = "queue_batcher"
	NotifierQueue      = "queue_notifier"
	RedisDbStoreKey    = "db_store"
	QueueSizeThershold = int64(4)
)

// Notifier
const (
	NOTIFIER_URL = "http://alert-notifier:3333/webhook"
)

// Redis keys
const (
	TooManyRequests     = "tooManyRequests"
	TooManyRequestsTrue = "1"
	Publishing          = "publishing"
	PublishingTrue      = "1"
)

// Payload
const (
	BatchLimit int64 = 2
)

// redis // error constants
const (
	ErrStringRequestProcessing = "request in processing state"
	// redis lock time in seconds
	DefaultRedisLockTime = 1
)
