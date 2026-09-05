//go:build integration

package postgres_test

import (
	"testing"

	"github.com/dracory/neat"
	_ "github.com/lib/pq"
)

// TestPostgreSQLErrorHandlingWrongPassword tests that connecting to PostgreSQL
// with a wrong password returns an error.
func TestPostgreSQLErrorHandlingWrongPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test with wrong password
	dsn := "postgres://test:wrongpassword@127.0.0.1:55432/test?sslmode=disable"
	_, err := neat.NewFromDSN(dsn)
	if err == nil {
		t.Error("Expected error for wrong PostgreSQL password, got nil")
	}
}
