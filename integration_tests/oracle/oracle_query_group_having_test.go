package oracle_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestOracleIntegrationGroupBySingleColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	common.TestGroupBySingleColumn(t, db)
}

func TestOracleIntegrationHavingClause(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	common.TestHavingClause(t, db)
}

func TestOracleIntegrationMultipleHavingClauses(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	common.TestMultipleHavingClauses(t, db)
}

func TestOracleIntegrationHavingWithSubqueryCallback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	common.TestHavingWithSubqueryCallback(t, db)
}

func TestOracleIntegrationHavingWithSubqueryInArgs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	common.TestHavingWithSubqueryInArgs(t, db)
}
