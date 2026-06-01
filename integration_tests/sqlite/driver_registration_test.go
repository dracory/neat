package sqlite

import (
	"testing"

	"github.com/dracory/neat"
)

// TestSQLiteDriverRegistration tests that SQLite driver can be registered and used
func TestSQLiteDriverRegistration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dsn := "sqlite://:memory:?multi_stmts=true"
	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to SQLite: %v", err)
	}
	defer func() { _ = db.Close() }()

	if db == nil {
		t.Fatal("Database is nil")
	}

	query := db.Query()
	if query == nil {
		t.Fatal("Query builder is nil")
	}
}

// TestSQLiteConnection tests basic SQLite connection
func TestSQLiteConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteConnection(t)
	if db == nil {
		t.Fatal("Database is nil")
	}

	query := db.Query()
	if query == nil {
		t.Fatal("Query builder is nil")
	}
}
