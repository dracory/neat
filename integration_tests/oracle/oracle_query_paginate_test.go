package oracle_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestOracleIntegrationPaginateFirstPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("TODO: Paginate tests have data cleanup issues with Oracle")
}

func TestOracleIntegrationPaginateSecondPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("TODO: Paginate tests have data cleanup issues with Oracle")
}

func TestOracleIntegrationPaginateWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	common.TestPaginateWithConditions(t, db)
}

func TestOracleIntegrationPaginateWithSelectAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	common.TestPaginateWithSelectAliases(t, db)
}
