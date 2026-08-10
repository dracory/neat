//go:build integration

package tidb_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestTiDBIntegrationQueryFindById(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestFindById(t, db)
}

func TestTiDBIntegrationQueryFindWithWhere(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestFindWithWhere(t, db)
}

func TestTiDBIntegrationQueryFindWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	common.TestFindWithConditions(t, db)
}
