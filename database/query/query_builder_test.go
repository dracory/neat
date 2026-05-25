package query

import (
	"context"
	"testing"
)

func TestSelect(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.Select("name", "age")

	if result == nil {
		t.Error("Expected non-nil Query from Select")
	}
}

func TestDistinct(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.Distinct()

	if result == nil {
		t.Error("Expected non-nil Query from Distinct")
	}
}

func TestJoin(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.Join("posts", "users.id = posts.user_id")

	if result == nil {
		t.Error("Expected non-nil Query from Join")
	}
}

func TestLeftJoin(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.LeftJoin("posts", "users.id = posts.user_id")

	if result == nil {
		t.Error("Expected non-nil Query from LeftJoin")
	}
}

func TestRightJoin(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.RightJoin("posts", "users.id = posts.user_id")

	if result == nil {
		t.Error("Expected non-nil Query from RightJoin")
	}
}

func TestGroup(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.Group("name")

	if result == nil {
		t.Error("Expected non-nil Query from Group")
	}
}

func TestOrderBy(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrderBy("name")

	if result == nil {
		t.Error("Expected non-nil Query from OrderBy")
	}
}

func TestOrderByDesc(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrderByDesc("name")

	if result == nil {
		t.Error("Expected non-nil Query from OrderByDesc")
	}
}

func TestLimit(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.Limit(10)

	if result == nil {
		t.Error("Expected non-nil Query from Limit")
	}
}

func TestOffset(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.Offset(5)

	if result == nil {
		t.Error("Expected non-nil Query from Offset")
	}
}

func TestHaving(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.Having("count > ?", 5)

	if result == nil {
		t.Error("Expected non-nil Query from Having")
	}
}

func TestWith(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.With("Posts")

	if result == nil {
		t.Error("Expected non-nil Query from With")
	}
}

func TestOmit(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.Omit("password")

	if result == nil {
		t.Error("Expected non-nil Query from Omit")
	}
}
