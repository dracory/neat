package mysql_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestMySQLIntegrationDistinctSingleColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	common.TestDistinctSingleColumn(t, db)
}

func TestMySQLIntegrationDistinctWithCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	common.TestDistinctWithCount(t, db)
}
