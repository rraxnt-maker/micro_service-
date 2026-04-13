package main

import (
	"log"
	"net/http"
	"user/internal/config"
	"user/internal/handler"
	"user/internal/storage"
)

func main() {
	// Инициализируем БД
	if err := storage.Init(); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/profile", handler.GetProfile)
	mux.HandleFunc("/profile/update", handler.UpdateProfile)
	mux.HandleFunc("/user/", handler.GetUserByID)
	mux.HandleFunc("/profile/delete", handler.DeleteProfile)
	mux.HandleFunc("/sync", handler.SyncUser)
	mux.HandleFunc("/health", handler.HealthCheck)

	log.Printf("Server starting on %s", config.ServerPort)
	log.Printf("Available endpoints:")
	log.Printf("  GET    /profile         - Get own profile")
	log.Printf("  PUT    /profile/update  - Update own profile")
	log.Printf("  DELETE /profile/delete  - Delete own profile")
	log.Printf("  GET    /user/{id}       - Get public user profile")
	log.Printf("  POST   /sync            - Sync user from auth service")
	log.Printf("  GET    /health          - Health check")

	if err := http.ListenAndServe(config.ServerPort, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}