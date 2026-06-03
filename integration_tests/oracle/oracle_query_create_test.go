package oracle_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestOracleIntegrationQueryCreateByStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	common.TestQueryCreateByStruct(t, db)
}

func TestOracleIntegrationQueryBatchCreateByStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	common.TestQueryBatchCreateByStruct(t, db)
}

func TestOracleIntegrationQueryInsertGetIdByStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	common.TestQueryInsertGetIdByStruct(t, db)
}

func TestOracleIntegrationQueryInsertGetIdByMap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	common.TestQueryInsertGetIdByMap(t, db)
}
