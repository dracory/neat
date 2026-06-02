package postgres_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestPostgresIntegrationPaginateFirstPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestPaginateFirstPage(t, db)
}

func TestPostgresIntegrationPaginateSecondPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestPaginateSecondPage(t, db)
}

func TestPostgresIntegrationPaginateWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestPaginateWithConditions(t, db)
}

func TestPostgresIntegrationPaginateWithSelectAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestPaginateWithSelectAliases(t, db)
}
