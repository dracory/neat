package main_test

import (
	"testing"

	mainpkg "github.com/dracory/neat/examples/array-driver-advanced"
)

func TestRunAdvancedExample(t *testing.T) {
	err := mainpkg.RunAdvancedExample()
	if err != nil {
		t.Fatalf("RunAdvancedExample failed: %v", err)
	}
}
