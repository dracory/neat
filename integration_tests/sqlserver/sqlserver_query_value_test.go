package sqlserver

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestSQLServerIntegrationQueryValueBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestValueBasic(t, db)
}

func TestSQLServerIntegrationQueryValueWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestValueWithWhere(t, db)
}

func TestSQLServerIntegrationQueryToSqlValue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	common.TestValueToSql(t, db)
}
