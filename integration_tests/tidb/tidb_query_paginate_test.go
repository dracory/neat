//go:build integration

package tidb_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestTiDBIntegrationPaginateFirstPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestPaginateFirstPage(t, db)
}

func TestTiDBIntegrationPaginateSecondPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestPaginateSecondPage(t, db)
}

func TestTiDBIntegrationPaginateWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestPaginateWithConditions(t, db)
}

func TestTiDBIntegrationPaginateWithSelectAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestPaginateWithSelectAliases(t, db)
}
