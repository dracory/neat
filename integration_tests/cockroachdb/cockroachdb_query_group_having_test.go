//go:build integration

package cockroachdb_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestCockroachDBIntegrationGroupBySingleColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	common.TestGroupBySingleColumn(t, db)
}

func TestCockroachDBIntegrationHavingClause(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	common.TestHavingClause(t, db)
}

func TestCockroachDBIntegrationMultipleHavingClauses(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	common.TestMultipleHavingClauses(t, db)
}

func TestCockroachDBIntegrationHavingWithSubqueryCallback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Skipping - nested subquery placeholder numbering needs more complex solution")

	db := SetupCockroachDBTest(t)
	common.TestHavingWithSubqueryCallback(t, db)
}

func TestCockroachDBIntegrationHavingWithSubqueryInArgs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	common.TestHavingWithSubqueryInArgs(t, db)
}
