package model

import "time"

// Status - текущий статус пользователя
type Status struct {
	UserID    string     `json:"user_id"`
	Text      string     `json:"text"`
	Emoji     string     `json:"emoji"`
	Type      string     `json:"type"`
	Activity  string     `json:"activity,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// StatusHistory - запись в истории
type StatusHistory struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Text      string    `json:"text"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}

// SetStatusRequest - запрос на установку статуса
type SetStatusRequest struct {
	Text      string `json:"text"`
	Emoji     string `json:"emoji,omitempty"`
	Type      string `json:"type,omitempty"`
	Activity  string `json:"activity,omitempty"`
	ExpiresIn int    `json:"expires_in,omitempty"` // в секундах
}

// DNDRequest - запрос на установку DND
type DNDRequest struct {
	Duration int `json:"duration,omitempty"` // в секундах
}

// BatchRequest - запрос на получение нескольких статусов
type BatchRequest struct {
	UserIDs []string `json:"user_ids"`
}

// SyncRequest - внутренний запрос на синхронизацию
type SyncRequest struct {
	UserID string `json:"user_id"`
}

// PublicStatus - публичный статус (без внутренних полей)
type PublicStatus struct {
	UserID    string     `json:"user_id"`
	Text      string     `json:"text"`
	Emoji     string     `json:"emoji"`
	Type      string     `json:"type"`
	Activity  string     `json:"activity,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}