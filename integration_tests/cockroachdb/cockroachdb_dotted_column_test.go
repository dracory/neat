package cockroachdb_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestCockroachDBIntegrationDottedColumnOrderBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	if db == nil {
		t.Skip("CockroachDB not available")
	}
	common.TestDottedColumnOrderBy(t, db)
}

func TestCockroachDBIntegrationDottedColumnOrderByDesc(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	if db == nil {
		t.Skip("CockroachDB not available")
	}
	common.TestDottedColumnOrderByDesc(t, db)
}

func TestCockroachDBIntegrationDottedColumnGroupBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	if db == nil {
		t.Skip("CockroachDB not available")
	}
	common.TestDottedColumnGroupBy(t, db)
}

func TestCockroachDBIntegrationDottedColumnWhereColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	if db == nil {
		t.Skip("CockroachDB not available")
	}
	common.TestDottedColumnWhereColumn(t, db)
}

func TestCockroachDBIntegrationDottedColumnDistinct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	if db == nil {
		t.Skip("CockroachDB not available")
	}
	common.TestDottedColumnDistinct(t, db)
}

func TestCockroachDBIntegrationDottedColumnGroupByHaving(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	if db == nil {
		t.Skip("CockroachDB not available")
	}
	common.TestDottedColumnGroupByHaving(t, db)
}
