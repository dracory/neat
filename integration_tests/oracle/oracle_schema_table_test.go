package oracle_test

import (
	"testing"
)

func TestOracleSchemaTableCreateHasDrop(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Oracle-specific table test - not yet implemented")
}

func TestOracleSchemaTableRename(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Oracle-specific table rename test - not yet implemented")
}

func TestOracleSchemaTableGetTables(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Oracle-specific get tables test - not yet implemented")
}

func TestOracleSchemaTableModify(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Oracle-specific table modify test - not yet implemented")
}
