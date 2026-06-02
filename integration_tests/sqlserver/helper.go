package sqlserver_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/dracory/neat"
	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/common"
	_ "github.com/microsoft/go-mssqldb"
)

var (
	tablesCreated bool
	tablesMutex   sync.Mutex
)

// GetSQLServerConfig builds a neat.DBConfig for SQL Server from environment variables.
// It reads the following variables with the shown defaults:
//
//   - SQLSERVER_HOST     (default: 127.0.0.1)
//   - SQLSERVER_PORT     (default: 1433)
//   - SQLSERVER_DATABASE (default: test)
//   - SQLSERVER_USER     (default: sa)
//   - SQLSERVER_PASS     (default: YourStrong@Passw0rd)
//
// The returned config uses "sqlserver" as the default connection name and applies
// sensible connection-pool defaults (5 idle / 10 open connections, 1-hour lifetimes).
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

// SetupSQLServerTest sets up a fully initialised SQL Server connection for integration tests.
//
// SQL Server containers (e.g. via Docker Compose) start with only the system databases
// (master, model, msdb, tempdb) present. Unlike MySQL/Postgres, they do not automatically
// create user databases from environment variables. This function therefore uses a two-step
// connection strategy:
//
//  1. Connect to the "master" system database (which always exists) and create the test
//     database if it does not already exist using an idempotent IF NOT EXISTS guard.
//  2. Connect directly to the test database, create the required test tables, and
//     truncate any pre-existing data so each test begins with a clean slate.
//
// The database connection is registered with t.Cleanup so it is closed automatically
// when the test (or sub-test) finishes.
func SetupSQLServerTest(t *testing.T) *database.Database {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	host := common.GetEnv("SQLSERVER_HOST", "127.0.0.1")
	port := common.GetEnvInt("SQLSERVER_PORT", 1433)
	db := common.GetEnv("SQLSERVER_DATABASE", "test")
	username := common.GetEnv("SQLSERVER_USER", "sa")
	password := common.GetEnv("SQLSERVER_PASS", "YourStrong@Passw0rd")

	// First connect to master to create the test database if it doesn't exist
	masterDSN := fmt.Sprintf("sqlserver://%s:%s@%s:%d/master?encrypt=disable",
		username, password, host, port)
	masterConn, err := neat.NewFromDSN(masterDSN)
	if err != nil {
		t.Fatalf("Failed to connect to SQL Server master: %v", err)
	}
	defer func() { _ = masterConn.Close() }()

	sqlDB, err := masterConn.DB()
	if err != nil {
		t.Fatalf("Failed to get SQL DB from master connection: %v", err)
	}

	// Create the test database if it doesn't exist
	_, err = sqlDB.Exec(fmt.Sprintf("IF NOT EXISTS (SELECT name FROM sys.databases WHERE name = '%s') CREATE DATABASE %s", db, db))
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Now connect to the test database
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d/%s?encrypt=disable",
		username, password, host, port, db)

	conn, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to SQL Server: %v", err)
	}

	createSQLServerTestTables(t, conn)
	cleanupSQLServerTestData(t, conn)

	t.Cleanup(func() {
		_ = conn.Close()
	})

	return conn
}

// SetupSQLServerConnection returns a bare SQL Server connection without creating tables
// or seeding data. It is intended for connection-level tests (e.g. pool settings,
// connection switching) that need a live database but manage their own schema.
//
// Like SetupSQLServerTest it uses a two-step connection strategy: it first connects to
// "master" to create the test database if absent, then returns a connection scoped to
// that test database. The connection is closed automatically via t.Cleanup.
func SetupSQLServerConnection(t *testing.T) *database.Database {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	host := common.GetEnv("SQLSERVER_HOST", "127.0.0.1")
	port := common.GetEnvInt("SQLSERVER_PORT", 1433)
	database := common.GetEnv("SQLSERVER_DATABASE", "test")
	username := common.GetEnv("SQLSERVER_USER", "sa")
	password := common.GetEnv("SQLSERVER_PASS", "YourStrong@Passw0rd")

	// First connect to master to create the test database if it doesn't exist
	masterDSN := fmt.Sprintf("sqlserver://%s:%s@%s:%d/master?encrypt=disable",
		username, password, host, port)
	masterConn, err := neat.NewFromDSN(masterDSN)
	if err != nil {
		t.Fatalf("Failed to connect to SQL Server master: %v", err)
	}
	defer func() { _ = masterConn.Close() }()

	sqlDB, err := masterConn.DB()
	if err != nil {
		t.Fatalf("Failed to get SQL DB from master connection: %v", err)
	}

	// Create the test database if it doesn't exist
	_, err = sqlDB.Exec(fmt.Sprintf("IF NOT EXISTS (SELECT name FROM sys.databases WHERE name = '%s') CREATE DATABASE %s", database, database))
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Now connect to the test database
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d/%s?encrypt=disable",
		username, password, host, port, database)

	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to SQL Server: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

// cleanupSQLServerTestData truncates all test tables so that each test starts
// with an empty dataset. Errors are silently ignored so the function is safe to
// call before the tables have been created for the first time.
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

// createSQLServerTestTables drops (if they exist) and recreates all tables used by
// the integration test models: users, addresses, books, and peoples. Tables are
// dropped in a safe order that respects foreign-key dependencies (books and addresses
// before users). The schema uses BIGINT IDENTITY primary keys and DATETIME2 columns
// (preferred over the legacy DATETIME type for better precision and range).
// This function uses a mutex to ensure tables are only created once across all tests.
func createSQLServerTestTables(t *testing.T, db *database.Database) {
	t.Helper()

	tablesMutex.Lock()
	defer tablesMutex.Unlock()

	if tablesCreated {
		// Tables already created, just cleanup data
		cleanupSQLServerTestData(t, db)
		return
	}

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
			deleted_at DATETIME2 NULL,
			created_at DATETIME2 NOT NULL DEFAULT GETDATE(),
			updated_at DATETIME2 NOT NULL DEFAULT GETDATE()
		)`,
		`CREATE TABLE addresses (
			id         BIGINT IDENTITY(1,1) PRIMARY KEY,
			name       NVARCHAR(255) NOT NULL DEFAULT '',
			user_id    BIGINT NULL,
			created_at DATETIME2 NOT NULL DEFAULT GETDATE(),
			updated_at DATETIME2 NOT NULL DEFAULT GETDATE()
		)`,
		`CREATE TABLE books (
			id         BIGINT IDENTITY(1,1) PRIMARY KEY,
			name       NVARCHAR(255) NOT NULL DEFAULT '',
			user_id    BIGINT NULL,
			created_at DATETIME2 NOT NULL DEFAULT GETDATE(),
			updated_at DATETIME2 NOT NULL DEFAULT GETDATE()
		)`,
		`CREATE TABLE peoples (
			id         BIGINT IDENTITY(1,1) PRIMARY KEY,
			body       NVARCHAR(MAX) NOT NULL,
			created_at DATETIME2 NOT NULL DEFAULT GETDATE(),
			updated_at DATETIME2 NOT NULL DEFAULT GETDATE()
		)`,
	}
	for _, stmt := range stmts {
		if _, err := sqlDB.Exec(stmt); err != nil {
			t.Fatalf("createSQLServerTestTables: %v", err)
		}
	}

	tablesCreated = true
}
