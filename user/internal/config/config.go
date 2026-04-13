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
	
	StatusServiceURL = getEnv("STATUS_SERVICE_URL", "http://localhost:8070")
	InternalToken    = getEnv("INTERNAL_TOKEN", "secret-internal-token-123")
)

var (
	AllowedMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	AllowedHeaders = []string{"Content-Type"}
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}