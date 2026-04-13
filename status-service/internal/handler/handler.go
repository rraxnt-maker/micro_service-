package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
	"fmt"
	"status-service/internal/config"
	"status-service/internal/model"
	"status-service/internal/storage"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// Вспомогательные функции
func validateUserID(id string) bool {
	return uuidRegex.MatchString(id)
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// PUT /status - установить статус
func SetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !validateUserID(userID) {
		writeJSONError(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(nil, r.Body, config.MaxBodySize)
	defer r.Body.Close()

	var req model.SetStatusRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&req)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			writeJSONError(w, "Bad JSON syntax", http.StatusBadRequest)
		case errors.As(err, &unmarshalTypeError):
			writeJSONError(w, "Bad JSON field type", http.StatusBadRequest)
		case errors.Is(err, io.EOF):
			writeJSONError(w, "Empty request body", http.StatusBadRequest)
		case strings.Contains(err.Error(), "unknown field"):
			writeJSONError(w, "Request contains unknown field", http.StatusBadRequest)
		default:
			writeJSONError(w, "Bad JSON", http.StatusBadRequest)
		}
		return
	}

	// Валидация
	req.Text = strings.TrimSpace(req.Text)
	if req.Text == "" {
		writeJSONError(w, "Text cannot be empty", http.StatusBadRequest)
		return
	}
	if len(req.Text) > 140 {
		writeJSONError(w, "Text must not exceed 140 characters", http.StatusBadRequest)
		return
	}

	if req.Emoji != "" && len(req.Emoji) > 10 {
		writeJSONError(w, "Emoji must not exceed 10 characters", http.StatusBadRequest)
		return
	}

	validTypes := map[string]bool{"normal": true, "dnd": true, "custom": true}
	if req.Type != "" && !validTypes[req.Type] {
		writeJSONError(w, "Invalid type. Must be: normal, dnd, custom", http.StatusBadRequest)
		return
	}

	if req.ExpiresIn < 0 {
		writeJSONError(w, "Expires in cannot be negative", http.StatusBadRequest)
		return
	}

	status, err := storage.DB.SetStatus(userID, &req)
	if err != nil {
		log.Printf("Database error: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status": "set",
		"data":   status,
	})
}

// GET /status - получить свой статус
func GetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !validateUserID(userID) {
		writeJSONError(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	status, err := storage.DB.GetStatus(userID)
	if err != nil {
		log.Printf("Database error: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if status == nil {
		writeJSONError(w, "Status not found", http.StatusNotFound)
		return
	}

	writeJSONResponse(w, http.StatusOK, status)
}

// GET /status/{user_id} - получить статус другого пользователя
func GetUserStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 2 || pathParts[0] != "status" {
		writeJSONError(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	userID := pathParts[1]
	if userID == "" {
		writeJSONError(w, "Missing user id", http.StatusBadRequest)
		return
	}

	if !validateUserID(userID) {
		writeJSONError(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	status, err := storage.DB.GetPublicStatus(userID)
	if err != nil {
		log.Printf("Database error: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if status == nil {
		writeJSONError(w, "Status not found", http.StatusNotFound)
		return
	}

	writeJSONResponse(w, http.StatusOK, status)
}

// POST /status/batch - получить статусы нескольких пользователей
func GetBatchStatuses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(nil, r.Body, config.MaxBodySize)
	defer r.Body.Close()

	var req model.BatchRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&req)
	if err != nil {
		writeJSONError(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	if len(req.UserIDs) == 0 {
		writeJSONError(w, "User IDs cannot be empty", http.StatusBadRequest)
		return
	}

	if len(req.UserIDs) > 100 {
		writeJSONError(w, "Maximum 100 user IDs allowed", http.StatusBadRequest)
		return
	}

	for _, id := range req.UserIDs {
		if !validateUserID(id) {
			writeJSONError(w, "Invalid user ID format: "+id, http.StatusBadRequest)
			return
		}
	}

	statuses, err := storage.DB.GetBatchStatuses(req.UserIDs)
	if err != nil {
		log.Printf("Database error: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"statuses": statuses,
	})
}

// DELETE /status - удалить свой статус
func DeleteStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !validateUserID(userID) {
		writeJSONError(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	err := storage.DB.DeleteStatus(userID)
	if err != nil {
		log.Printf("Database error: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{
		"status": "cleared",
	})
}

// POST /status/dnd - установить режим "Не беспокоить"
func SetDND(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !validateUserID(userID) {
		writeJSONError(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(nil, r.Body, config.MaxBodySize)
	defer r.Body.Close()

	var req model.DNDRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&req)
	if err != nil {
		// Пустое тело - ок, используем значения по умолчанию
		if !errors.Is(err, io.EOF) {
			writeJSONError(w, "Bad JSON", http.StatusBadRequest)
			return
		}
	}

	if req.Duration < 0 {
		writeJSONError(w, "Duration cannot be negative", http.StatusBadRequest)
		return
	}

	status, err := storage.DB.SetDND(userID, req.Duration)
	if err != nil {
		log.Printf("Database error: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status": "dnd_enabled",
		"data":   status,
	}
	if status.ExpiresAt != nil {
		response["until"] = status.ExpiresAt
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// GET /status/history - получить историю статусов
func GetHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !validateUserID(userID) {
		writeJSONError(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	// Парсим query параметры
	limit := 10
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if _, err := fmt.Sscan(l, &limit); err != nil || limit < 1 || limit > 100 {
			writeJSONError(w, "Limit must be between 1 and 100", http.StatusBadRequest)
			return
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if _, err := fmt.Sscan(o, &offset); err != nil || offset < 0 {
			writeJSONError(w, "Offset must be non-negative", http.StatusBadRequest)
			return
		}
	}

	history, total, err := storage.DB.GetHistory(userID, limit, offset)
	if err != nil {
		log.Printf("Database error: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"history": history,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// POST /internal/sync - синхронизация пользователя (внутренний эндпоинт)
func InternalSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем внутренний токен
	token := r.Header.Get("X-Internal-Token")
	if token != config.InternalToken {
		writeJSONError(w, "Forbidden", http.StatusForbidden)
		return
	}

	r.Body = http.MaxBytesReader(nil, r.Body, config.MaxBodySize)
	defer r.Body.Close()

	var req model.SyncRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSONError(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		writeJSONError(w, "Missing user_id", http.StatusBadRequest)
		return
	}

	if !validateUserID(req.UserID) {
		writeJSONError(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	err = storage.DB.SyncUser(req.UserID)
	if err != nil {
		log.Printf("Database error: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{
		"status": "synced",
	})
}

// DELETE /internal/user/{user_id} - удаление пользователя (внутренний эндпоинт)
func InternalDeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем внутренний токен
	token := r.Header.Get("X-Internal-Token")
	if token != config.InternalToken {
		writeJSONError(w, "Forbidden", http.StatusForbidden)
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 3 || pathParts[0] != "internal" || pathParts[1] != "user" {
		writeJSONError(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	userID := pathParts[2]
	if userID == "" {
		writeJSONError(w, "Missing user id", http.StatusBadRequest)
		return
	}

	if !validateUserID(userID) {
		writeJSONError(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	err := storage.DB.DeleteUser(userID)
	if err != nil {
		log.Printf("Database error: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{
		"status": "deleted",
	})
}

// GET /health - проверка работоспособности
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := storage.DB.Ping(); err != nil {
		writeJSONError(w, "Database connection failed", http.StatusServiceUnavailable)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"database":  "connected",
	})
}