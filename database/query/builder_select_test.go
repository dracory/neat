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

func TestBuildSelectAggregateSkipsOrderBy(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.aggregate = "COUNT"
	q.aggregateCol = "*"
	q.orders = []orderClause{{column: "name", direction: "asc"}}
	b := NewBuilder(q)

	sql, _ := b.BuildSelect()
	if strings.Contains(sql, "ORDER BY") {
		t.Errorf("Expected no ORDER BY in aggregate query, got: %s", sql)
	}
	if !strings.Contains(sql, "COUNT(*)") {
		t.Errorf("Expected COUNT(*) in aggregate query, got: %s", sql)
	}
}

func TestBuildSelectAggregateSkipsLimit(t *testing.T) {
	limit := 10
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.aggregate = "COUNT"
	q.aggregateCol = "*"
	q.limit = &limit
	b := NewBuilder(q)

	sql, _ := b.BuildSelect()
	if strings.Contains(sql, "LIMIT") {
		t.Errorf("Expected no LIMIT in aggregate query, got: %s", sql)
	}
}

func TestBuildSelectAggregateSkipsOffset(t *testing.T) {
	offset := 5
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.aggregate = "COUNT"
	q.aggregateCol = "*"
	q.offset = &offset
	b := NewBuilder(q)

	sql, _ := b.BuildSelect()
	if strings.Contains(sql, "OFFSET") {
		t.Errorf("Expected no OFFSET in aggregate query, got: %s", sql)
	}
}

func TestBuildSelectAggregateSkipsLockForUpdate(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.aggregate = "COUNT"
	q.aggregateCol = "*"
	q.lockForUpdate = true
	b := NewBuilder(q)

	sql, _ := b.BuildSelect()
	if strings.Contains(sql, "FOR UPDATE") {
		t.Errorf("Expected no FOR UPDATE in aggregate query, got: %s", sql)
	}
}

func TestBuildSelectAggregateSkipsSharedLock(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.aggregate = "SUM"
	q.aggregateCol = "amount"
	q.sharedLock = true
	b := NewBuilder(q)

	sql, _ := b.BuildSelect()
	if strings.Contains(sql, "FOR SHARE") || strings.Contains(sql, "LOCK IN SHARE MODE") {
		t.Errorf("Expected no shared lock clause in aggregate query, got: %s", sql)
	}
}

func TestBuildSelectNonAggregateIncludesOrderBy(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.orders = []orderClause{{column: "name", direction: "asc"}}
	b := NewBuilder(q)

	sql, _ := b.BuildSelect()
	if !strings.Contains(sql, "ORDER BY") {
		t.Errorf("Expected ORDER BY in non-aggregate query, got: %s", sql)
	}
}

func TestBuildSelectNonAggregateIncludesLimit(t *testing.T) {
	limit := 20
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.limit = &limit
	b := NewBuilder(q)

	sql, _ := b.BuildSelect()
	if !strings.Contains(sql, "LIMIT 20") {
		t.Errorf("Expected LIMIT 20 in non-aggregate query, got: %s", sql)
	}
}

func TestBuildSelectSubqueryFromPlaceholders(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "(SELECT id FROM orders WHERE amount > ?) sub"
	q.tableArgs = []any{100}
	b := NewBuilder(q)

	sql, args := b.BuildSelect()
	if !strings.Contains(sql, "FROM") {
		t.Errorf("Expected FROM clause in subquery query, got: %s", sql)
	}
	if !strings.Contains(sql, "sub") {
		t.Errorf("Expected subquery alias in SQL, got: %s", sql)
	}
	if len(args) != 1 || args[0] != 100 {
		t.Errorf("Expected args [100], got %v", args)
	}
}

func TestBuildSelectDialectPlaceholderInSelect(t *testing.T) {
	q := NewQuery(context.TODO(), nil, &FakeDriver{DialectName: "postgres"}, "", nil, nil)
	q.table = "users"
	q.selects = []selectClause{{expr: "COALESCE(score, ?)", args: []any{0}}}
	b := NewBuilder(q)

	sql, args := b.BuildSelect()
	if !strings.Contains(sql, "$1") {
		t.Errorf("Expected postgres placeholder $1 in SELECT, got: %s", sql)
	}
	if len(args) != 1 || args[0] != 0 {
		t.Errorf("Expected args [0], got %v", args)
	}
}

func TestBuildSelectWithLimitSQLServer(t *testing.T) {
	limit := 10
	q := NewQuery(context.TODO(), nil, &FakeDriver{DialectName: "sqlserver"}, "users", nil, nil)
	q.limit = &limit
	b := NewBuilder(q)

	sql, _ := b.BuildSelect()
	if !strings.Contains(sql, "SELECT TOP 10") {
		t.Error("Expected TOP 10 in SQL for SQL Server")
	}
}
