package oracle_test

import (
	"context"
	"testing"

	"github.com/dracory/neat/database/driver"
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
	oracleDriver := driver.NewOracle()
	err := oracleDriver.Ping(context.Background(), db)
	if err != nil {
		t.Fatalf("Failed to ping Oracle: %v", err)
	}

	// Test dialect
	if oracleDriver.Dialect() != "oracle" {
		t.Errorf("Expected dialect 'oracle', got '%s'", oracleDriver.Dialect())
	}

	// Test placeholder
	if oracleDriver.Placeholder(1) != ":1" {
		t.Errorf("Expected placeholder ':1', got '%s'", oracleDriver.Placeholder(1))
	}
}
