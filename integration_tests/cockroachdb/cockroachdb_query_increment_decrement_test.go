//go:build integration

package cockroachdb_test

import (
	"testing"
)

func TestCockroachDBIntegrationIncrementID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Incrementing/decrementing an auto-increment ID is an invalid operation")
}

func TestCockroachDBIntegrationDecrementID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Decrementing/decrementing an auto-increment ID is an invalid operation")
}

func TestCockroachDBIntegrationIncrementNonExistentColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Increment feature not yet implemented for CockroachDB")
}

func TestCockroachDBIntegrationDecrementNonExistentColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Decrement feature not yet implemented for CockroachDB")
}
