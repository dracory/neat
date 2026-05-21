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

func TestToSqlCreate(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	model := struct {
		Name string
	}{Name: "John"}

	toSql := NewToSql(q)
	sql := toSql.Create(model)

	if sql == "" {
		t.Error("Expected SQL to be generated for Create")
	}
}

func TestToSqlDelete(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	toSql := NewToSql(q)
	sql := toSql.Delete()

	if sql == "" {
		t.Error("Expected SQL to be generated for Delete")
	}
}

func TestToSqlFirst(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	toSql := NewToSql(q)
	sql := toSql.First(nil)

	if sql == "" {
		t.Error("Expected SQL to be generated for First")
	}
}

func TestToSqlGet(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	toSql := NewToSql(q)
	sql := toSql.Get(nil)

	if sql == "" {
		t.Error("Expected SQL to be generated for Get")
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

func TestToSqlUpdate(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	toSql := NewToSql(q)
	sql := toSql.Update("name", "John")

	if sql == "" {
		t.Error("Expected SQL to be generated for Update")
	}
}

func TestToSqlIncrement(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	toSql := NewToSql(q)
	sql := toSql.Increment("age")

	if sql == "" {
		t.Error("Expected SQL to be generated for Increment")
	}
}

func TestToSqlDecrement(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	toSql := NewToSql(q)
	sql := toSql.Decrement("age")

	if sql == "" {
		t.Error("Expected SQL to be generated for Decrement")
	}
}

func TestQueryToSqlMethod(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	toSql := q.ToSql()
	if toSql == nil {
		t.Error("Expected ToSql instance from Query.ToSql()")
	}
}

func TestQueryToRawSqlMethod(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	toSql := q.ToRawSql()
	if toSql == nil {
		t.Error("Expected ToSql instance from Query.ToRawSql()")
	}
}
