package query

import (
	"context"
	"testing"

	"github.com/dracory/neat/contracts/database/orm"
)

func TestWhereIn(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereIn("id", []any{1, 2, 3})

	if result == nil {
		t.Error("Expected non-nil Query from WhereIn")
	}
}

func TestOrWhereIn(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrWhereIn("id", []any{1, 2, 3})

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereIn")
	}
}

func TestWhereNotIn(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereNotIn("id", []any{1, 2, 3})

	if result == nil {
		t.Error("Expected non-nil Query from WhereNotIn")
	}
}

func TestOrWhereNotIn(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrWhereNotIn("id", []any{1, 2, 3})

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereNotIn")
	}
}

func TestWhereBetween(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereBetween("age", 18, 65)

	if result == nil {
		t.Error("Expected non-nil Query from WhereBetween")
	}
}

func TestOrWhereBetween(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrWhereBetween("age", 18, 65)

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereBetween")
	}
}

func TestWhereNotBetween(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereNotBetween("age", 18, 65)

	if result == nil {
		t.Error("Expected non-nil Query from WhereNotBetween")
	}
}

func TestOrWhereNotBetween(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrWhereNotBetween("age", 18, 65)

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereNotBetween")
	}
}

func TestWhereNull(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereNull("deleted_at")

	if result == nil {
		t.Error("Expected non-nil Query from WhereNull")
	}
}

func TestOrWhereNull(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrWhereNull("deleted_at")

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereNull")
	}
}

func TestWhereNotNull(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereNotNull("deleted_at")

	if result == nil {
		t.Error("Expected non-nil Query from WhereNotNull")
	}
}

func TestWhereColumn(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereColumn("updated_at", "=", "created_at")

	if result == nil {
		t.Error("Expected non-nil Query from WhereColumn")
	}
}

func TestOrWhereColumn(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrWhereColumn("updated_at", "=", "created_at")

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereColumn")
	}
}

func TestWhereExists(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereExists(func(q orm.Query) orm.Query {
		return q
	})

	if result == nil {
		t.Error("Expected non-nil Query from WhereExists")
	}
}

func TestWhereNot(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereNot("id = ?", 1)

	if result == nil {
		t.Error("Expected non-nil Query from WhereNot")
	}
}

func TestOrWhereNot(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrWhereNot("id = ?", 1)

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereNot")
	}
}

func TestWhereAny(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereAny([]string{"name", "email"}, "=", "John")

	if result == nil {
		t.Error("Expected non-nil Query from WhereAny")
	}
}

func TestWhereAll(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereAll([]string{"name", "email"}, "=", "John")

	if result == nil {
		t.Error("Expected non-nil Query from WhereAll")
	}
}

func TestWhereNone(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereNone([]string{"name", "email"}, "=", "John")

	if result == nil {
		t.Error("Expected non-nil Query from WhereNone")
	}
}

func TestWhereJsonContains(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereJsonContains("data", "value")

	if result == nil {
		t.Error("Expected non-nil Query from WhereJsonContains")
	}
}

func TestOrWhereJsonContains(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrWhereJsonContains("data", "value")

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereJsonContains")
	}
}

func TestWhereJsonDoesntContain(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereJsonDoesntContain("data", "value")

	if result == nil {
		t.Error("Expected non-nil Query from WhereJsonDoesntContain")
	}
}

func TestOrWhereJsonDoesntContain(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrWhereJsonDoesntContain("data", "value")

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereJsonDoesntContain")
	}
}

func TestWhereJsonContainsKey(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereJsonContainsKey("data->key")

	if result == nil {
		t.Error("Expected non-nil Query from WhereJsonContainsKey")
	}
}

func TestOrWhereJsonContainsKey(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrWhereJsonContainsKey("data->key")

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereJsonContainsKey")
	}
}

func TestWhereJsonDoesntContainKey(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereJsonDoesntContainKey("data->key")

	if result == nil {
		t.Error("Expected non-nil Query from WhereJsonDoesntContainKey")
	}
}

func TestOrWhereJsonDoesntContainKey(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrWhereJsonDoesntContainKey("data->key")

	if result == nil {
		t.Error("Expected non-nil Query from OrWhereJsonDoesntContainKey")
	}
}

func TestWhereJsonLength(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.WhereJsonLength("data", ">", 0)

	if result == nil {
		t.Error("Expected non-nil Query from WhereJsonLength")
	}
}
