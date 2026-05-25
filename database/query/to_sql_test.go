package query

import (
	"context"
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
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

func TestSelectWithSubqueryCallback(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	// Test callback as first parameter with alias
	q.Select(func(q contractsorm.Query) contractsorm.Query {
		return q.Table("users").Select("name").Where("id = ?", 1)
	}, "sub_name")

	sql, args := NewBuilder(q).BuildSelect()
	if sql == "" {
		t.Error("Expected SQL to be generated for Select with subquery callback")
	}
	if !contains(sql, "SELECT") {
		t.Error("Expected SELECT in generated SQL")
	}
	if !contains(sql, "sub_name") {
		t.Error("Expected alias 'sub_name' in generated SQL")
	}
	if len(args) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(args))
	}
}

func TestSelectWithSubqueryCallbackNoAlias(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	// Test callback as first parameter without alias
	q.Select(func(q contractsorm.Query) contractsorm.Query {
		return q.Table("users").Select("name").Where("id = ?", 1)
	})

	sql, args := NewBuilder(q).BuildSelect()
	if sql == "" {
		t.Error("Expected SQL to be generated for Select with subquery callback")
	}
	if !contains(sql, "SELECT") {
		t.Error("Expected SELECT in generated SQL")
	}
	if len(args) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(args))
	}
}

func TestSelectWithSubqueryCallbackInArgs(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	// Test callback in args parameter
	q.Select("? as sub_name", func(q contractsorm.Query) contractsorm.Query {
		return q.Table("users").Select("name").Where("id = ?", 1)
	})

	sql, args := NewBuilder(q).BuildSelect()
	if sql == "" {
		t.Error("Expected SQL to be generated for Select with subquery callback in args")
	}
	if !contains(sql, "SELECT") {
		t.Error("Expected SELECT in generated SQL")
	}
	if !contains(sql, "sub_name") {
		t.Error("Expected alias 'sub_name' in generated SQL")
	}
	if len(args) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(args))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
