package postgres

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestPostgresIntegrationQueryScopesLocalWithoutParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.SeedScopesTestData(t, db)
	common.TestScopesWithoutParameters(t, db)
}

func TestPostgresIntegrationQueryScopesLocalWithParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.SeedScopesTestData(t, db)
	common.TestScopesWithParameters(t, db)
}

func TestPostgresIntegrationQueryScopesMultipleChaining(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.SeedScopesTestData(t, db)
	common.TestScopesMultipleChaining(t, db)
}
