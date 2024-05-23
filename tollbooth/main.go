package main

import (
	"encoding/json"
	"log"
	"net/http"

	tollbooth "github.com/didip/tollbooth/v7"
)

type Message struct {
	Status string `json:"status"`
	Body string `json:"body"`
}

func endpointHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set(`Content-Type`, `application/json`)
	writer.WriteHeader(http.StatusOK)
	message := Message {
		Status: "OK",
        Body: "Hi You have reached the API endpoint",
	}
	err := json.NewEncoder(writer).Encode(&message)
	if err != nil {
		return 
	}
}

func main() {

	message := Message{
		Status: "error",
		Body: "API reached it capacity limit",
	}

	jsonMessage, _ := json.Marshal(message)
    tollbthLimiter := tollbooth.NewLimiter(1,nil)
	tollbthLimiter.SetMessageContentType("application/json")
	tollbthLimiter.SetMessage(string(jsonMessage))
	http.Handle("/ping", tollbooth.LimitFuncHandler(tollbthLimiter, endpointHandler))
	err := http.ListenAndServe(":8080", nil)
	if err!= nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
