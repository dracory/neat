//go:build integration

package tidb_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestTiDBIntegrationQueryAggregateSum(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestAggregateSum(t, db)
}

func TestTiDBIntegrationQueryAggregateSumWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestAggregateSumWithWhere(t, db)
}

func TestTiDBIntegrationQueryAggregateAvg(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestAggregateAvg(t, db)
}

func TestTiDBIntegrationQueryAggregateMax(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestAggregateMax(t, db)
}

func TestTiDBIntegrationQueryAggregateMin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestAggregateMin(t, db)
}

func TestTiDBIntegrationQueryAggregateGroupBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestAggregateGroupBy(t, db)
}

func TestTiDBIntegrationQueryAggregateInvalidColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestAggregateInvalidColumn(t, db)
}

func TestTiDBIntegrationQueryAggregateNilPointer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestAggregateNilPointer(t, db)
}

func TestTiDBIntegrationQueryAggregateEmptyResult(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestAggregateEmptyResult(t, db)
}

func TestTiDBIntegrationQueryAggregateNonNumericColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestAggregateNonNumericColumn(t, db)
}
