package main

import (
	"alert_system/alert_initiator/internal/handler"
	"alert_system/alert_initiator/internal/queue"
	"alert_system/alert_initiator/internal/services/consumer"
	"alert_system/alert_initiator/internal/services/monitor"
	"alert_system/constants"
	"flag"
	"fmt"
	"log"
	"net/http"
)

func startHttpServer() {
	initError := queue.InitializeConnection()
	if initError != nil {
		log.Panic("Unable to initialize redis queue, error: %v", initError)
	}
	//start monitor service
	go func() {
		log.Println("in lambda function from s3 event")
		monitor.MontiorService()
	}()

	http.HandleFunc("/upload", handler.UploadFileHandler)
	http.HandleFunc("/retry", handler.RetryHandler)
	fmt.Println("Swilly Alert System is running on :3333")
	http.ListenAndServe(":3333", nil)
}

func startWorker() {

	initError := queue.InitializeConnection()
	if initError != nil {
		log.Panic("Unable to initialize redis queue, error: %v", initError)
	}

	// Put all unacked messages back to ready state
	queue.CleanQueues()

	//Start consumers
	consumerErrors := consumer.StartConsumers()
	if len(consumerErrors) > 0 {
		log.Printf("Unable to initialize redis queue workers, errors: %v\n", consumerErrors)
	}

	//wait for termination for handling it gracefully
	queue.WaitforQueueTerminationSignal()
}

func main() {
	mode := flag.String("mode", constants.SERVER_MODE_HTTP_SERVER, "")
	flag.Parse()
	switch *mode {
	case constants.SERVER_MODE_HTTP_SERVER:
		startHttpServer()
	case constants.SERVER_MODE_WORKER:
		startWorker()
	}
}
