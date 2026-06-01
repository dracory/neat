package sqlite

import (
	"testing"

	"github.com/dracory/neat"
	"github.com/dracory/neat/database"
)

// SetupSQLiteConnection creates a database connection without setting up tables
func SetupSQLiteConnection(t *testing.T) *database.Database {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dsn := "sqlite://:memory:?multi_stmts=true"
	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to SQLite: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

// SetupSQLiteTest creates a database connection and sets up test tables
func SetupSQLiteTest(t *testing.T) *database.Database {
	db := SetupSQLiteConnection(t)
	createTestTables(t, db)
	// Clean up any existing data before each test
	cleanupTestData(t, db)
	return db
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
			user_id    INTEGER,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS books (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			name       TEXT NOT NULL DEFAULT '',
			user_id    INTEGER,
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
