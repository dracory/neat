package postgres

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestPostgreSQLIntegrationQueryCountBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.QueryCountBasic(t, db)
}
