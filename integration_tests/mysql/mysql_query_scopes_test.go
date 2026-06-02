package mysql_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestMySQLIntegrationQueryScopesWithoutParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	common.SeedScopesTestData(t, db)
	common.TestScopesWithoutParameters(t, db)
}

func TestMySQLIntegrationQueryScopesWithParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	common.SeedScopesTestData(t, db)
	common.TestScopesWithParameters(t, db)
}

func TestMySQLIntegrationQueryScopesMultipleChaining(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	common.SeedScopesTestData(t, db)
	common.TestScopesMultipleChaining(t, db)
}
