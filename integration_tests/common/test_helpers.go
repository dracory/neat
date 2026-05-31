package common

import (
	"fmt"
	"os"
)

// GetEnv gets an environment variable or returns a default value.
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt gets an environment variable as an integer or returns a default value.
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var parsed int
		_, err := fmt.Sscanf(value, "%d", &parsed)
		if err == nil {
			return parsed
		}
	}
	return defaultValue
}
