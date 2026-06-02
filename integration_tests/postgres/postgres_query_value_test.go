package postgres_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestPostgresIntegrationQueryValueBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestValueBasic(t, db)
}

func TestPostgresIntegrationQueryValueWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestValueWithWhere(t, db)
}

func TestPostgresIntegrationQueryToSqlValue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestValueToSql(t, db)
}
