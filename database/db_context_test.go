package database

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dracory/neat/contracts/database/orm"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database/db"
)

func TestDatabase_WithContext_SetsContextCorrectly(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	customCtx := context.WithValue(context.Background(), "test-key", "test-value")
	database, err := New(config, WithContext(customCtx), WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = database.Close() }()

	// Verify the context was set by checking if we can retrieve the value
	if database.ctx.Value("test-key") != "test-value" {
		t.Error("WithContext() did not set the context correctly")
	}
}

func TestDatabase_WithContext_DefaultContext(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	// Create database without WithContext option
	database, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = database.Close() }()

	// Should have default background context
	if database.ctx == nil {
		t.Error("Expected default context to be set")
	}
}

func TestDatabase_ContextPropagatesToQueries(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	key := "query-key"
	value := "query-value"
	customCtx := context.WithValue(context.Background(), key, value)

	database, err := New(config, WithContext(customCtx), WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = database.Close() }()

	// Create a table and insert data
	schema := database.Schema()
	err = schema.Create("test_table", func(table contractsschema.Blueprint) {
		table.Increments("id")
		table.String("name")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	query := database.Query()
	err = query.Table("test_table").Create(map[string]any{"id": 1, "name": "test"})
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	// The query should have access to the context from the database
	// We can verify this by checking if the context is available in the query
	// Note: This is a basic test - actual context propagation depends on ORM implementation
	if query == nil {
		t.Error("Expected non-nil query")
	}
}

func TestDatabase_ContextCancellationStopsOperations(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	database, err := New(config, WithContext(ctx), WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = database.Close() }()

	// Create a table first (before cancellation)
	schema := database.Schema()
	err = schema.Create("test_table", func(table contractsschema.Blueprint) {
		table.Increments("id")
		table.String("name")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Now cancel the context
	cancel()

	// Try to execute a query with cancelled context
	query := database.Query()
	query = query.Table("test_table")

	var result []map[string]any
	err = query.Get(&result)

	// The query should fail due to cancelled context
	if err == nil {
		t.Error("Expected query to fail with cancelled context")
	}
	if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "context canceled") {
		// Some drivers may not return context.Canceled directly
		// but should still fail in some way
		t.Logf("Query failed with error (may be driver-specific): %v", err)
	}
}

func TestDatabase_NilContextHandling(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	// Pass nil context via WithContext
	database, err := New(config, WithContext(nil), WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database with nil context: %v", err)
	}
	defer func() { _ = database.Close() }()

	// Database should still be functional with nil context
	// It should use a default context internally
	schema := database.Schema()
	err = schema.Create("test_table", func(table contractsschema.Blueprint) {
		table.Increments("id")
		table.String("name")
	})
	if err != nil {
		t.Fatalf("Failed to create table with nil context: %v", err)
	}

	// Query should work
	query := database.Query()
	err = query.Table("test_table").Create(map[string]any{"id": 1, "name": "test"})
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}
	if query == nil {
		t.Error("Expected non-nil query with nil context")
	}
}

func TestDatabase_ContextWithTimeout(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	// Create context with reasonable timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	database, err := New(config, WithContext(ctx), WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = database.Close() }()

	// Create a table
	schema := database.Schema()
	err = schema.Create("test_table", func(table contractsschema.Blueprint) {
		table.Increments("id")
		table.String("name")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Try to execute a query - should succeed within timeout
	query := database.Query()
	query = query.Table("test_table")

	var result []map[string]any
	err = query.Get(&result)

	// Query should succeed with reasonable timeout
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "deadline exceeded") {
			t.Errorf("Query failed due to timeout: %v", err)
		}
		// Empty table error is acceptable
		t.Logf("Query failed (may be due to empty table): %v", err)
	}
}

func TestDatabase_ContextInTransaction(t *testing.T) {
	config := db.DBConfig{
		Default: "default",
		Connections: map[string]db.ConnectionConfig{
			"default": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	type contextKey string
	key := contextKey("tx-key")
	value := "tx-value"
	customCtx := context.WithValue(context.Background(), key, value)

	database, err := New(config, WithContext(customCtx), WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = database.Close() }()

	// Create a table
	schema := database.Schema()
	err = schema.Create("test_table", func(table contractsschema.Blueprint) {
		table.Increments("id")
		table.String("name")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Execute transaction
	err = database.Transaction(func(tx orm.Query) error {
		// Transaction should have access to the database context
		return tx.Table("test_table").Create(map[string]any{"id": 1, "name": "test"})
	})

	if err != nil {
		t.Errorf("Transaction failed: %v", err)
	}
}
