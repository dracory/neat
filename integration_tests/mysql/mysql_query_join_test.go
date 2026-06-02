package mysql_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestMySQLIntegrationJoinInner(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	if db == nil {
		t.Skip("MySQL not available")
	}
	common.TestJoinInner(t, db)
}

func TestMySQLIntegrationJoinInnerWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	if db == nil {
		t.Skip("MySQL not available")
	}
	common.TestJoinInnerWithConditions(t, db)
}

func TestMySQLIntegrationJoinInnerWithAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	if db == nil {
		t.Skip("MySQL not available")
	}
	common.TestJoinInnerWithAliases(t, db)
}

func TestMySQLIntegrationJoinLeft(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	if db == nil {
		t.Skip("MySQL not available")
	}
	common.TestJoinLeft(t, db)
}

func TestMySQLIntegrationJoinLeftWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	if db == nil {
		t.Skip("MySQL not available")
	}
	common.TestJoinLeftWithConditions(t, db)
}

func TestMySQLIntegrationJoinLeftWithAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	if db == nil {
		t.Skip("MySQL not available")
	}
	common.TestJoinLeftWithAliases(t, db)
}

func TestMySQLIntegrationJoinRight(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	if db == nil {
		t.Skip("MySQL not available")
	}
	common.TestJoinRight(t, db)
}

func TestMySQLIntegrationJoinRightWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	if db == nil {
		t.Skip("MySQL not available")
	}
	common.TestJoinRightWithConditions(t, db)
}

func TestMySQLIntegrationJoinRightWithAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	if db == nil {
		t.Skip("MySQL not available")
	}
	common.TestJoinRightWithAliases(t, db)
}

func TestMySQLIntegrationJoinCross(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	if db == nil {
		t.Skip("MySQL not available")
	}
	common.TestJoinCross(t, db)
}

func TestMySQLIntegrationJoinCrossWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	if db == nil {
		t.Skip("MySQL not available")
	}
	common.TestJoinCrossWithConditions(t, db)
}

func TestMySQLIntegrationJoinCrossWithSelect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	if db == nil {
		t.Skip("MySQL not available")
	}
	common.TestJoinCrossWithSelect(t, db)
}
