package tidb_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestTiDBIntegrationQueryChunkBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	if db == nil {
		t.Skip("TiDB not available")
	}
	_, _ = db.Query().Table("users").Where("name LIKE ?", "chunk_user_%").Delete()
	common.SeedChunkTestData(t, db)
	common.TestChunkBasic(t, db)
}

func TestTiDBIntegrationQueryChunkCustomBatchSize(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	if db == nil {
		t.Skip("TiDB not available")
	}
	_, _ = db.Query().Table("users").Where("name LIKE ?", "chunk_user_%").Delete()
	common.SeedChunkTestData(t, db)
	common.TestChunkCustomBatchSize(t, db)
}

func TestTiDBIntegrationQueryChunkErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	if db == nil {
		t.Skip("TiDB not available")
	}
	_, _ = db.Query().Table("users").Where("name LIKE ?", "chunk_user_%").Delete()
	common.SeedChunkTestData(t, db)
	common.TestChunkErrorHandling(t, db)
}
