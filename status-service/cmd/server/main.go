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
	// Инициализируем БД
	if err := storage.Init(); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	// Запускаем очиститель истекших статусов
	cln := cleaner.NewCleaner(1 * time.Minute)
	cln.Start()
	defer cln.Stop()

	// Регистрируем маршруты
	mux := http.NewServeMux()
	
	// Публичные эндпоинты
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetStatus(w, r)
		case http.MethodPut:
			handler.SetStatus(w, r)
		case http.MethodDelete:
			handler.DeleteStatus(w, r)
		default:
			handler.GetStatus(w, r) // fallback
		}
	})
	mux.HandleFunc("/status/", handler.GetUserStatus)
	mux.HandleFunc("/status/batch", handler.GetBatchStatuses)
	mux.HandleFunc("/status/dnd", handler.SetDND)
	mux.HandleFunc("/status/history", handler.GetHistory)
	
	// Внутренние эндпоинты
	mux.HandleFunc("/internal/sync", handler.InternalSync)
	mux.HandleFunc("/internal/user/", handler.InternalDeleteUser)
	
	// Health check
	mux.HandleFunc("/health", handler.HealthCheck)

	// Запускаем сервер
	server := &http.Server{
		Addr:         config.ServerPort,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🚀 Status Service starting on %s", config.ServerPort)
	log.Printf("📋 Available endpoints:")
	log.Printf("   GET    /status           - Get own status")
	log.Printf("   PUT    /status           - Set status")
	log.Printf("   DELETE /status           - Delete status")
	log.Printf("   GET    /status/{id}      - Get user status")
	log.Printf("   POST   /status/batch     - Get multiple statuses")
	log.Printf("   POST   /status/dnd       - Set DND mode")
	log.Printf("   GET    /status/history   - Get status history")
	log.Printf("   GET    /health           - Health check")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}