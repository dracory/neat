
package postgres

import (
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
)

// TestPostgresSchemaColumnMethods tests column method operations
func TestPostgresSchemaColumnMethods(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)

	tableName := "test_column_methods"

	// Create a table with multiple columns
	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.String("name")
		table.String("age")
		table.String("weight")
		table.String("height")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Verify table was created
	if !db.Schema().HasTable(tableName) {
		t.Error("Table should exist")
	}

	// Test HasColumn - verify individual columns exist
	t.Run("HasColumn", func(t *testing.T) {
		if !db.Schema().HasColumn(tableName, "name") {
			t.Error("Column 'name' should exist")
		}
		if !db.Schema().HasColumn(tableName, "age") {
			t.Error("Column 'age' should exist")
		}
		if !db.Schema().HasColumn(tableName, "weight") {
			t.Error("Column 'weight' should exist")
		}
		if !db.Schema().HasColumn(tableName, "height") {
			t.Error("Column 'height' should exist")
		}
		if db.Schema().HasColumn(tableName, "nonexistent") {
			t.Error("Nonexistent column should not exist")
		}
	})

	// Test HasColumns - verify multiple columns exist
	t.Run("HasColumns", func(t *testing.T) {
		if !db.Schema().HasColumns(tableName, []string{"name", "age", "weight"}) {
			t.Error("All columns should exist")
		}
		if !db.Schema().HasColumns(tableName, []string{"id", "name"}) {
			t.Error("ID and name should exist")
		}
		if db.Schema().HasColumns(tableName, []string{"name", "nonexistent"}) {
			t.Error("Should fail if one column doesn't exist")
		}
		if db.Schema().HasColumns(tableName, []string{"nonexistent1", "nonexistent2"}) {
			t.Error("Should fail if all columns don't exist")
		}
	})

	// Test DropColumn using Table method
	t.Run("DropColumn", func(t *testing.T) {
		err := db.Schema().Table(tableName, func(table schema.Blueprint) {
			table.DropColumn("name", "age")
		})
		if err != nil {
			t.Fatalf("Failed to drop columns: %v", err)
		}

		// Verify columns were dropped
		if db.Schema().HasColumn(tableName, "name") {
			t.Error("Column 'name' should not exist after drop")
		}
		if db.Schema().HasColumn(tableName, "age") {
			t.Error("Column 'age' should not exist after drop")
		}
		if !db.Schema().HasColumn(tableName, "weight") {
			t.Error("Column 'weight' should still exist")
		}
		if !db.Schema().HasColumn(tableName, "height") {
			t.Error("Column 'height' should still exist")
		}
	})

	// Test DropColumns method
	t.Run("DropColumns", func(t *testing.T) {
		err := db.Schema().DropColumns(tableName, []string{"weight"})
		if err != nil {
			t.Fatalf("Failed to drop columns via DropColumns: %v", err)
		}

		// Verify column was dropped
		if db.Schema().HasColumn(tableName, "weight") {
			t.Error("Column 'weight' should not exist after DropColumns")
		}
		if !db.Schema().HasColumn(tableName, "height") {
			t.Error("Column 'height' should still exist")
		}
	})

	// Test HasColumns after all drops
	t.Run("HasColumnsAfterDrops", func(t *testing.T) {
		if db.Schema().HasColumns(tableName, []string{"name", "age", "weight"}) {
			t.Error("All dropped columns should not exist")
		}
		if !db.Schema().HasColumns(tableName, []string{"id", "height"}) {
			t.Error("Remaining columns should exist")
		}
	})

	// Clean up
	err = db.Schema().Drop(tableName)
	if err != nil {
		t.Fatalf("Failed to drop table: %v", err)
	}
}

// TestPostgresSchemaColumnMethodsSingle tests column methods with single column operations
func TestPostgresSchemaColumnMethodsSingle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)

	tableName := "test_column_methods_single"

	// Create a table
	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.String("single_column")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test single column operations
	t.Run("SingleColumn", func(t *testing.T) {
		if !db.Schema().HasColumn(tableName, "single_column") {
			t.Error("single_column should exist")
		}
		if !db.Schema().HasColumns(tableName, []string{"single_column"}) {
			t.Error("single_column should exist in HasColumns")
		}

		// Drop single column
		err := db.Schema().Table(tableName, func(table schema.Blueprint) {
			table.DropColumn("single_column")
		})
		if err != nil {
			t.Fatalf("Failed to drop single column: %v", err)
		}

		if db.Schema().HasColumn(tableName, "single_column") {
			t.Error("single_column should not exist after drop")
		}
		if db.Schema().HasColumns(tableName, []string{"single_column"}) {
			t.Error("single_column should not exist in HasColumns after drop")
		}
	})

	// Clean up
	err = db.Schema().Drop(tableName)
	if err != nil {
		t.Fatalf("Failed to drop table: %v", err)
	}
}

// TestPostgresSchemaColumnMethodsEmpty tests edge cases with empty column lists
func TestPostgresSchemaColumnMethodsEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)

	tableName := "test_column_methods_empty"

	// Create a table
	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.String("test_column")
	})
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test empty column list
	t.Run("EmptyColumnList", func(t *testing.T) {
		// HasColumns with empty list should return true (no columns to check)
		if !db.Schema().HasColumns(tableName, []string{}) {
			t.Error("HasColumns with empty list should return true")
		}
	})

	// Clean up
	err = db.Schema().Drop(tableName)
	if err != nil {
		t.Fatalf("Failed to drop table: %v", err)
	}
}
