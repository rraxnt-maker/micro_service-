package config

import "os"

const (
	ServerPort  = ":8070"
	MaxBodySize = 1 << 20 // 1MB
)

var (
	DBHost     = getEnv("DB_HOST", "localhost")
	DBPort     = getEnv("DB_PORT", "5431")
	DBUser     = getEnv("DB_USER", "statusadmin")
	DBPassword = getEnv("DB_PASSWORD", "statuspass")
	DBName     = getEnv("DB_NAME", "statusdb")
	
	InternalToken = getEnv("INTERNAL_TOKEN", "secret-internal-token-123")
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}