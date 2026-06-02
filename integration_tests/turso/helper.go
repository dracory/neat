package turso

import (
	"os"
	"testing"

	"github.com/dracory/neat"
	"github.com/dracory/neat/database"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

// SetupTursoConnection creates a database connection without setting up tables
func SetupTursoConnection(t *testing.T) *database.Database {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Turso DSN format: libsql://[auth-token]@[host]/[database] or file://local.db
	// For local testing, we can use a file-based SQLite database
	dsn := getTursoDSN()
	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to Turso: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

// SetupTursoTest creates a database connection and sets up test tables
func SetupTursoTest(t *testing.T) *database.Database {
	db := SetupTursoConnection(t)
	createTestTables(t, db)
	// Clean up any existing data before each test
	cleanupTestData(t, db)
	return db
}

// getTursoDSN returns the Turso DSN from environment or uses a local file
func getTursoDSN() string {
	// Check for Turso environment variables
	url := os.Getenv("TURSO_URL")
	authToken := os.Getenv("TURSO_AUTH_TOKEN")

	if url != "" && authToken != "" {
		// Use remote Turso database
		// Format: libsql://auth-token@host/database
		return "libsql://" + authToken + "@" + url
	}

	// Fall back to in-memory SQLite database for testing
	// Turso is SQLite-based, so we can use SQLite driver for local testing
	// Using :memory: for in-memory database (faster than file-based)
	return "sqlite://:memory:?multi_stmts=true"
}

// cleanupTestData removes all data from test tables
func cleanupTestData(t *testing.T, db *database.Database) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("cleanupTestData: DB(): %v", err)
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
			t.Fatalf("cleanupTestData: %v", err)
		}
	}
}

// createTestTables creates all tables required by the integration test models.
func createTestTables(t *testing.T, db *database.Database) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("createTestTables: DB(): %v", err)
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			name       TEXT NOT NULL DEFAULT '',
			avatar     TEXT NOT NULL DEFAULT '',
			bio        TEXT,
			votes      INTEGER NOT NULL DEFAULT 0,
			deleted_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS addresses (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			name       TEXT NOT NULL DEFAULT '',
			user_id    INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS books (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			name       TEXT NOT NULL DEFAULT '',
			user_id    INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS peoples (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			body       TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS json_datas (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			data       TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, stmt := range stmts {
		if _, err := sqlDB.Exec(stmt); err != nil {
			t.Fatalf("createTestTables: %v", err)
		}
	}
}
