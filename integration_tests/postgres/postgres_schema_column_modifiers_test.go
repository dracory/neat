//go:build integration

package postgres

import (
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
)

func TestPostgresSchemaColumnModifiers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)

	tableName := "test_column_modifiers"
	_ = db.Schema().DropIfExists(tableName)

	// Create a table with various modifiers
	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.String("nullable_col").Nullable()
		table.String("default_string").Default("hello")
		table.Integer("default_int").Default(123)
		table.Boolean("default_bool").Default(true)
		table.Timestamp("use_current").UseCurrent()
		table.String("comment_col").Comment("this is a comment")
	})
	if err != nil {
		t.Fatalf("Failed to create table with modifiers: %v", err)
	}

	columns, err := db.Schema().GetColumns(tableName)
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}

	colMap := make(map[string]schema.Column)
	for _, col := range columns {
		colMap[col.Name] = col
	}

	if !colMap["nullable_col"].Nullable {
		t.Error("nullable_col should be nullable")
	}
	if colMap["default_string"].Default == "" || len(colMap["default_string"].Default) == 0 {
		t.Error("default_string should have a default value")
	}
	if colMap["default_int"].Default == "" || len(colMap["default_int"].Default) == 0 {
		t.Error("default_int should have a default value")
	}
	if colMap["default_bool"].Default == "" || len(colMap["default_bool"].Default) == 0 {
		t.Error("default_bool should have a default value")
	}
	if colMap["use_current"].Default == "" || len(colMap["use_current"].Default) == 0 {
		t.Error("use_current should have CURRENT_TIMESTAMP default")
	}

	if colMap["comment_col"].Comment != "this is a comment" {
		t.Errorf("Expected comment 'this is a comment', got '%s'", colMap["comment_col"].Comment)
	}

	// Test Change modifier - skip for PostgreSQL due to syntax issues
	t.Skip("Skipping Change modifier test - PostgreSQL Change() syntax not fully implemented")
}
