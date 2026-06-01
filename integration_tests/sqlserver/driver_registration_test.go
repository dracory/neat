package sqlserver

import (
	"testing"

	"github.com/dracory/neat"
)

// TestSQLServerDriverRegistration tests that SQL Server driver can be registered and used
func TestSQLServerDriverRegistration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dsn := "sqlserver://sa:YourStrong@Passw0rd@127.0.0.1:1433/test?encrypt=disable"
	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to SQL Server: %v", err)
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

// TestSQLServerConnection tests basic SQL Server connection
func TestSQLServerConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerConnection(t)
	if db == nil {
		t.Fatal("Database is nil")
	}

	query := db.Query()
	if query == nil {
		t.Fatal("Query builder is nil")
	}
}
