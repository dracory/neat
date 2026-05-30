package query

import (
	"context"
	"strings"
	"testing"
)

func TestBuildSelect(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	b := NewBuilder(q)

	sql, args := b.BuildSelect()
	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildSelectWithRawSQL(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.rawSQL = "SELECT * FROM users WHERE id = ?"
	q.rawArgs = []any{1}
	b := NewBuilder(q)

	sql, args := b.BuildSelect()
	if sql != "SELECT * FROM users WHERE id = ?" {
		t.Errorf("Expected raw SQL, got %q", sql)
	}
	if len(args) != 1 || args[0] != 1 {
		t.Errorf("Expected args [1], got %v", args)
	}
}

func TestBuildSelectWithDistinct(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.distinct = true
	b := NewBuilder(q)

	sql, _ := b.BuildSelect()
	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
}

func TestBuildSelectWithLimit(t *testing.T) {
	limit := 10
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.limit = &limit
	b := NewBuilder(q)

	sql, _ := b.BuildSelect()
	if !strings.Contains(sql, "LIMIT 10") {
		t.Error("Expected LIMIT 10 in SQL")
	}
}

func TestBuildSelectWithOffset(t *testing.T) {
	offset := 5
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.offset = &offset
	b := NewBuilder(q)

	sql, _ := b.BuildSelect()
	if !strings.Contains(sql, "OFFSET 5") {
		t.Error("Expected OFFSET 5 in SQL")
	}
}

func TestBuildSelectWithOrderBy(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.orders = []orderClause{{column: "name", direction: "ASC"}}
	b := NewBuilder(q)

	sql, _ := b.BuildSelect()
	if !strings.Contains(sql, "ORDER BY") {
		t.Error("Expected ORDER BY in SQL")
	}
}

func TestBuildSelectWithGroupBy(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.groups = []string{"department"}
	b := NewBuilder(q)

	sql, _ := b.BuildSelect()
	if !strings.Contains(sql, "GROUP BY") {
		t.Error("Expected GROUP BY in SQL")
	}
}

func TestBuildSelectWithHaving(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.havings = []havingClause{{query: "COUNT(*) > ?", args: []any{5}}}
	b := NewBuilder(q)

	sql, _ := b.BuildSelect()
	if !strings.Contains(sql, "HAVING") {
		t.Error("Expected HAVING in SQL")
	}
}

func TestBuildSelectWithAggregate(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.aggregate = "COUNT"
	q.aggregateCol = "*"
	b := NewBuilder(q)

	sql, _ := b.BuildSelect()
	if !strings.Contains(sql, "COUNT") {
		t.Error("Expected COUNT in SQL")
	}
}

func TestBuildSelectWithJoin(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.joins = []joinClause{{_type: "LEFT JOIN", query: "profiles ON users.id = profiles.user_id"}}
	b := NewBuilder(q)

	sql, _ := b.BuildSelect()
	if !strings.Contains(sql, "LEFT JOIN") {
		t.Error("Expected LEFT JOIN in SQL")
	}
}
