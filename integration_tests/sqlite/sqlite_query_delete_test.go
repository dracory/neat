package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestSQLiteIntegrationQueryDeleteByModel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	common.TestQueryDeleteByModel(t, db)
}

func TestSQLiteIntegrationQueryDeleteByTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	common.TestQueryDeleteByTable(t, db)
}

func TestSQLiteIntegrationQueryDeleteByModelWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	common.TestQueryDeleteByModelWithWhere(t, db)
}
