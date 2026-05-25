package query

import (
	"context"
	"testing"
)

func TestToSqlCount(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	toSql := NewToSql(q)
	sql := toSql.Count()

	if sql == "" {
		t.Error("Expected SQL to be generated for Count")
	}
}

func TestToSqlPluck(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	toSql := NewToSql(q)
	sql := toSql.Pluck("name", nil)

	if sql == "" {
		t.Error("Expected SQL to be generated for Pluck")
	}
}

func TestToSqlValue(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	toSql := NewToSql(q)
	sql := toSql.Value("name", nil)

	if sql == "" {
		t.Error("Expected SQL to be generated for Value")
	}
}

func TestToSqlAvg(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	toSql := NewToSql(q)
	sql := toSql.Avg("age", nil)

	if sql == "" {
		t.Error("Expected SQL to be generated for Avg")
	}
}

func TestToSqlMax(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	toSql := NewToSql(q)
	sql := toSql.Max("age", nil)

	if sql == "" {
		t.Error("Expected SQL to be generated for Max")
	}
}

func TestToSqlMin(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	toSql := NewToSql(q)
	sql := toSql.Min("age", nil)

	if sql == "" {
		t.Error("Expected SQL to be generated for Min")
	}
}

func TestToSqlSum(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	toSql := NewToSql(q)
	sql := toSql.Sum("age", nil)

	if sql == "" {
		t.Error("Expected SQL to be generated for Sum")
	}
}
