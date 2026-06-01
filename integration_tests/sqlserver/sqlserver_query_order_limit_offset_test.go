package sqlserver

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestSQLServerIntegrationOrderByAscending(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestOrderByAscending(t, db)
}

func TestSQLServerIntegrationOrderByDescending(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestOrderByDescending(t, db)
}

func TestSQLServerIntegrationOrderByDescMethod(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestOrderByDescMethod(t, db)
}

func TestSQLServerIntegrationMultipleOrderByClauses(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestMultipleOrderByClauses(t, db)
}

func TestSQLServerIntegrationLimitClause(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestLimitClause(t, db)
}

func TestSQLServerIntegrationLimitWithOrderBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestLimitWithOrderBy(t, db)
}

func TestSQLServerIntegrationOffsetWithLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("SQL Server cannot combine TOP and OFFSET in same query - skipping")
}
