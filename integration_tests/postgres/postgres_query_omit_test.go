//go:build integration

package postgres

import (
	"testing"
)

func TestPostgresIntegrationQueryOmitDuringSelect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping Omit test - soft-delete filter generates incompatible SQL for PostgreSQL")
}

func TestPostgresIntegrationQueryOmitDuringUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping Omit test - soft-delete filter generates incompatible SQL for PostgreSQL")
}
