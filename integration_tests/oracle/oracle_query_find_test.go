package oracle_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestOracleIntegrationQueryFindById(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	common.TestFindById(t, db)
}

func TestOracleIntegrationQueryFindWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	common.TestFindWithWhere(t, db)
}

func TestOracleIntegrationQueryFindWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	common.TestFindWithConditions(t, db)
}
