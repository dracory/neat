package main_test

import (
	"testing"

	mainpkg "github.com/dracory/neat/examples/array-driver"
)

func TestRunExample(t *testing.T) {
	err := mainpkg.RunExample()
	if err != nil {
		t.Fatalf("RunExample failed: %v", err)
	}
}
