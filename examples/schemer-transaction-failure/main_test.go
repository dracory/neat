package main_test

import (
	"testing"

	mainpkg "github.com/dracory/neat/examples/schemer-transaction-failure"
)

func TestRunTransactionFailureExample(t *testing.T) {
	// Use in-memory SQLite for testing
	err := mainpkg.RunTransactionFailureExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunTransactionFailureExample failed: %v", err)
	}
}
