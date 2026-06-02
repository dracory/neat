package oracle_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestOracleIntegrationQueryChunkBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	if db == nil {
		t.Skip("Oracle not available")
	}
	_, _ = db.Query().Table("users").Where("name LIKE ?", "chunk_user_%").Delete()
	common.SeedChunkTestData(t, db)
	common.TestChunkBasic(t, db)
}

func TestOracleIntegrationQueryChunkCustomBatchSize(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	if db == nil {
		t.Skip("Oracle not available")
	}
	_, _ = db.Query().Table("users").Where("name LIKE ?", "chunk_user_%").Delete()
	common.SeedChunkTestData(t, db)
	common.TestChunkCustomBatchSize(t, db)
}

func TestOracleIntegrationQueryChunkErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	if db == nil {
		t.Skip("Oracle not available")
	}
	_, _ = db.Query().Table("users").Where("name LIKE ?", "chunk_user_%").Delete()
	common.SeedChunkTestData(t, db)
	common.TestChunkErrorHandling(t, db)
}
