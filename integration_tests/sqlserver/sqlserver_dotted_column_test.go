package sqlserver_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestSQLServerIntegrationDottedColumn_OrderBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	if db == nil {
		t.Skip("SQL Server not available")
	}
	common.TestDottedColumnOrderBy(t, db)
}

func TestSQLServerIntegrationDottedColumn_OrderByDesc(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	if db == nil {
		t.Skip("SQL Server not available")
	}
	common.TestDottedColumnOrderByDesc(t, db)
}

func TestSQLServerIntegrationDottedColumn_GroupBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	if db == nil {
		t.Skip("SQL Server not available")
	}
	common.TestDottedColumnGroupBy(t, db)
}

func TestSQLServerIntegrationDottedColumn_WhereColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	if db == nil {
		t.Skip("SQL Server not available")
	}
	common.TestDottedColumnWhereColumn(t, db)
}

func TestSQLServerIntegrationDottedColumn_Distinct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	if db == nil {
		t.Skip("SQL Server not available")
	}
	common.TestDottedColumnDistinct(t, db)
}

func TestSQLServerIntegrationDottedColumn_GroupByHaving(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	if db == nil {
		t.Skip("SQL Server not available")
	}
	common.TestDottedColumnGroupByHaving(t, db)
}
