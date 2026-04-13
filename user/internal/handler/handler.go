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
	"user/internal/config"
	"user/internal/model"
	"user/internal/storage"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// Вспомогательные функции
func validateUserID(id string) bool {
	return uuidRegex.MatchString(id)
}

func validateEmail(email string) bool {
	return strings.Contains(email, "@") && len(email) > 5
}

func validateAge(age int) bool {
	return age >= 0 && age <= 150
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

// GET /profile - получить свой профиль
func GetProfile(w http.ResponseWriter, r *http.Request) {
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

	user, err := storage.DB.GetUser(userID)
	if err != nil {
		log.Printf("Database error: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		writeJSONError(w, "User not found", http.StatusNotFound)
		return
	}

	writeJSONResponse(w, http.StatusOK, user)
}

// PUT /profile/update - обновить свой профиль
func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
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

	var rawUpdates map[string]interface{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&rawUpdates)
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

	if len(rawUpdates) == 0 {
		writeJSONError(w, "No fields to update", http.StatusBadRequest)
		return
	}

	updates := make(map[string]interface{})

	if username, ok := rawUpdates["username"].(string); ok {
		username = strings.TrimSpace(username)
		if username == "" {
			writeJSONError(w, "Username cannot be empty", http.StatusBadRequest)
			return
		}
		if len(username) < 3 || len(username) > 50 {
			writeJSONError(w, "Username must be between 3 and 50 characters", http.StatusBadRequest)
			return
		}
		updates["username"] = username
	}

	if fullName, ok := rawUpdates["full_name"].(string); ok {
		fullName = strings.TrimSpace(fullName)
		if fullName == "" {
			writeJSONError(w, "Full name cannot be empty", http.StatusBadRequest)
			return
		}
		if len(fullName) > 100 {
			writeJSONError(w, "Full name must not exceed 100 characters", http.StatusBadRequest)
			return
		}
		updates["full_name"] = fullName
	}

	if age, ok := rawUpdates["age"].(float64); ok {
		intAge := int(age)
		if !validateAge(intAge) {
			writeJSONError(w, "Age must be between 0 and 150", http.StatusBadRequest)
			return
		}
		updates["age"] = intAge
	}

	if len(updates) == 0 {
		writeJSONError(w, "No valid fields to update", http.StatusBadRequest)
		return
	}

	user, err := storage.DB.UpdateUser(userID, updates)
	if err != nil {
		log.Printf("Database error: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		writeJSONError(w, "User not found", http.StatusNotFound)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status": "updated",
		"user":   user,
	})
}

// GET /user/{id} - получить профиль другого пользователя
func GetUserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 2 || pathParts[0] != "user" {
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

	user, err := storage.DB.GetPublicUser(userID)
	if err != nil {
		log.Printf("Database error: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		writeJSONError(w, "User not found", http.StatusNotFound)
		return
	}

	writeJSONResponse(w, http.StatusOK, user)
}

// DELETE /profile/delete - удалить свой аккаунт
func DeleteProfile(w http.ResponseWriter, r *http.Request) {
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

	deleted, err := storage.DB.DeleteUser(userID)
	if err != nil {
		log.Printf("Database error: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !deleted {
		writeJSONError(w, "User not found", http.StatusNotFound)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{
		"status":  "deleted",
		"user_id": userID,
	})
}

// POST /sync - синхронизация от Auth Service
func SyncUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(nil, r.Body, config.MaxBodySize)
	defer r.Body.Close()

	var data model.SyncRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&data)
	if err != nil {
		writeJSONError(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	if data.ID == "" {
		writeJSONError(w, "Missing id", http.StatusBadRequest)
		return
	}

	if !validateUserID(data.ID) {
		writeJSONError(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	if data.Email == "" {
		writeJSONError(w, "Missing email", http.StatusBadRequest)
		return
	}

	if !validateEmail(data.Email) {
		writeJSONError(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	status, err := storage.DB.CreateOrUpdateUser(data.ID, data.Email)
	if err != nil {
		log.Printf("Database error: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":  status,
		"user_id": data.ID,
	})
}

// GET /health - проверка работоспособности
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := storage.DB.Ping(); err != nil {
		writeJSONError(w, "Database connection failed", http.StatusServiceUnavailable)
		return
	}

	count, err := storage.DB.CountUsers()
	if err != nil {
		log.Printf("Database error: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"users":     count,
		"timestamp": time.Now(),
		"database":  "connected",
	})
}