package oracle_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/common"
)

func TestOracleIntegrationJoinInner(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	if db == nil {
		t.Skip("Oracle not available")
	}
	common.TestJoinInner(t, db)
}

func TestOracleIntegrationJoinInnerWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	if db == nil {
		t.Skip("Oracle not available")
	}
	common.TestJoinInnerWithConditions(t, db)
}

func TestOracleIntegrationJoinInnerWithAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("TODO: Oracle SQL syntax differs for join aliases - ORA-00933 error")
}

func TestOracleIntegrationJoinLeft(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	if db == nil {
		t.Skip("Oracle not available")
	}
	common.TestJoinLeft(t, db)
}

func TestOracleIntegrationJoinLeftWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	if db == nil {
		t.Skip("Oracle not available")
	}
	common.TestJoinLeftWithConditions(t, db)
}

func TestOracleIntegrationJoinLeftWithAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("TODO: Oracle SQL syntax differs for join aliases - ORA-00933 error")
}

func TestOracleIntegrationJoinRight(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	if db == nil {
		t.Skip("Oracle not available")
	}
	common.TestJoinRight(t, db)
}

func TestOracleIntegrationJoinRightWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	if db == nil {
		t.Skip("Oracle not available")
	}
	common.TestJoinRightWithConditions(t, db)
}

func TestOracleIntegrationJoinRightWithAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("TODO: Oracle SQL syntax differs for join aliases - ORA-00933 error")
}

func TestOracleIntegrationJoinCross(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	if db == nil {
		t.Skip("Oracle not available")
	}
	common.TestJoinCross(t, db)
}

func TestOracleIntegrationJoinCrossWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	if db == nil {
		t.Skip("Oracle not available")
	}
	common.TestJoinCrossWithConditions(t, db)
}

func TestOracleIntegrationJoinCrossWithSelect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)
	if db == nil {
		t.Skip("Oracle not available")
	}
	common.TestJoinCrossWithSelect(t, db)
}
