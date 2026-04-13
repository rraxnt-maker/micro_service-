package main

import (
	"log"
	"net/http"
	"time"
	"status-service/internal/cleaner"
	"status-service/internal/config"
	"status-service/internal/handler"
	"status-service/internal/storage"
)

func main() {
	if err := storage.Init(); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	cln := cleaner.NewCleaner(1 * time.Minute)
	cln.Start()
	defer cln.Stop()

	mux := http.NewServeMux()

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetStatus(w, r)
		case http.MethodPut:
			handler.SetStatus(w, r)
		case http.MethodDelete:
			handler.DeleteStatus(w, r)
		}
	})
	mux.HandleFunc("/status/batch", handler.GetBatchStatuses)
	mux.HandleFunc("/status/dnd", handler.SetDND)
	mux.HandleFunc("/status/history", handler.GetHistory)
	mux.HandleFunc("/internal/sync", handler.InternalSync)
	mux.HandleFunc("/internal/user", handler.InternalDeleteUser)
	mux.HandleFunc("/health", handler.HealthCheck)

	log.Printf("Status Service starting on %s", config.ServerPort)

	if err := http.ListenAndServe(config.ServerPort, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}