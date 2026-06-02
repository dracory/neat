package sqlserver_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

// TestSQLServerIntegrationQueryCreateByStruct verifies that a struct can be
// inserted via Create() and that the resulting record is queryable by name,
// with a non-zero ID assigned by the database.
func TestSQLServerIntegrationQueryCreateByStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestQueryCreateByStruct(t, db)
}

// TestSQLServerIntegrationQueryBatchCreateByStruct verifies that a slice of
// structs can be inserted in a single Create() call and that all rows are
// subsequently retrievable.
func TestSQLServerIntegrationQueryBatchCreateByStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestQueryBatchCreateByStruct(t, db)
}

// TestSQLServerIntegrationQueryInsertGetIdByStruct verifies that InsertGetId()
// with a struct returns a non-zero ID and writes that ID back to the struct's
// ID field via the OUTPUT clause.
func TestSQLServerIntegrationQueryInsertGetIdByStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestQueryInsertGetIdByStruct(t, db)
}

// TestSQLServerIntegrationQueryInsertGetIdByMap verifies that InsertGetId() with
// a map returns a non-zero ID that can be used to retrieve the inserted row.
func TestSQLServerIntegrationQueryInsertGetIdByMap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestQueryInsertGetIdByMap(t, db)
}
