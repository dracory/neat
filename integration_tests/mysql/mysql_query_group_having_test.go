package mysql

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestMySQLIntegrationGroupBySingleColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	common.TestGroupBySingleColumn(t, db)
}

func TestMySQLIntegrationHavingClause(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	common.TestHavingClause(t, db)
}

func TestMySQLIntegrationMultipleHavingClauses(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	common.TestMultipleHavingClauses(t, db)
}

func TestMySQLIntegrationHavingWithSubqueryCallback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	common.TestHavingWithSubqueryCallback(t, db)
}

func TestMySQLIntegrationHavingWithSubqueryInArgs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	common.TestHavingWithSubqueryInArgs(t, db)
}
