package cockroachdb_test

import (
	"testing"
)

func TestCockroachDBIntegrationArray(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping array test - custom CockroachDB array types interfere with standard test table setup")
}
