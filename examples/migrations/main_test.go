package main_test

import (
	"testing"

	mainpkg "github.com/dracory/neat/examples/migrations"
)

func TestRunSchemaBuilderExample(t *testing.T) {
	// Use in-memory SQLite for testing
	err := mainpkg.RunSchemaBuilderExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunSchemaBuilderExample failed: %v", err)
	}
}

func TestRunMigrationSystemExample(t *testing.T) {
	// Use in-memory SQLite for testing
	err := mainpkg.RunMigrationSystemExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunMigrationSystemExample failed: %v", err)
	}
}
