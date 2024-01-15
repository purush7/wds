package handler

import (
	"alert_system/alert_initiator/internal/queue"
	"alert_system/constants"
	responsehandler "alert_system/response_handler"
	"alert_system/util"
	"fmt"
	"math"
	"net/http"
)

func UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form data
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit for the entire form
	if err != nil {
		err = fmt.Errorf("Unable to parse form:%v", err)
		responsehandler.CustomError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Retrieve the file from the form data
	file, _, err := r.FormFile("file")
	if err != nil {
		err = fmt.Errorf("Error retrieving file from form data:%v", err)
		responsehandler.CustomError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer file.Close()

	//Save file
	absPath, err := util.SaveFile(file)
	if err != nil {
		err = fmt.Errorf("Error while saving the file:%v", err)
		responsehandler.CustomError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var response responsehandler.GenericResponse
	response.Data = absPath
	responsehandler.RespondwithJSON(w, http.StatusOK, response)
}

func RetryHandler(w http.ResponseWriter, r *http.Request) {

	batcherQueue, err := queue.GetOpenQueue(constants.BatcherQueue)
	if err != nil {
		err = fmt.Errorf("Error while connecting to batcher queue:%v", err)
		responsehandler.CustomError(w, http.StatusInternalServerError, err.Error())
		return
	}

	batcherRejectedCount, err := batcherQueue.Queue.ReturnRejected(math.MaxInt64)
	if err != nil {
		err = fmt.Errorf("Error while retrying the rejected tasks of batcher:%v", err)
		responsehandler.CustomError(w, http.StatusInternalServerError, err.Error())
		return
	}

	notifierQueue, err := queue.GetOpenQueue(constants.NotifierQueue)
	if err != nil {
		err = fmt.Errorf("Error while connecting to notifier queue:%v", err)
		responsehandler.CustomError(w, http.StatusInternalServerError, err.Error())
		return
	}

	notifierRejectedCount, err := notifierQueue.Queue.ReturnRejected(math.MaxInt64)
	if err != nil {
		err = fmt.Errorf("Error while retrying the rejected tasks of notifier:%v", err)
		responsehandler.CustomError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var rejectedCount = map[string]int64{
		constants.BatcherQueue:  batcherRejectedCount,
		constants.NotifierQueue: notifierRejectedCount,
	}

	var response responsehandler.GenericResponse
	response.Data = rejectedCount
	responsehandler.RespondwithJSON(w, http.StatusOK, response)
}
