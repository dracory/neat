//go:build integration

package oracle_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestOracleIntegrationDottedColumn_OrderBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	if db == nil {
		t.Skip("Oracle not available")
	}
	common.TestDottedColumnOrderBy(t, db)
}

func TestOracleIntegrationDottedColumn_OrderByDesc(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	if db == nil {
		t.Skip("Oracle not available")
	}
	common.TestDottedColumnOrderByDesc(t, db)
}

func TestOracleIntegrationDottedColumn_GroupBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	if db == nil {
		t.Skip("Oracle not available")
	}
	common.TestDottedColumnGroupBy(t, db)
}

func TestOracleIntegrationDottedColumn_WhereColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	if db == nil {
		t.Skip("Oracle not available")
	}
	common.TestDottedColumnWhereColumn(t, db)
}

func TestOracleIntegrationDottedColumn_Distinct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	if db == nil {
		t.Skip("Oracle not available")
	}
	common.TestDottedColumnDistinct(t, db)
}

func TestOracleIntegrationDottedColumn_GroupByHaving(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	if db == nil {
		t.Skip("Oracle not available")
	}
	common.TestDottedColumnGroupByHaving(t, db)
}
