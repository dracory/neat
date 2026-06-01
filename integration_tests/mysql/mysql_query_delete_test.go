package mysql

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestMySQLIntegrationQueryDeleteByModel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	common.TestQueryDeleteByModel(t, db)
}

func TestMySQLIntegrationQueryDeleteByTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	common.TestQueryDeleteByTable(t, db)
}

func TestMySQLIntegrationQueryDeleteByModelWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	common.TestQueryDeleteByModelWithWhere(t, db)
}
