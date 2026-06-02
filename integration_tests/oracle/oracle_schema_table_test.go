package oracle_test

import (
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
)

func TestOracleSchemaTableCreateHasDrop(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("TODO: Oracle schema builder has case sensitivity issues with HasTable and table listing")
}

func TestOracleSchemaTableRename(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("TODO: Oracle schema builder has case sensitivity issues with HasTable")
}

func TestOracleSchemaTableGetTables(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("TODO: Oracle schema builder has case sensitivity issues with table listing")
}

func TestOracleSchemaTableModify(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	tableName := "MODIFY_TABLE"
	_ = db.Schema().DropIfExists(tableName)

	if err := db.Schema().Create(tableName, func(table schema.Blueprint) {
		table.ID()
	}); err != nil {
		t.Fatalf("Failed to create modify table: %v", err)
	}

	if err := db.Schema().Table(tableName, func(table schema.Blueprint) {
		table.String("name")
	}); err != nil {
		t.Fatalf("Failed to modify table: %v", err)
	}

	_ = db.Schema().Drop(tableName)
}
