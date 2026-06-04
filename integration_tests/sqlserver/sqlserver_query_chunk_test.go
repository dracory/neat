package sqlserver_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestSQLServerIntegrationQueryChunkBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	if db == nil {
		t.Skip("SQL Server not available")
	}
	_, _ = db.Query().Table("users").Where("name LIKE ?", "chunk_user_%").Delete()
	common.SeedChunkTestData(t, db)
	common.TestChunkBasic(t, db)
}

func TestSQLServerIntegrationQueryChunkCustomBatchSize(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	if db == nil {
		t.Skip("SQL Server not available")
	}
	_, _ = db.Query().Table("users").Where("name LIKE ?", "chunk_user_%").Delete()
	common.SeedChunkTestData(t, db)
	common.TestChunkCustomBatchSize(t, db)
}

func TestSQLServerIntegrationQueryChunkErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLServerTest(t)
	if db == nil {
		t.Skip("SQL Server not available")
	}
	_, _ = db.Query().Table("users").Where("name LIKE ?", "chunk_user_%").Delete()
	common.SeedChunkTestData(t, db)
	common.TestChunkErrorHandling(t, db)
}
