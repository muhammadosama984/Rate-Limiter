package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)
type Message struct {
	Status string `json:"status"`
	Body string `json:"body"`
}

func perClientRateLimiter(next func(w http.ResponseWriter, r *http.Request))  http.Handler {
    type Client struct {
		limiter *rate.Limiter 
		lastSeen time.Time  
	}

	var (
		mu sync.Mutex
		clients = make(map[string]*Client)
	)

	go func() {
		for {
            time.Sleep(time.Minute)
            mu.Lock()
            for ip, client := range clients {
                if time.Since(client.lastSeen) > 3*time.Minute {
                    delete(clients, ip)
                }
            }
            mu.Unlock()
        }
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return 
		}

		mu.Lock()
		if _, found := clients[ip]; !found{
			clients[ip] = &Client{limiter: rate.NewLimiter(2,4)}
		}
		clients[ip].lastSeen = time.Now()
		if !clients[ip].limiter.Allow(){
				mu.Unlock()
				message := Message{
					Status: "error",
					Body: "API reached it capacity limit",
				}
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(&message)
				return
		} 
		next(w, r)
		mu.Unlock()

	})

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
    http.Handle("/ping", perClientRateLimiter(endpointHandler))
    err := http.ListenAndServe(":8080", nil)
	if err!= nil {
        log.Fatal("ListenAndServe: ", err)
    }
}