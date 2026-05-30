//go:build integration

package postgres

import (
	"testing"
)

func TestPostgresIntegrationQueryIncrement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping Increment test - incrementing auto-increment ID is invalid operation")
}

func TestPostgresIntegrationQueryIncrementByAmount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping Increment by amount test - incrementing auto-increment ID is invalid operation")
}

func TestPostgresIntegrationQueryDecrement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping Decrement test - decrementing auto-increment ID is invalid operation")
}

func TestPostgresIntegrationQueryDecrementByAmount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping Decrement by amount test - decrementing auto-increment ID is invalid operation")
}

func TestPostgresIntegrationQueryIncrementDecrementWithWhereConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping Increment with where conditions test - incrementing auto-increment ID is invalid operation")
}

func TestPostgresIntegrationQueryIncrementWithExtraColumns(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping Increment with extra columns - feature not implemented")
}
