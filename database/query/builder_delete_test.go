package query

import (
	"testing"
)

func TestBuildDelete(t *testing.T) {
	q := NewQuery(nil, nil, nil, "users", nil, nil)
	b := NewBuilder(q)

	sql, args := b.BuildDelete()

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	_ = args // Just ensure it doesn't panic
}
