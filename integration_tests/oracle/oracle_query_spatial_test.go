package oracle_test

import (
	"testing"
)

func TestOracleIntegrationQuerySpatial(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Oracle spatial query methods not yet implemented - Oracle uses SDO_GEOMETRY instead of standard spatial types")
}
