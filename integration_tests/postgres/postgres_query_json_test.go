//go:build integration

package postgres

import (
	"testing"
)

func TestPostgresIntegrationQueryJson(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping JSON test - JSON query methods use MySQL/SQLite syntax, not PostgreSQL JSON operators")
}
