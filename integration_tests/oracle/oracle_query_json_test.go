package oracle_test

import (
	"testing"
)

func TestOracleIntegrationQueryJson(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Oracle JSON query methods not yet implemented - Oracle uses JSON_VALUE, JSON_EXISTS, JSON_TABLE functions instead of -> operator syntax")
}
