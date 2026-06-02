package oracle_test

import (
	"testing"
)

func TestOracleSchemaTimestamp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("TODO: Oracle schema builder has case sensitivity issues")
}
