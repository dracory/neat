//go:build integration

package tidb_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestTiDBIntegrationJoinInner(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	if db == nil {
		t.Skip("TiDB not available")
	}
	common.TestJoinInner(t, db)
}

func TestTiDBIntegrationJoinInnerWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	if db == nil {
		t.Skip("TiDB not available")
	}
	common.TestJoinInnerWithConditions(t, db)
}

func TestTiDBIntegrationJoinInnerWithAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	if db == nil {
		t.Skip("TiDB not available")
	}
	common.TestJoinInnerWithAliases(t, db)
}

func TestTiDBIntegrationJoinLeft(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	if db == nil {
		t.Skip("TiDB not available")
	}
	common.TestJoinLeft(t, db)
}

func TestTiDBIntegrationJoinLeftWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	if db == nil {
		t.Skip("TiDB not available")
	}
	common.TestJoinLeftWithConditions(t, db)
}

func TestTiDBIntegrationJoinLeftWithAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	if db == nil {
		t.Skip("TiDB not available")
	}
	common.TestJoinLeftWithAliases(t, db)
}

func TestTiDBIntegrationJoinRight(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	if db == nil {
		t.Skip("TiDB not available")
	}
	common.TestJoinRight(t, db)
}

func TestTiDBIntegrationJoinRightWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	if db == nil {
		t.Skip("TiDB not available")
	}
	common.TestJoinRightWithConditions(t, db)
}

func TestTiDBIntegrationJoinRightWithAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	if db == nil {
		t.Skip("TiDB not available")
	}
	common.TestJoinRightWithAliases(t, db)
}

func TestTiDBIntegrationJoinCross(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	if db == nil {
		t.Skip("TiDB not available")
	}
	common.TestJoinCross(t, db)
}

func TestTiDBIntegrationJoinCrossWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	if db == nil {
		t.Skip("TiDB not available")
	}
	common.TestJoinCrossWithConditions(t, db)
}

func TestTiDBIntegrationJoinCrossWithSelect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTiDBTest(t)
	if db == nil {
		t.Skip("TiDB not available")
	}
	common.TestJoinCrossWithSelect(t, db)
}
