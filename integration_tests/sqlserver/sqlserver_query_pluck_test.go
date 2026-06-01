package sqlserver

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestSQLServerIntegrationPluckSingleColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestPluckSingleColumn(t, db)
}

func TestSQLServerIntegrationPluckWithDistinct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestPluckWithDistinct(t, db)
}
