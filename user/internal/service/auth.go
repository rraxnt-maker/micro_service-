// service/auth.go
package service

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type AuthService struct {
	authURL string
}

func NewAuthService(authURL string) *AuthService {
	return &AuthService{authURL: authURL}
}

func (a *AuthService) ValidateToken(token string) (string, error) {
	// Запрос к Auth Service
	reqBody, _ := json.Marshal(map[string]string{"token": token})
	resp, err := http.Post(a.authURL+"/validate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		UserID string `json:"user_id"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.UserID, nil
}
