package mysql_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestMysqlIntegrationQueryValueBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	common.TestValueBasic(t, db)
}

func TestMysqlIntegrationQueryValueWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	common.TestValueWithWhere(t, db)
}

func TestMysqlIntegrationQueryToSqlValue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	common.TestValueToSql(t, db)
}
