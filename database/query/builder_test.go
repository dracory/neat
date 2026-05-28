package query

import (
	"testing"
)

func TestNewBuilder(t *testing.T) {
	q := NewQuery(nil, nil, nil, "", nil, nil)
	b := NewBuilder(q)
	if b == nil {
		t.Error("Expected non-nil Builder")
	}
	if b.query != q {
		t.Error("Expected Builder to have the provided query")
	}
}
