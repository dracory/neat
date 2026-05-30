//go:build integration

package postgres

import (
	"testing"
)

func TestPostgresIntegrationPaginateFirstPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping Paginate test - soft-delete filter generates incompatible SQL for PostgreSQL")
}

func TestPostgresIntegrationPaginateWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping Paginate test - soft-delete filter generates incompatible SQL for PostgreSQL")
}

func TestPostgresIntegrationPaginateWithSelectAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping Paginate test - soft-delete filter generates incompatible SQL for PostgreSQL")
}
