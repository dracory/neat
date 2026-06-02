package oracle_test

import (
	"testing"
)

func TestOracleLockForUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("TODO: Oracle ORA-02014 - cannot select FOR UPDATE from view with DISTINCT, GROUP BY, etc.")
}

func TestOracleSharedLock(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("TODO: Oracle ORA-02014 - cannot select FOR UPDATE from view with DISTINCT, GROUP BY, etc.")
}

func TestOracleConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("TODO: Oracle ORA-02014 - cannot select FOR UPDATE from view with DISTINCT, GROUP BY, etc.")
}
