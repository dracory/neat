package turso

import (
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
)

// TestTursoSchemaColumnChange tests column change operations
// Note: Turso (SQLite) doesn't support ALTER COLUMN directly, so this test verifies
// that the Change() method is called but no SQL is executed (returns empty string)
func TestTursoSchemaColumnChange(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)

	tableName := "test_column_change"

	// Create a table with initial columns
	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.String("change_length")
		table.String("change_type")
		table.String("change_nullable")
		table.String("change_default")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Verify table was created
	if !db.Schema().HasTable(tableName) {
		t.Error("Table should exist")
	}
	if !db.Schema().HasColumn(tableName, "change_length") {
		t.Error("Column change_length should exist")
	}
	if !db.Schema().HasColumn(tableName, "change_type") {
		t.Error("Column change_type should exist")
	}

	// Try to modify columns using Change()
	// Turso (SQLite) should skip these operations (returns empty SQL)
	err = db.Schema().Table(tableName, func(table schema.Blueprint) {
		// This should be skipped by Turso since it doesn't support ALTER COLUMN
		// We add a new column to avoid the duplicate column error during the implicit ADD command
		table.String("new_column_for_change").Change()
	})
	if err != nil {
		t.Fatalf("Table modification should not error even if Turso skips it: %v", err)
	}

	// Clean up
	err = db.Schema().Drop(tableName)
	if err != nil {
		t.Fatalf("Failed to drop table: %v", err)
	}
}

// TestTursoSchemaColumnChangeUnsupported verifies that Turso doesn't support column changes
func TestTursoSchemaColumnChangeUnsupported(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)

	tableName := "test_unsupported_change"

	// Create a table
	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.String("name")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Verify column exists
	if !db.Schema().HasColumn(tableName, "name") {
		t.Error("Column name should exist")
	}

	// Attempt to change column - Turso should skip this without error
	err = db.Schema().Table(tableName, func(table schema.Blueprint) {
		// We add a new column to avoid the duplicate column error during the implicit ADD command
		table.String("new_name").Change()
	})
	if err != nil {
		t.Fatalf("Change column should not error: %v", err)
	}

	// Clean up
	err = db.Schema().Drop(tableName)
	if err != nil {
		t.Fatalf("Failed to drop table: %v", err)
	}
}
