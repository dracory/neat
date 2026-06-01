package mysql

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestMySQLIntegrationQueryCountBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	common.QueryCountBasic(t, db)
}
