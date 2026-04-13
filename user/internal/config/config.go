package config

import "os"

const (
	ServerPort  = ":8069"
	MaxBodySize = 1 << 20
)

var (
	DBHost     = getEnv("DB_HOST", "localhost")
	DBPort     = getEnv("DB_PORT", "5430")
	DBUser     = getEnv("DB_USER", "useradmin")
	DBPassword = getEnv("DB_PASSWORD", "secretpassword")
	DBName     = getEnv("DB_NAME", "userdb")
	
	// Новые переменные для Status Service
	StatusServiceURL   = getEnv("STATUS_SERVICE_URL", "http://localhost:8070")
	InternalToken      = getEnv("INTERNAL_TOKEN", "secret-internal-token-123")
)

var (
	AllowedMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	AllowedHeaders = []string{"Content-Type", "X-User-ID"}
)

const (
	// ServerPort - порт сервера
	ServerPort = ":8069"
	
	// MaxBodySize - максимальный размер тела запроса (1MB)
	MaxBodySize = 1 << 20
)

// Database config с поддержкой переменных окружения
var (
	DBHost     = getEnv("DB_HOST", "localhost")
	DBPort     = getEnv("DB_PORT", "5430")  // По умолчанию 5430
	DBUser     = getEnv("DB_USER", "useradmin")
	DBPassword = getEnv("DB_PASSWORD", "secretpassword")
	DBName     = getEnv("DB_NAME", "userdb")
)

var (
	// AllowedMethods - разрешенные HTTP методы
	AllowedMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	
	// AllowedHeaders - разрешенные заголовки
	AllowedHeaders = []string{"Content-Type", "X-User-ID"}
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}