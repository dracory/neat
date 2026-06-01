package postgres

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestPostgreSQLIntegrationQueryDeleteByModel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestQueryDeleteByModel(t, db)
}

func TestPostgreSQLIntegrationQueryDeleteByTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestQueryDeleteByTable(t, db)
}

func TestPostgreSQLIntegrationQueryDeleteByModelWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestQueryDeleteByModelWithWhere(t, db)
}
