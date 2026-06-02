package postgres_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestPostgresIntegrationDistinctSingleColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestDistinctSingleColumn(t, db)
}

func TestPostgresIntegrationDistinctWithCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestDistinctWithCount(t, db)
}
