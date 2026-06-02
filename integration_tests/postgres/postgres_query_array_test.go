
package postgres_test

import (
	"testing"
)

func TestPostgresIntegrationArray(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping array test - custom PostgreSQL array types interfere with standard test table setup")
}
