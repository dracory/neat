package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestSQLiteIntegrationQueryCountBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	common.QueryCountBasic(t, db)
}

func TestSQLiteIntegrationQueryCountWithTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	common.QueryCountWithTable(t, db)
}
