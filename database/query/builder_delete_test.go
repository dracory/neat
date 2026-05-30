package query

import (
	"context"
	"strings"
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

func TestBuildDeleteWithLimitMySQL(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.driver = &FakeDriver{DialectName: "mysql"}
	limit := 10
	q.limit = &limit
	q.Where("id > ?", 5)
	b := NewBuilder(q)

	sql, args := b.BuildDelete()

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	if !strings.Contains(sql, "LIMIT 10") {
		t.Error("Expected LIMIT clause in SQL for MySQL")
	}
	if !strings.Contains(sql, "WHERE") {
		t.Error("Expected WHERE clause in SQL")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildDeleteWithLimitSQLite(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.driver = &FakeDriver{DialectName: "sqlite"}
	limit := 10
	q.limit = &limit
	q.Where("id > ?", 5)
	b := NewBuilder(q)

	sql, args := b.BuildDelete()

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	// SQLite should use rowid subquery workaround
	if !strings.Contains(sql, "rowid IN") {
		t.Error("Expected rowid subquery workaround for SQLite")
	}
	if !strings.Contains(sql, "SELECT rowid FROM") {
		t.Error("Expected SELECT rowid subquery for SQLite")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildDeleteWithLimitSQLiteWithOrder(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.driver = &FakeDriver{DialectName: "sqlite"}
	limit := 10
	q.limit = &limit
	q.Where("id > ?", 5)
	q.OrderBy("id", "desc")
	b := NewBuilder(q)

	sql, args := b.BuildDelete()

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	if !strings.Contains(sql, "rowid IN") {
		t.Error("Expected rowid subquery workaround for SQLite")
	}
	if !strings.Contains(sql, "ORDER BY") {
		t.Error("Expected ORDER BY in subquery for SQLite")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildDeleteWithLimitPostgreSQL(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.driver = &FakeDriver{DialectName: "postgres"}
	limit := 10
	q.limit = &limit
	q.Where("id > ?", 5)
	b := NewBuilder(q)

	sql, args := b.BuildDelete()

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	if !strings.Contains(sql, "LIMIT 10") {
		t.Error("Expected LIMIT clause in SQL for PostgreSQL")
	}
	if !strings.Contains(sql, "WHERE") {
		t.Error("Expected WHERE clause in SQL")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildDeleteWithLimitNoWhere(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.driver = &FakeDriver{DialectName: "mysql"}
	limit := 10
	q.limit = &limit
	b := NewBuilder(q)

	sql, args := b.BuildDelete()

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	if !strings.Contains(sql, "LIMIT 10") {
		t.Error("Expected LIMIT clause in SQL")
	}
	// Args can be nil when there's no WHERE clause
	_ = args
}

func TestBuildDeleteWithLimitSQLServer(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	q.driver = &FakeDriver{DialectName: "sqlserver"}
	limit := 10
	q.limit = &limit
	q.Where("id > ?", 5)
	b := NewBuilder(q)

	sql, args := b.BuildDelete()

	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	// SQL Server uses TOP instead of LIMIT
	if !strings.Contains(sql, "DELETE TOP (10)") {
		t.Error("Expected TOP clause in SQL for SQL Server")
	}
	if !strings.Contains(sql, "WHERE") {
		t.Error("Expected WHERE clause in SQL")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}
