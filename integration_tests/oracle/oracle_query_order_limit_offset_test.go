package oracle_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestOracleIntegrationOrderByAscending(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	common.TestOrderByAscending(t, db)
}

func TestOracleIntegrationOrderByDescending(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	common.TestOrderByDescending(t, db)
}

func TestOracleIntegrationOrderByDescMethod(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	common.TestOrderByDescMethod(t, db)
}

func TestOracleIntegrationMultipleOrderByClauses(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	common.TestMultipleOrderByClauses(t, db)
}

func TestOracleIntegrationLimitClause(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	common.TestLimitClause(t, db)
}

func TestOracleIntegrationLimitWithOrderBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	common.TestLimitWithOrderBy(t, db)
}

func TestOracleIntegrationOffsetWithLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	common.TestOffsetWithLimit(t, db)
}
