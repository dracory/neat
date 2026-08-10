//go:build integration

package cockroachdb_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestCockroachDBIntegrationScopesWithoutParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	common.SeedScopesTestData(t, db)
	common.TestScopesWithoutParameters(t, db)
}

func TestCockroachDBIntegrationScopesWithParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	common.SeedScopesTestData(t, db)
	common.TestScopesWithParameters(t, db)
}

func TestCockroachDBIntegrationScopesMultipleChaining(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	common.SeedScopesTestData(t, db)
	common.TestScopesMultipleChaining(t, db)
}
