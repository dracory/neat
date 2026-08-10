//go:build integration

package cockroachdb_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestCockroachDBIntegrationPaginateFirstPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	common.TestPaginateFirstPage(t, db)
}

func TestCockroachDBIntegrationPaginateSecondPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	common.TestPaginateSecondPage(t, db)
}

func TestCockroachDBIntegrationPaginateWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	common.TestPaginateWithConditions(t, db)
}

func TestCockroachDBIntegrationPaginateWithSelectAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	common.TestPaginateWithSelectAliases(t, db)
}

func TestCockroachDBIntegrationPaginateLastPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	common.TestPaginateLastPage(t, db)
}

func TestCockroachDBIntegrationPaginatePageBeyondBounds(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	common.TestPaginatePageBeyondBounds(t, db)
}

func TestCockroachDBIntegrationPaginateEmptyResults(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	common.TestPaginateEmptyResults(t, db)
}

func TestCockroachDBIntegrationCountWithSelectAlias(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	common.TestCountWithSelectAlias(t, db)
}
