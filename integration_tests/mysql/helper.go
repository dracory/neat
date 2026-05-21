//go:build integration

package mysql

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database"
)

// TestModel is a simple model for integration testing
type TestModel struct {
	ID        uint      `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Age       int       `db:"age"`
	Active    bool      `db:"active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (TestModel) TableName() string {
	return "test_models"
}

// GetMySQLConfig returns a MySQL connection config from environment variables
func GetMySQLConfig() neat.DBConfig {
	host := getEnv("MYSQL_HOST", "127.0.0.1")
	port := getEnvInt("MYSQL_PORT", 3306)
	database := getEnv("MYSQL_DATABASE", "test")
	username := getEnv("MYSQL_USER", "root")
	password := getEnv("MYSQL_PASS", "root")

	return neat.DBConfig{
		Default: "mysql",
		Connections: map[string]neat.ConnectionConfig{
			"mysql": {
				Driver:   "mysql",
				Host:     host,
				Port:     port,
				Database: database,
				Username: username,
				Password: password,
				Charset:  "utf8mb4",
				Loc:      "Local",
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

// SetupTestDB creates a database connection and sets up test tables
func SetupTestDB(config neat.DBConfig) (*database.Database, error) {
	db, err := neat.New(config, database.WithLogger(log.NewNoopLogger()))
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Note: Schema builder setup is skipped for now - requires proper blueprint configuration
	// Tests should handle their own table setup as needed

	return db, nil
}

// TeardownTestDB drops test tables and closes the connection
func TeardownTestDB(db *database.Database) error {
	// Drop test table
	if db != nil {
		if err := db.Schema().Drop("test_models"); err != nil {
			return err
		}
		return db.Close()
	}
	return nil
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

// SetupMySQLTest creates a database connection and sets up test tables for MySQL
func SetupMySQLTest(t *testing.T) *database.Database {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	host := getEnv("MYSQL_HOST", "127.0.0.1")
	port := getEnvInt("MYSQL_PORT", 3306)
	database := getEnv("MYSQL_DATABASE", "test")
	username := getEnv("MYSQL_USER", "root")
	password := getEnv("MYSQL_PASS", "root")
	dsn := fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true",
		username, password, host, port, database)

	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to MySQL: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// SetupMySQLConnection creates a database connection without setting up tables
func SetupMySQLConnection(t *testing.T) *database.Database {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	host := getEnv("MYSQL_HOST", "127.0.0.1")
	port := getEnvInt("MYSQL_PORT", 3306)
	database := getEnv("MYSQL_DATABASE", "test")
	username := getEnv("MYSQL_USER", "root")
	password := getEnv("MYSQL_PASS", "")
	dsn := fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		username, password, host, port, database)

	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to MySQL: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}
