package mysql

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestMySQLIntegrationQueryCreateByStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	common.TestQueryCreateByStruct(t, db)
}

func TestMySQLIntegrationQueryBatchCreateByStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	common.TestQueryBatchCreateByStruct(t, db)
}

func TestMySQLIntegrationQueryInsertGetIdByStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	common.TestQueryInsertGetIdByStruct(t, db)
}

func TestMySQLIntegrationQueryInsertGetIdByMap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	common.TestQueryInsertGetIdByMap(t, db)
}
