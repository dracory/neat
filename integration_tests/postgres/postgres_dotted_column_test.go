package postgres_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestPostgresIntegrationDottedColumn_OrderBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	if db == nil {
		t.Skip("Postgres not available")
	}
	common.TestDottedColumnOrderBy(t, db)
}

func TestPostgresIntegrationDottedColumn_OrderByDesc(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	if db == nil {
		t.Skip("Postgres not available")
	}
	common.TestDottedColumnOrderByDesc(t, db)
}

func TestPostgresIntegrationDottedColumn_GroupBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	if db == nil {
		t.Skip("Postgres not available")
	}
	common.TestDottedColumnGroupBy(t, db)
}

func TestPostgresIntegrationDottedColumn_WhereColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	if db == nil {
		t.Skip("Postgres not available")
	}
	common.TestDottedColumnWhereColumn(t, db)
}

func TestPostgresIntegrationDottedColumn_Distinct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	if db == nil {
		t.Skip("Postgres not available")
	}
	common.TestDottedColumnDistinct(t, db)
}

func TestPostgresIntegrationDottedColumn_GroupByHaving(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	if db == nil {
		t.Skip("Postgres not available")
	}
	common.TestDottedColumnGroupByHaving(t, db)
}
