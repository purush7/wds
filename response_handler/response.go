package responsehandler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type GenericResponse struct {
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
}

func CustomError(w http.ResponseWriter, httpStatusCode int, msg string) {
	errJSON := GenericResponse{
		Data:  nil,
		Error: msg,
	}
	RespondwithJSON(w, httpStatusCode, errJSON)
}

func RespondwithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		response = []byte(fmt.Sprintf("error while marshalling the payload: %v", err))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
