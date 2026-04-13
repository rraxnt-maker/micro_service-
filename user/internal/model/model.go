package model

import "time"

// User - основная модель пользователя
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	FullName  string    `json:"full_name"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PublicUser - модель для публичных данных
type PublicUser struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	FullName  string    `json:"full_name"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
}

// SyncRequest - запрос на синхронизацию
type SyncRequest struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}