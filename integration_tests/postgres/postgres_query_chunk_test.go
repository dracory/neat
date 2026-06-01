package postgres

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestPostgresIntegrationQueryChunkBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.SeedChunkTestData(t, db)
	common.TestChunkBasic(t, db)
}

func TestPostgresIntegrationQueryChunkCustomBatchSize(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.SeedChunkTestData(t, db)
	common.TestChunkCustomBatchSize(t, db)
}

func TestPostgresIntegrationQueryChunkErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	common.SeedChunkTestData(t, db)
	common.TestChunkErrorHandling(t, db)
}
