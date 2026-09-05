//go:build integration

package common_test

import (
	"testing"

	"github.com/dracory/neat"
	_ "modernc.org/sqlite"
)

// TestErrorHandlingIntegration tests error handling scenarios across databases
func TestErrorHandlingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("InvalidDSN", func(t *testing.T) {
		_, err := neat.NewFromDSN("invalid://dsn")
		if err == nil {
			t.Error("Expected error for invalid DSN, got nil")
		}
	})

	t.Run("InvalidConfig", func(t *testing.T) {
		config := neat.DBConfig{
			Default: "invalid",
			Connections: map[string]neat.ConnectionConfig{
				"invalid": {
					Driver:   "invalid_driver",
					Host:     "localhost",
					Port:     3306,
					Database: "test",
					Username: "root",
					Password: "root",
				},
			},
		}
		_, err := neat.New(config)
		if err == nil {
			t.Error("Expected error for invalid driver, got nil")
		}
	})

	t.Run("EmptyConfig", func(t *testing.T) {
		config := neat.DBConfig{
			Default:     "",
			Connections: map[string]neat.ConnectionConfig{},
		}
		_, err := neat.New(config)
		if err == nil {
			t.Error("Expected error for empty config, got nil")
		}
	})
}

// TestErrorHandlingIntegrationSQLite tests SQLite-specific error handling
func TestErrorHandlingIntegrationSQLite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test with invalid database path
	dsn := "sqlite:///invalid/path/to/database.db"
	_, err := neat.NewFromDSN(dsn)
	if err == nil {
		t.Error("Expected error for invalid SQLite path, got nil")
	}
}

// TestErrorHandlingIntegrationQuery tests query error handling
func TestErrorHandlingIntegrationQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dsn := "sqlite://:memory:?multi_stmts=true"
	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to SQLite: %v", err)
	}
	defer func() { _ = db.Close() }()

	query := db.Query()
	if query == nil {
		t.Fatal("Query builder is nil")
	}

	// Test querying non-existent table
	tableQuery := query.Table("nonexistent_table")
	if tableQuery == nil {
		t.Fatal("Table query should not be nil even for non-existent table")
	}
}
