package sqlserver

import (
	"fmt"
	"testing"
	"time"

	"github.com/dracory/neat"
	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/common"
)

// GetSQLServerConfig returns a SQL Server connection config from environment variables
func GetSQLServerConfig() neat.DBConfig {
	host := common.GetEnv("SQLSERVER_HOST", "127.0.0.1")
	port := common.GetEnvInt("SQLSERVER_PORT", 1433)
	database := common.GetEnv("SQLSERVER_DATABASE", "test")
	username := common.GetEnv("SQLSERVER_USER", "sa")
	password := common.GetEnv("SQLSERVER_PASS", "YourStrong@Passw0rd")

	return neat.DBConfig{
		Default: "sqlserver",
		Connections: map[string]neat.ConnectionConfig{
			"sqlserver": {
				Driver:   "sqlserver",
				Host:     host,
				Port:     port,
				Database: database,
				Username: username,
				Password: password,
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

// SetupSQLServerTest creates a database connection and registers cleanup.
func SetupSQLServerTest(t *testing.T) *database.Database {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	host := common.GetEnv("SQLSERVER_HOST", "127.0.0.1")
	port := common.GetEnvInt("SQLSERVER_PORT", 1433)
	db := common.GetEnv("SQLSERVER_DATABASE", "test")
	username := common.GetEnv("SQLSERVER_USER", "sa")
	password := common.GetEnv("SQLSERVER_PASS", "YourStrong@Passw0rd")
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d/%s?encrypt=disable",
		username, password, host, port, db)

	conn, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to SQL Server: %v", err)
	}

	createSQLServerTestTables(t, conn)
	cleanupSQLServerTestData(t, conn)

	t.Cleanup(func() {
		conn.Close()
	})

	return conn
}

// SetupSQLServerConnection creates a database connection without setting up tables
func SetupSQLServerConnection(t *testing.T) *database.Database {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	host := common.GetEnv("SQLSERVER_HOST", "127.0.0.1")
	port := common.GetEnvInt("SQLSERVER_PORT", 1433)
	database := common.GetEnv("SQLSERVER_DATABASE", "test")
	username := common.GetEnv("SQLSERVER_USER", "sa")
	password := common.GetEnv("SQLSERVER_PASS", "YourStrong@Passw0rd")
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d/%s?encrypt=disable",
		username, password, host, port, database)

	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to SQL Server: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// cleanupSQLServerTestData removes all data from test tables
func cleanupSQLServerTestData(t *testing.T, db *database.Database) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("cleanupSQLServerTestData: DB(): %v", err)
	}
	stmts := []string{
		`DELETE FROM users`,
		`DELETE FROM addresses`,
		`DELETE FROM books`,
		`DELETE FROM peoples`,
	}
	for _, stmt := range stmts {
		if _, err := sqlDB.Exec(stmt); err != nil {
			// Ignore errors if tables don't exist yet
			continue
		}
	}
}

// createSQLServerTestTables creates all tables required by the integration test models.
func createSQLServerTestTables(t *testing.T, db *database.Database) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("createSQLServerTestTables: DB(): %v", err)
	}
	stmts := []string{
		`IF OBJECT_ID('books', 'U') IS NOT NULL DROP TABLE books`,
		`IF OBJECT_ID('addresses', 'U') IS NOT NULL DROP TABLE addresses`,
		`IF OBJECT_ID('users', 'U') IS NOT NULL DROP TABLE users`,
		`IF OBJECT_ID('peoples', 'U') IS NOT NULL DROP TABLE peoples`,
		`CREATE TABLE users (
			id         BIGINT IDENTITY(1,1) PRIMARY KEY,
			name       NVARCHAR(255) NOT NULL DEFAULT '',
			avatar     NVARCHAR(255) NOT NULL DEFAULT '',
			bio        NVARCHAR(MAX),
			votes      INT NOT NULL DEFAULT 0,
			deleted_at DATETIME NULL,
			created_at DATETIME NOT NULL DEFAULT GETDATE(),
			updated_at DATETIME NOT NULL DEFAULT GETDATE()
		)`,
		`CREATE TABLE addresses (
			id         BIGINT IDENTITY(1,1) PRIMARY KEY,
			name       NVARCHAR(255) NOT NULL DEFAULT '',
			user_id    BIGINT NULL,
			created_at DATETIME NOT NULL DEFAULT GETDATE(),
			updated_at DATETIME NOT NULL DEFAULT GETDATE()
		)`,
		`CREATE TABLE books (
			id         BIGINT IDENTITY(1,1) PRIMARY KEY,
			name       NVARCHAR(255) NOT NULL DEFAULT '',
			user_id    BIGINT NULL,
			created_at DATETIME NOT NULL DEFAULT GETDATE(),
			updated_at DATETIME NOT NULL DEFAULT GETDATE()
		)`,
		`CREATE TABLE peoples (
			id         BIGINT IDENTITY(1,1) PRIMARY KEY,
			body       NVARCHAR(MAX) NOT NULL,
			created_at DATETIME NOT NULL DEFAULT GETDATE(),
			updated_at DATETIME NOT NULL DEFAULT GETDATE()
		)`,
	}
	for _, stmt := range stmts {
		if _, err := sqlDB.Exec(stmt); err != nil {
			t.Fatalf("createSQLServerTestTables: %v", err)
		}
	}
}
