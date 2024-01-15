package main

import (
	"alert_system/constants"
	responsehandler "alert_system/response_handler"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth/v7/limiter"
)

type webhookRequest struct {
	UserID       string `json:"userId"`
	AlertMessage string `json:"alertMessage"`
}

var rateLimiter = tollbooth.NewLimiter(constants.RatelimiterRPS, &limiter.ExpirableOptions{DefaultExpirationTTL: 10 * time.Minute})
var timeString string = time.Now().Format("20060102150405.003059_")

func webhookHandler(w http.ResponseWriter, r *http.Request) {

	limitError := tollbooth.LimitByKeys(rateLimiter, []string{strings.Split(r.RemoteAddr, ":")[0], r.URL.Path})
	if limitError != nil {
		responsehandler.CustomError(w, http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests))
		return
	}

	var req webhookRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		responsehandler.CustomError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	fmt.Printf("Received webhook - UserID: %s, AlertMessage: %s\n", req.UserID, req.AlertMessage)

	fileName := timeString + ".csv"
	absPath := filepath.Join("/tmp/output", fileName)

	dirPath := filepath.Dir(absPath)
	// Create directories recursively
	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		err = fmt.Errorf("Error creating directories:%v", err)
		responsehandler.CustomError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Open the CSV file for appending
	file, err := os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		err = fmt.Errorf("Error opening file:%v", err)
		responsehandler.CustomError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Data to append
	newData := []string{req.UserID, req.AlertMessage}

	// Append data to the CSV file
	err = writer.Write(newData)
	if err != nil {
		err = fmt.Errorf("Error writing to file:%v", err)
		responsehandler.CustomError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var response responsehandler.GenericResponse
	response.Data = "success"
	responsehandler.RespondwithJSON(w, http.StatusOK, response)
}

func main() {
	http.HandleFunc("/webhook", webhookHandler)
	fmt.Println("Swilly Webhook Server is running on :3333")
	http.ListenAndServe(":3333", nil)
}
