//go:build integration

package mysql_test

import (
	"testing"

	"github.com/dracory/neat"
	_ "github.com/go-sql-driver/mysql"
)

// TestMySQLErrorHandlingWrongPassword tests that connecting to MySQL with a
// wrong password returns an error.
func TestMySQLErrorHandlingWrongPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test with wrong password
	dsn := "mysql://root:wrongpassword@127.0.0.1:3306/test?charset=utf8mb4"
	_, err := neat.NewFromDSN(dsn)
	if err == nil {
		t.Error("Expected error for wrong MySQL password, got nil")
	}
}

// TestMySQLErrorHandlingConnection tests connection error handling against a
// live MySQL service: connecting with valid credentials succeeds, and
// requesting a non-existent connection returns an error.
func TestMySQLErrorHandlingConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dsn := "mysql://root:root@127.0.0.1:3306/test"
	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Test getting non-existent connection
	_, err = db.Connection("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent connection, got nil")
	}
}
