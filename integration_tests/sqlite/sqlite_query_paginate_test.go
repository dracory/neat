package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestSQLiteIntegrationPaginateFirstPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	common.TestPaginateFirstPage(t, db)
}

func TestSQLiteIntegrationPaginateSecondPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	common.TestPaginateSecondPage(t, db)
}

func TestSQLiteIntegrationPaginateLastPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	common.TestPaginateLastPage(t, db)
}

func TestSQLiteIntegrationPaginateWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	common.TestPaginateWithConditions(t, db)
}

func TestSQLiteIntegrationPaginatePageBeyondBounds(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	common.TestPaginatePageBeyondBounds(t, db)
}

func TestSQLiteIntegrationPaginateEmptyResults(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	common.TestPaginateEmptyResults(t, db)
}

func TestSQLiteIntegrationPaginateWithSelectAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	common.TestPaginateWithSelectAliases(t, db)
}

func TestSQLiteIntegrationCountWithSelectAlias(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	common.TestCountWithSelectAlias(t, db)
}
