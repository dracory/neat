package oracle_test

import (
	"context"
	"testing"
)

func TestOracleIntegrationConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleConnection(t)
	if db == nil {
		t.Fatal("Database is nil")
	}

	// Test ping
	driver := NewOracle()
	err := driver.Ping(context.Background(), db)
	if err != nil {
		t.Fatalf("Failed to ping Oracle: %v", err)
	}

	// Test dialect
	if driver.Dialect() != "oracle" {
		t.Errorf("Expected dialect 'oracle', got '%s'", driver.Dialect())
	}

	// Test placeholder
	if driver.Placeholder(1) != ":1" {
		t.Errorf("Expected placeholder ':1', got '%s'", driver.Placeholder(1))
	}
}
