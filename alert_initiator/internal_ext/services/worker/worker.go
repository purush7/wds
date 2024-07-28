package worker

import (
	"alert_system/alert_initiator/internal_ext/client"
	"alert_system/alert_initiator/internal_ext/entities"
	"alert_system/alert_initiator/internal_ext/model"
	"alert_system/alert_initiator/internal_ext/services/producer"
	"alert_system/constants"
	"alert_system/infra/redis"
	"alert_system/util"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var notifierClient = client.GetNotfierClient()
var errorBadRequest = fmt.Errorf("error: got bad request")
var errorTooManyRequests = fmt.Errorf("error: retrying,got too may requests")
var errorInternalServer = fmt.Errorf("error: got 5xx")

func ProcessBatcherTopic(payload []byte) error {
	var data entities.BatcherPayload
	err := json.Unmarshal(payload, &data)
	if err != nil {
		return err
	}

	//Check for 429
	tooManyRequests, err := redis.Get(constants.TooManyRequests)
	if err != nil && err != redis.Nil {
		return err
	}

	if tooManyRequests == constants.TooManyRequestsTrue {

		for {
			status, err := redis.Get(constants.Publishing)
			if err != nil && err != redis.Nil {
				return err
			}
			if status == constants.PublishingTrue || err == redis.Nil {
				time.Sleep(time.Second)
				continue
			}
			break
		}

		//Write into db and ack
		model.WriteIntoDb(data.Id, data.FilePath)
		// time.Sleep(time.Second)
		return nil
	}

	// Open file
	file, err := os.Open(data.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := csv.NewReader(file)

	offset := int64(0)
	offsetString, err := redis.Get(data.Id)
	if err != redis.Nil {
		offset, err = strconv.ParseInt(offsetString, 10, 64)
		if err != nil {
			return err
		}
	}
	// Seek to the desired offset
	for i := int64(0); i < offset; i++ {
		_, err := reader.Read()
		if err == io.EOF {
			redis.Delete(data.Id)
			return nil
		} else if err != nil {
			fmt.Println("Error reading file:", err)
			return err
		}
	}

	for {
		// Read one record (row) from the CSV file
		record, err := reader.Read()
		// If we reach the end of the file, break the loop
		if err == io.EOF {
			break
		}

		// Handle other errors
		if err != nil {
			log.Println("error while reading file: ", err)
			return err
		}

		//Skip user notification if row data format is wrong
		if len(record) < 2 || record[0] == "" || record[1] == "" {
			continue
		}

		err = producer.PublishIntoQueues(constants.NotifierQueue, entities.NotifierPayload{
			UserId:       record[0],
			AlertMessage: record[1],
		})
		if err != nil {
			return err
		}
		offset++
		err = redis.Incr(data.Id)
		if err != nil {
			log.Println("error while incrementing the users count of a file %s : ", data.FilePath, err)
			return err
		}
		// test for data integrity as we shouldn't miss any user nor alerts should be sent to the user multiple times
		// if offset  == 2{
		// 	return fmt.Errorf("exiting")
		// }
	}

	redis.Delete(data.Id)
	return nil
}

func ProcessNotifierTopic(payload []byte) error {

	type response struct {
		Data  string `json:"data"`
		Error string `json:"error"`
	}

	var webhookResponse response
	var statusCode int

	//11 retries and 1s+jitter sleep in between  ~ 1day
	err := util.CustomRetry(1, 1*time.Second, func() error {

		ctx, cancelFn := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancelFn()
		req, err := http.NewRequestWithContext(ctx, "POST", constants.NOTIFIER_URL, bytes.NewBuffer(payload))
		if err != nil {
			return util.NewStop(err.Error())
		}
		res, err := notifierClient.Do(req)
		if err != nil {
			return util.NewStop(err.Error())
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return util.NewStop(err.Error())
		}
		defer res.Body.Close()

		err = json.Unmarshal(body, &webhookResponse)
		if err != nil {
			return util.NewStop(err.Error())
		}

		statusCode = res.StatusCode

		//429
		if statusCode == http.StatusTooManyRequests {
			//trigger alert,
			err = errorTooManyRequests
			log.Println(err.Error())
			redis.Set(constants.TooManyRequests, constants.TooManyRequestsTrue, time.Second)
			// have a sleep to respect the rate-limiting (slow-down)
			time.Sleep(time.Second)
			return err
		}

		//4xx
		if statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError {
			err = errorBadRequest
			return util.NewStop(err.Error())
		}

		//5xx
		if statusCode >= http.StatusInternalServerError {
			err = errorInternalServer
			return err //retries
		}

		return nil
	})

	if err == errorBadRequest {
		return fmt.Errorf(webhookResponse.Error)
	}

	//5xx,429,payload error
	//return error will unack test it
	if err != nil {
		return err
	}

	return nil
}
