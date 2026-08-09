package cockroachdb_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestCockroachDBIntegrationChunkBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	common.SeedChunkTestData(t, db)
	common.TestChunkBasic(t, db)
}

func TestCockroachDBIntegrationChunkCustomBatchSize(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	common.SeedChunkTestData(t, db)
	common.TestChunkCustomBatchSize(t, db)
}

func TestCockroachDBIntegrationChunkErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	common.SeedChunkTestData(t, db)
	common.TestChunkErrorHandling(t, db)
}
