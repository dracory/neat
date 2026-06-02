package postgres_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestPostgreSQLIntegrationQueryFindById(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestFindById(t, db)
}

func TestPostgreSQLIntegrationQueryFindWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestFindWithWhere(t, db)
}

func TestPostgreSQLIntegrationQueryFindWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.TestFindWithConditions(t, db)
}
