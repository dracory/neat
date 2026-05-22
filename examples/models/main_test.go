package main_test

import (
	"testing"

	mainpkg "github.com/dracory/neat/examples/models"
)

func TestRunExample(t *testing.T) {
	// Use in-memory SQLite for testing
	err := mainpkg.RunExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunExample failed: %v", err)
	}
}
