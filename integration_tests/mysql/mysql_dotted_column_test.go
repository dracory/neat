package mysql_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestMySQLIntegrationDottedColumn_OrderBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	if db == nil {
		t.Skip("MySQL not available")
	}
	common.TestDottedColumnOrderBy(t, db)
}

func TestMySQLIntegrationDottedColumn_OrderByDesc(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	if db == nil {
		t.Skip("MySQL not available")
	}
	common.TestDottedColumnOrderByDesc(t, db)
}

func TestMySQLIntegrationDottedColumn_GroupBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	if db == nil {
		t.Skip("MySQL not available")
	}
	common.TestDottedColumnGroupBy(t, db)
}

func TestMySQLIntegrationDottedColumn_WhereColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	if db == nil {
		t.Skip("MySQL not available")
	}
	common.TestDottedColumnWhereColumn(t, db)
}

func TestMySQLIntegrationDottedColumn_Distinct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	if db == nil {
		t.Skip("MySQL not available")
	}
	common.TestDottedColumnDistinct(t, db)
}

func TestMySQLIntegrationDottedColumn_GroupByHaving(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	if db == nil {
		t.Skip("MySQL not available")
	}
	common.TestDottedColumnGroupByHaving(t, db)
}
