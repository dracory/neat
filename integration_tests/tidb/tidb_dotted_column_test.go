//go:build integration

package tidb_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestTiDBIntegrationDottedColumn_OrderBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	if db == nil {
		t.Skip("TiDB not available")
	}
	common.TestDottedColumnOrderBy(t, db)
}

func TestTiDBIntegrationDottedColumn_OrderByDesc(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	if db == nil {
		t.Skip("TiDB not available")
	}
	common.TestDottedColumnOrderByDesc(t, db)
}

func TestTiDBIntegrationDottedColumn_GroupBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	if db == nil {
		t.Skip("TiDB not available")
	}
	common.TestDottedColumnGroupBy(t, db)
}

func TestTiDBIntegrationDottedColumn_WhereColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	if db == nil {
		t.Skip("TiDB not available")
	}
	common.TestDottedColumnWhereColumn(t, db)
}

func TestTiDBIntegrationDottedColumn_Distinct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	if db == nil {
		t.Skip("TiDB not available")
	}
	common.TestDottedColumnDistinct(t, db)
}

func TestTiDBIntegrationDottedColumn_GroupByHaving(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	if db == nil {
		t.Skip("TiDB not available")
	}
	common.TestDottedColumnGroupByHaving(t, db)
}
