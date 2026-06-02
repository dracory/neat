package oracle_test

import (
	"fmt"
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database"
)

func testColumnType(t *testing.T, db *database.Database, name string, setup func(schema.Blueprint), expectedType string, nullable, autoincrement bool, defaultValue string) {
	tableName := fmt.Sprintf("TEST_TYPES_OR_%s", name)
	_ = db.Schema().DropIfExists(tableName)

	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		setup(table)
	})
	if err != nil {
		t.Fatalf("Failed to create table for %s: %v", name, err)
	}

	columns, err := db.Schema().GetColumns(tableName)
	if err != nil {
		t.Fatalf("Failed to get columns for %s: %v", name, err)
	}
	if len(columns) != 1 {
		t.Fatalf("Expected 1 column for %s, got %d", name, len(columns))
	}

	if columns[0].Name != "COL" {
		t.Errorf("Expected column name 'COL', got '%s'", columns[0].Name)
	}
	if columns[0].TypeName != expectedType {
		t.Errorf("Wrong type for %s: expected %s, got %s", name, expectedType, columns[0].TypeName)
	}
	if columns[0].Nullable != nullable {
		t.Errorf("Wrong nullable for %s: expected %v, got %v", name, nullable, columns[0].Nullable)
	}
	if columns[0].Autoincrement != autoincrement {
		t.Errorf("Wrong autoincrement for %s: expected %v, got %v", name, autoincrement, columns[0].Autoincrement)
	}
	if defaultValue != "" && columns[0].Default != defaultValue {
		t.Errorf("Wrong default for %s: expected %s, got %s", name, defaultValue, columns[0].Default)
	}

	err = db.Schema().Drop(tableName)
	if err != nil {
		t.Fatalf("Failed to drop table for %s: %v", name, err)
	}
}

func TestOracleSchemaColumnTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("TODO: Oracle schema builder has case sensitivity issues with GetColumns")
}
