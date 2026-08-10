package main

import "testing"

func TestRunExample(t *testing.T) {
	if err := RunExample(); err != nil {
		t.Fatalf("RunExample failed: %v", err)
	}
}
