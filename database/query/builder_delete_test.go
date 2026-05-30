package query

import (
	"context"
	"testing"
)

func TestBuildDelete(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	b := NewBuilder(q)

	sql, args := b.BuildDelete()

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	_ = args // Just ensure it doesn't panic
}
