package main

import (
	"log"
	"net/http"
	"user/internal/config"
	"user/internal/handler"
	"user/internal/storage"
)

func main() {
	if err := storage.Init(); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/profile", handler.GetProfile)
	mux.HandleFunc("/profile/update", handler.UpdateProfile)
	mux.HandleFunc("/profile/delete", handler.DeleteProfile)
	mux.HandleFunc("/user", handler.GetUserByID)
	mux.HandleFunc("/sync", handler.SyncUser)
	mux.HandleFunc("/health", handler.HealthCheck)

	log.Printf("Server starting on %s", config.ServerPort)

	if err := http.ListenAndServe(config.ServerPort, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}