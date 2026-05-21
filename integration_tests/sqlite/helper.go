//go:build integration

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
		db.Close()
	})

	return db
}
