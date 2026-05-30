package postgres

import (
	"testing"

	"github.com/dracory/neat"
)

// TestPostgresDriverRegistration tests that PostgreSQL driver can be registered and used
func TestPostgresDriverRegistration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dsn := "postgres://test:test@127.0.0.1:55432/test?sslmode=disable"
	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("Database is nil")
	}

	query := db.Query()
	if query == nil {
		t.Fatal("Query builder is nil")
	}
}

// TestPostgresConnection tests basic PostgreSQL connection
func TestPostgresConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresConnection(t)
	if db == nil {
		t.Fatal("Database is nil")
	}

	query := db.Query()
	if query == nil {
		t.Fatal("Query builder is nil")
	}
}
