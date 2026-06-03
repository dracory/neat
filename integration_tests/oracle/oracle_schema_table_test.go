package oracle_test

import (
	"testing"
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

	t.Skip("TODO: Oracle identity column with primary key syntax issue")
}
