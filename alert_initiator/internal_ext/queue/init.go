package queue

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"alert_system/infra/redis"

	"github.com/adjust/rmq/v5"
)

var (
	connection rmq.Connection
	once       sync.Once
)

func InitializeConnection() (err error) {
	once.Do(func() {
		errChan := make(chan error, 10)
		go logErrors(errChan)
		connection, err = rmq.OpenConnectionWithRedisClient("swilly_queue", redis.GetClient(), errChan)
		if err != nil {
			log.Print("Unable to create redis connection for queues: ", err)
		}
	})

	return err
}

func GetConnection() rmq.Connection {
	return connection
}

func logErrors(errChan <-chan error) {
	for err := range errChan {
		switch err := err.(type) {
		case *rmq.HeartbeatError:
			if err.Count == rmq.HeartbeatErrorLimit {
				log.Print("heartbeat error (limit): ", err)
			} else {
				log.Print("heartbeat error: ", err)
			}
		case *rmq.ConsumeError:
			log.Print("consume error: ", err)
		case *rmq.DeliveryError:
			log.Print("delivery error: ", err.Delivery, err)
		default:
			log.Print("other error: ", err)
		}
	}
}

// waitforQueueTerminationSignal implements Usecase.
func WaitforQueueTerminationSignal() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT)
	defer signal.Stop(signals)

	<-signals // wait for signal
	go func() {
		<-signals // hard exit on second signal (in case shutdown gets stuck)
		os.Exit(1)
	}()

	// Check if there are any running workers
	log.Println("Termination request received")

	// wait for all Consume() calls to finish
	<-GetConnection().StopAllConsuming()
	log.Println("Service gracefully terminated")
}
