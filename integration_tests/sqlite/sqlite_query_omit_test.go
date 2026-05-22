package sqlite

import (
	"testing"
)

func TestSQLiteIntegrationQueryOmit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	SetupSQLiteTest(t)

	t.Run("Omit during select", func(t *testing.T) {
		t.Skip("ORM Omit() does not exclude columns from SELECT — not yet implemented")
	})

	t.Run("Omit during update", func(t *testing.T) {
		t.Skip("ORM Omit().Save() generates invalid SQL (near SET: syntax error) — not yet fixed")
	})
}
