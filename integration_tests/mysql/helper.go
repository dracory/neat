package mysql_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/common"
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
	host := common.GetEnv("MYSQL_HOST", "127.0.0.1")
	port := common.GetEnvInt("MYSQL_PORT", 3306)
	database := common.GetEnv("MYSQL_DATABASE", "test")
	username := common.GetEnv("MYSQL_USER", "root")
	password := common.GetEnv("MYSQL_PASS", "root")

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

// SetupMySQLTest creates a database connection and sets up test tables for MySQL
func SetupMySQLTest(t *testing.T) *database.Database {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	host := common.GetEnv("MYSQL_HOST", "127.0.0.1")
	port := common.GetEnvInt("MYSQL_PORT", 3306)
	dbName := common.GetEnv("MYSQL_DATABASE", "test")
	username := common.GetEnv("MYSQL_USER", "root")
	password := common.GetEnv("MYSQL_PASS", "root")
	dsn := fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		username, password, host, port, dbName)

	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to MySQL: %v", err)
	}

	createMySQLTestTables(t, db)
	// Clean up any existing data before each test
	cleanupMySQLTestData(t, db)

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

// cleanupMySQLTestData removes all data from test tables
func cleanupMySQLTestData(t *testing.T, db *database.Database) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("cleanupMySQLTestData: DB(): %v", err)
	}
	stmts := []string{
		`DELETE IGNORE FROM users`,
		`DELETE IGNORE FROM addresses`,
		`DELETE IGNORE FROM books`,
		`DELETE IGNORE FROM peoples`,
		`DELETE IGNORE FROM json_datas`,
	}
	for _, stmt := range stmts {
		if _, err := sqlDB.Exec(stmt); err != nil {
			// Ignore errors if table doesn't exist
			continue
		}
	}
}

// createMySQLTestTables creates all tables required by the integration test models.
func createMySQLTestTables(t *testing.T, db *database.Database) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("createMySQLTestTables: DB(): %v", err)
	}
	stmts := []string{
		`DROP TABLE IF EXISTS books`,
		`DROP TABLE IF EXISTS addresses`,
		`DROP TABLE IF EXISTS users`,
		`DROP TABLE IF EXISTS peoples`,
		`DROP TABLE IF EXISTS json_datas`,
		`CREATE TABLE IF NOT EXISTS users (
			id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			name       VARCHAR(255) NOT NULL DEFAULT '',
			avatar     VARCHAR(255) NOT NULL DEFAULT '',
			bio        TEXT,
			votes      INT NOT NULL DEFAULT 0,
			deleted_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS addresses (
			id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			name       VARCHAR(255) NOT NULL DEFAULT '',
			user_id    BIGINT UNSIGNED,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS books (
			id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			name       VARCHAR(255) NOT NULL DEFAULT '',
			user_id    BIGINT UNSIGNED,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS peoples (
			id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			body       TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS json_datas (
			id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			data       JSON NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	for _, stmt := range stmts {
		if _, err := sqlDB.Exec(stmt); err != nil {
			t.Fatalf("createMySQLTestTables: %v", err)
		}
	}
}

// SetupMySQLConnection creates a database connection without setting up tables
func SetupMySQLConnection(t *testing.T) *database.Database {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	host := common.GetEnv("MYSQL_HOST", "127.0.0.1")
	port := common.GetEnvInt("MYSQL_PORT", 3306)
	database := common.GetEnv("MYSQL_DATABASE", "test")
	username := common.GetEnv("MYSQL_USER", "root")
	password := common.GetEnv("MYSQL_PASS", "root")
	dsn := fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		username, password, host, port, database)

	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to MySQL: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}
