package oracle_test

import (
	"testing"
)

func TestOracleSchemaForeignKey(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
}
