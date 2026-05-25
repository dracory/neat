package query

import (
	"context"
	"testing"
)

func TestNewToSql(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	toSql := NewToSql(q)

	if toSql == nil {
		t.Error("Expected non-nil ToSql")
	}
	if toSql.query != q {
		t.Error("Expected ToSql to have the provided query")
	}
}

func TestToSqlUseValues(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	toSql := NewToSql(q)

	if toSql.useValues {
		t.Error("Expected useValues to be false by default")
	}

	toSqlWithValues := q.ToRawSql()
	if toSqlWithValues == nil {
		t.Error("Expected non-nil ToSql from ToRawSql")
	}
}

func TestReplacePlaceholders(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	toSql := NewToSql(q)

	sql := toSql.replacePlaceholders("SELECT * FROM users WHERE id = ?", []any{1})
	if sql != "SELECT * FROM users WHERE id = ?" {
		t.Errorf("Expected 'SELECT * FROM users WHERE id = ?', got %q", sql)
	}
}

func TestReplacePlaceholdersWithValues(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	toSql := NewToSql(q)

	sql := toSql.replacePlaceholdersWithValues("SELECT * FROM users WHERE id = ?", []any{1})
	if sql != "SELECT * FROM users WHERE id = 1" {
		t.Errorf("Expected 'SELECT * FROM users WHERE id = 1', got %q", sql)
	}
}

func TestReplacePlaceholdersWithValuesString(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	toSql := NewToSql(q)

	sql := toSql.replacePlaceholdersWithValues("SELECT * FROM users WHERE name = ?", []any{"John"})
	if sql != "SELECT * FROM users WHERE name = 'John'" {
		t.Errorf("Expected \"SELECT * FROM users WHERE name = 'John'\", got %q", sql)
	}
}

func TestReplacePlaceholdersWithValuesNil(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	toSql := NewToSql(q)

	sql := toSql.replacePlaceholdersWithValues("SELECT * FROM users WHERE name = ?", []any{nil})
	if sql != "SELECT * FROM users WHERE name = NULL" {
		t.Errorf("Expected 'SELECT * FROM users WHERE name = NULL', got %q", sql)
	}
}
