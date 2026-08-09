package tidb_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestTiDBIntegrationOrderByAscending(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestOrderByAscending(t, db)
}

func TestTiDBIntegrationOrderByDescending(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestOrderByDescending(t, db)
}

func TestTiDBIntegrationOrderByDescMethod(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestOrderByDescMethod(t, db)
}

func TestTiDBIntegrationMultipleOrderByClauses(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestMultipleOrderByClauses(t, db)
}

func TestTiDBIntegrationLimitClause(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestLimitClause(t, db)
}

func TestTiDBIntegrationLimitWithOrderBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestLimitWithOrderBy(t, db)
}

func TestTiDBIntegrationOffsetWithLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestOffsetWithLimit(t, db)
}
