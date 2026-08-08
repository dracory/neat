package main_test

import (
	"testing"

	mainpkg "github.com/dracory/neat/examples/array-driver-dotted-columns"
)

func TestRunDottedColumnExample(t *testing.T) {
	err := mainpkg.RunDottedColumnExample()
	if err != nil {
		t.Fatalf("RunDottedColumnExample failed: %v", err)
	}
}
