package mysql_test

import (
	"testing"

	"github.com/dracory/neat"
)

// TestMySQLDriverRegistration tests that MySQL driver can be registered and used
func TestMySQLDriverRegistration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dsn := "mysql://root:root@127.0.0.1:3306/test"
	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to MySQL: %v", err)
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

// TestMySQLConnection tests basic MySQL connection
func TestMySQLConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLConnection(t)
	if db == nil {
		t.Fatal("Database is nil")
	}

	query := db.Query()
	if query == nil {
		t.Fatal("Query builder is nil")
	}
}
