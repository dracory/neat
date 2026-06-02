package sqlserver_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

// TestSQLServerIntegrationQueryFindById verifies that a user can be found by its
// primary key ID using First() with a WHERE clause.
func TestSQLServerIntegrationQueryFindById(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestFindById(t, db)
}

// TestSQLServerIntegrationQueryFindWithWhere verifies that First() returns the
// correct user when filtered by an exact name match.
func TestSQLServerIntegrationQueryFindWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestFindWithWhere(t, db)
}

// TestSQLServerIntegrationQueryFindWithConditions verifies that Find() with a
// WHERE condition returns only the rows that satisfy the filter, not all rows.
func TestSQLServerIntegrationQueryFindWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestFindWithConditions(t, db)
}
