//go:build integration

package postgres

import (
	"fmt"
	"testing"
	"time"

	"github.com/dracory/neat"
	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/common"
)

// GetPostgresConfig returns a PostgreSQL connection config from environment variables
func GetPostgresConfig() neat.DBConfig {
	host := common.GetEnv("POSTGRES_HOST", "127.0.0.1")
	port := common.GetEnvInt("POSTGRES_PORT", 55432)
	database := common.GetEnv("POSTGRES_DATABASE", "test")
	username := common.GetEnv("POSTGRES_USER", "test")
	password := common.GetEnv("POSTGRES_PASS", "test")
	sslmode := common.GetEnv("POSTGRES_SSLMODE", "disable")

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

// SetupPostgresTest creates a database connection and registers cleanup.
func SetupPostgresTest(t *testing.T) *database.Database {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	host := common.GetEnv("POSTGRES_HOST", "127.0.0.1")
	port := common.GetEnvInt("POSTGRES_PORT", 55432)
	db := common.GetEnv("POSTGRES_DATABASE", "test")
	username := common.GetEnv("POSTGRES_USER", "test")
	password := common.GetEnv("POSTGRES_PASS", "test")
	sslmode := common.GetEnv("POSTGRES_SSLMODE", "disable")
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

	host := common.GetEnv("POSTGRES_HOST", "127.0.0.1")
	port := common.GetEnvInt("POSTGRES_PORT", 55432)
	database := common.GetEnv("POSTGRES_DATABASE", "test")
	username := common.GetEnv("POSTGRES_USER", "test")
	password := common.GetEnv("POSTGRES_PASS", "test")
	sslmode := common.GetEnv("POSTGRES_SSLMODE", "disable")
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
