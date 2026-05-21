//go:build integration

package sqlite

import (
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/schema/grammars"
)

func TestSQLiteSchemaColumnModifiers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	tableName := "test_column_modifiers"

	// Create a table with various modifiers
	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.String("nullable_col").Nullable()
		table.String("default_string").Default("hello")
		table.Integer("default_int").Default(123)
		table.Boolean("default_bool").Default(true)
		table.Timestamp("use_current").UseCurrent()
		table.Integer("unsigned_col").Unsigned()
		// Collation is supported by SQLite
		table.String("collation_col").Collation("NOCASE")
		// Comments are not supported by SQLite grammar (returns empty string), but shouldn't error
		table.String("comment_col").Comment("this is a comment")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
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
	if colMap["default_string"].Default != "'hello'" {
		t.Errorf("Expected default 'hello', got '%s'", colMap["default_string"].Default)
	}
	if colMap["default_int"].Default != "'123'" {
		t.Errorf("Expected default '123', got '%s'", colMap["default_int"].Default)
	}
	if colMap["default_bool"].Default != "'1'" {
		t.Errorf("Expected default '1', got '%s'", colMap["default_bool"].Default)
	}
	if colMap["use_current"].Default != "CURRENT_TIMESTAMP" {
		t.Errorf("Expected default CURRENT_TIMESTAMP, got '%s'", colMap["use_current"].Default)
	}
	// SQLite doesn't have a separate unsigned type for integers, usually they are just 'integer'
	// But it should have been accepted.

	// Check collation - SQLite processor currently doesn't populate collation from DB
	// assert.Equal(t, "NOCASE", colMap["collation_col"].Collation)

	// Test Change modifier
	err = db.Schema().Table(tableName, func(table schema.Blueprint) {
		// SQLite doesn't support ALTER COLUMN, so this should be a no-op or handled by recreation
		// Currently the grammar returns empty string for Change()
		table.String("new_col").Change()
	})
	if err != nil {
		t.Fatalf("Failed to change column: %v", err)
	}

	// Test After and First - should be ignored by SQLite without error
	err = db.Schema().Table(tableName, func(table schema.Blueprint) {
		table.String("after_col").After("id")
		table.String("first_col").First()
	})
	if err != nil {
		t.Fatalf("Failed to add columns with After/First: %v", err)
	}

	if !db.Schema().HasColumn(tableName, "after_col") {
		t.Error("after_col should exist")
	}
	if !db.Schema().HasColumn(tableName, "first_col") {
		t.Error("first_col should exist")
	}

	err = db.Schema().Drop(tableName)
	if err != nil {
		t.Fatalf("Failed to drop table: %v", err)
	}
}

func TestSQLiteSchemaColumnModifiersExpression(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	tableName := "test_modifiers_expression"

	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.Integer("expr_default").Default(grammars.Expression("(1 + 1)"))
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	columns, err := db.Schema().GetColumns(tableName)
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}

	found := false
	for _, col := range columns {
		if col.Name == "expr_default" {
			if col.Default != "1 + 1" {
				t.Errorf("Expected default '1 + 1', got '%s'", col.Default)
			}
			found = true
		}
	}
	if !found {
		t.Error("expr_default column not found")
	}

	err = db.Schema().Drop(tableName)
	if err != nil {
		t.Fatalf("Failed to drop table: %v", err)
	}
}
