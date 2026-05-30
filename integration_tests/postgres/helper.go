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

	createPostgresTestTables(t, conn)
	cleanupPostgresTestData(t, conn)

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

// cleanupPostgresTestData removes all data from test tables
func cleanupPostgresTestData(t *testing.T, db *database.Database) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("cleanupPostgresTestData: DB(): %v", err)
	}
	stmts := []string{
		`DELETE FROM users`,
		`DELETE FROM addresses`,
		`DELETE FROM books`,
		`DELETE FROM peoples`,
		`DELETE FROM json_datas`,
	}
	for _, stmt := range stmts {
		if _, err := sqlDB.Exec(stmt); err != nil {
			// Ignore errors if tables don't exist yet
			continue
		}
	}
}

// createPostgresTestTables creates all tables required by the integration test models.
func createPostgresTestTables(t *testing.T, db *database.Database) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("createPostgresTestTables: DB(): %v", err)
	}
	stmts := []string{
		`DROP TABLE IF EXISTS books CASCADE`,
		`DROP TABLE IF EXISTS addresses CASCADE`,
		`DROP TABLE IF EXISTS users CASCADE`,
		`DROP TABLE IF EXISTS peoples CASCADE`,
		`DROP TABLE IF EXISTS json_datas CASCADE`,
		`CREATE TABLE IF NOT EXISTS users (
			id         BIGSERIAL PRIMARY KEY,
			name       VARCHAR(255) NOT NULL DEFAULT '',
			avatar     VARCHAR(255) NOT NULL DEFAULT '',
			bio        TEXT,
			votes      INTEGER NOT NULL DEFAULT 0,
			deleted_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS addresses (
			id         BIGSERIAL PRIMARY KEY,
			name       VARCHAR(255) NOT NULL DEFAULT '',
			user_id    BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS books (
			id         BIGSERIAL PRIMARY KEY,
			name       VARCHAR(255) NOT NULL DEFAULT '',
			user_id    BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS peoples (
			id         BIGSERIAL PRIMARY KEY,
			body       TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS json_datas (
			id         BIGSERIAL PRIMARY KEY,
			data       JSONB NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, stmt := range stmts {
		if _, err := sqlDB.Exec(stmt); err != nil {
			t.Fatalf("createPostgresTestTables: %v", err)
		}
	}
}
