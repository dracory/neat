package turso

import (
	"testing"

	"github.com/dracory/neat"
)

// TestTursoDriverRegistration tests that Turso driver can be registered and used
func TestTursoDriverRegistration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dsn := getTursoDSN()
	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to Turso: %v", err)
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

// TestTursoConnection tests basic Turso connection
func TestTursoConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoConnection(t)
	if db == nil {
		t.Fatal("Database is nil")
	}

	query := db.Query()
	if query == nil {
		t.Fatal("Query builder is nil")
	}
}
