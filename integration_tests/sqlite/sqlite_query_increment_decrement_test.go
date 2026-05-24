package sqlite

import (
	"testing"
)

func TestSQLiteIntegrationQueryIncrement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("User model needs a numeric column for increment testing")
}
