package main_test

import (
	"testing"

	mainpkg "github.com/dracory/neat/examples/migrator-transactions"
)

func TestRunTransactionExample(t *testing.T) {
	// Use in-memory SQLite for testing
	err := mainpkg.RunTransactionExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunTransactionExample failed: %v", err)
	}
}
