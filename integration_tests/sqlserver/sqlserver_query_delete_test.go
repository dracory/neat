package sqlserver_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

// TestSQLServerIntegrationQueryDeleteByModel verifies that Delete() via a model
// removes exactly one row and reports RowsAffected = 1.
func TestSQLServerIntegrationQueryDeleteByModel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestQueryDeleteByModel(t, db)
}

// TestSQLServerIntegrationQueryDeleteByTable verifies that Delete() via a raw
// table name with a WHERE clause removes exactly one matching row.
func TestSQLServerIntegrationQueryDeleteByTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestQueryDeleteByTable(t, db)
}

// TestSQLServerIntegrationQueryDeleteByModelWithWhere verifies that a targeted
// DELETE only removes the row matching the WHERE clause, leaving other rows intact.
func TestSQLServerIntegrationQueryDeleteByModelWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestQueryDeleteByModelWithWhere(t, db)
}
