package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestSQLiteIntegrationQueryScopesLocalWithoutParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	common.SeedScopesTestData(t, db)
	common.TestScopesWithoutParameters(t, db)
}

func TestSQLiteIntegrationQueryScopesLocalWithParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	common.SeedScopesTestData(t, db)
	common.TestScopesWithParameters(t, db)
}

func TestSQLiteIntegrationQueryScopesMultipleChaining(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	common.SeedScopesTestData(t, db)
	common.TestScopesMultipleChaining(t, db)
}
