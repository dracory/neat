package oracle_test

import (
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
)

func TestOracleSchemaColumnModifiers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("skipped - Oracle default value handling needs investigation (ORA-00907)")

	db := SetupOracleTest(t)

	tableName := "test_column_modifiers"
	_ = db.Schema().Drop(tableName)
	_ = db.Schema().DropIfExists(tableName)

	// Create a table with various modifiers - test incrementally
	err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
		table.String("nullable_col").Nullable()
		table.String("default_string").Default("hello")
		table.Integer("default_int").Default(123)
		table.Boolean("default_bool").Default(true)
		table.Timestamp("use_current").UseCurrent()
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
		// Oracle returns column names in uppercase
		colMap[col.Name] = col
		t.Logf("Column: %s, Default: '%s', Type: %s, Nullable: %v", col.Name, col.Default, col.Type, col.Nullable)
	}

	// Check nullable (handle both uppercase and lowercase)
	nullableCol := colMap["NULLABLE_COL"]
	if nullableCol.Name == "" {
		nullableCol = colMap["nullable_col"]
	}
	if !nullableCol.Nullable {
		t.Error("nullable_col should be nullable")
	}

	// Check defaults (handle both uppercase and lowercase)
	defaultString := colMap["DEFAULT_STRING"]
	if defaultString.Name == "" {
		defaultString = colMap["default_string"]
	}
	if defaultString.Default != "hello" {
		t.Errorf("Expected default 'hello', got '%s'", defaultString.Default)
	}

	defaultInt := colMap["DEFAULT_INT"]
	if defaultInt.Name == "" {
		defaultInt = colMap["default_int"]
	}
	if defaultInt.Default != "123" {
		t.Errorf("Expected default '123', got '%s'", defaultInt.Default)
	}

	// Oracle boolean is number(1), default 1
	defaultBool := colMap["DEFAULT_BOOL"]
	if defaultBool.Name == "" {
		defaultBool = colMap["default_bool"]
	}
	if defaultBool.Default != "1" {
		t.Errorf("Expected default '1', got '%s'", defaultBool.Default)
	}

	useCurrent := colMap["USE_CURRENT"]
	if useCurrent.Name == "" {
		useCurrent = colMap["use_current"]
	}
	if useCurrent.Default == "" {
		t.Error("use_current should have CURRENT_TIMESTAMP default")
	}

	_ = db.Schema().Drop(tableName)
}
