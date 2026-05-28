package main_test

import (
	"testing"

	mainpkg "github.com/dracory/neat/examples/factory"
)

func TestRunExample(t *testing.T) {
	err := mainpkg.RunExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunExample failed: %v", err)
	}
}
