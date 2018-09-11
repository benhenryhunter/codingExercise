package main

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

var (
	password = "angryMonkey"
	stats    = Stats{
		Total:       0,
		AverageTime: 0,
	}
)

type Stats struct {
	Total       int `json:"total"`
	AverageTime int `json:"average"`
}

func main() {
	// running server on port 8080
	srv := &http.Server{Addr: ":8080"}

	// handling requests on route /hash
	http.HandleFunc("/hash", func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		// We're expecting application/x-www-form-urlencoded so we parse the form
		// Then we get the password field and if it doesn't exist throw 400 error
		r.ParseForm()
		password := r.Form["password"]
		if len(password) != 1 {
			w.WriteHeader(400)
			w.Write([]byte("Improperly formed post body"))
		}

		// Generate new sha512 hash and write the value to it
		hash := sha512.New()
		hash.Write([]byte(password[0]))

		//return base64 encoded hash
		w.Write([]byte(base64.StdEncoding.EncodeToString(hash.Sum(nil))))
		duration := time.Now().Sub(startTime)
		stats.AverageTime = ((stats.AverageTime * stats.Total) + int(duration.Nanoseconds()*1000)) / (stats.Total + 1)
		stats.Total++
	})

	http.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		// Notify the user
		w.Write([]byte("Shutting down... Good Bye"))

		// Running a go routine because I wasn't getting the write through without it
		go func() {
			// http's shutdown method handles the graceful shutdown
			if err := srv.Shutdown(context.Background()); err != nil {
				log.Printf("HTTP server Shutdown: %v", err)
			}
		}()
	})

	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		// Set content type header
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(&stats); err != nil {
			// Some err happened with the encoder
			w.WriteHeader(500)
			w.Write([]byte("Error encoding struct: " + err.Error()))
		}
	})

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("HTTP server ListenAndServe: %v", err)
	}
}
