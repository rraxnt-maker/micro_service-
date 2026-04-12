package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// In-memory "база данных" (потом заменишь на реальную БД)
var (
	users = make(map[string]*User) // key = user_id
	mu    sync.RWMutex
)

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	FullName  string `json:"full_name"`
	Age       int    `json:"age"`
	CreatedAt string `json:"created_at"`
}

// GET /profile - получить свой профиль
func GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	mu.RLock()
	user, exists := users[userID]
	mu.RUnlock()

	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// PUT /profile/update - обновить свой профиль
func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var updates map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	mu.Lock()
	user, exists := users[userID]
	if !exists {
		mu.Unlock()
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if name, ok := updates["username"].(string); ok {
		user.Username = name
	}
	if name, ok := updates["full_name"].(string); ok {
		user.FullName = name
	}
	if age, ok := updates["age"].(float64); ok {
		user.Age = int(age)
	}
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// GET /user/{id} - получить профиль другого пользователя
func GetUserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := strings.TrimPrefix(r.URL.Path, "/user/")
	if userID == "" {
		http.Error(w, "Missing user id", http.StatusBadRequest)
		return
	}

	mu.RLock()
	user, exists := users[userID]
	mu.RUnlock()

	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Возвращаем только публичные данные (без email)
	publicData := map[string]interface{}{
		"id":        user.ID,
		"username":  user.Username,
		"full_name": user.FullName,
		"age":       user.Age,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(publicData)
}

// DELETE /profile/delete - удалить свой аккаунт
func DeleteProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	mu.Lock()
	delete(users, userID)
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// POST /sync - синхронизация от Auth Service (при регистрации)
func SyncUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if data.ID == "" || data.Email == "" {
		http.Error(w, "Missing id or email", http.StatusBadRequest)
		return
	}

	mu.Lock()
	_, exists := users[data.ID]
	if !exists {
		// Создаём нового пользователя с пустым профилем
		users[data.ID] = &User{
			ID:        data.ID,
			Email:     data.Email,
			Username:  "",
			FullName:  "",
			Age:       0,
			CreatedAt: "now",
		}
	}
	mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "synced"})
}
