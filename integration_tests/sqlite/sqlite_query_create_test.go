package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestSQLiteIntegrationQueryCreateByStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	common.TestQueryCreateByStruct(t, db)
}

func TestSQLiteIntegrationQueryBatchCreateByStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	common.TestQueryBatchCreateByStruct(t, db)
}

func TestSQLiteIntegrationQueryCreateByMap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	common.TestQueryCreateByMap(t, db)
}

func TestSQLiteIntegrationQueryInsertGetIdByStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	common.TestQueryInsertGetIdByStruct(t, db)
}

func TestSQLiteIntegrationQueryInsertGetIdByMap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	common.TestQueryInsertGetIdByMap(t, db)
}
