//go:build integration

package postgres

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dracory/neat"
	"github.com/dracory/neat/database"
)

// GetPostgresConfig returns a PostgreSQL connection config from environment variables
func GetPostgresConfig() neat.DBConfig {
	host := getEnv("POSTGRES_HOST", "127.0.0.1")
	port := getEnvInt("POSTGRES_PORT", 5432)
	database := getEnv("POSTGRES_DATABASE", "test")
	username := getEnv("POSTGRES_USER", "test")
	password := getEnv("POSTGRES_PASS", "test")
	sslmode := getEnv("POSTGRES_SSLMODE", "disable")

	return neat.DBConfig{
		Default: "postgres",
		Connections: map[string]neat.ConnectionConfig{
			"postgres": {
				Driver:   "postgres",
				Host:     host,
				Port:     port,
				Database: database,
				Username: username,
				Password: password,
				SSLMode:  sslmode,
				Schema:   "public",
			},
		},
		Pool: neat.PoolConfig{
			MaxIdleConns:    5,
			MaxOpenConns:    10,
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: time.Hour,
		},
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an environment variable as an integer or returns a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var parsed int
		_, err := fmt.Sscanf(value, "%d", &parsed)
		if err == nil {
			return parsed
		}
	}
	return defaultValue
}

// SetupPostgresTest creates a database connection and registers cleanup.
func SetupPostgresTest(t *testing.T) *database.Database {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	host := getEnv("POSTGRES_HOST", "127.0.0.1")
	port := getEnvInt("POSTGRES_PORT", 55432)
	db := getEnv("POSTGRES_DATABASE", "test")
	username := getEnv("POSTGRES_USER", "test")
	password := getEnv("POSTGRES_PASS", "test")
	sslmode := getEnv("POSTGRES_SSLMODE", "disable")
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		username, password, host, port, db, sslmode)

	conn, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	t.Cleanup(func() {
		conn.Close()
	})

	return conn
}

// SetupPostgresConnection creates a database connection without setting up tables
func SetupPostgresConnection(t *testing.T) *database.Database {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	host := getEnv("POSTGRES_HOST", "127.0.0.1")
	port := getEnvInt("POSTGRES_PORT", 55432)
	database := getEnv("POSTGRES_DATABASE", "test")
	username := getEnv("POSTGRES_USER", "test")
	password := getEnv("POSTGRES_PASS", "test")
	sslmode := getEnv("POSTGRES_SSLMODE", "disable")
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		username, password, host, port, database, sslmode)

	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}
