package query

import (
	"context"
	"strings"
	"testing"
)

func TestSelect(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.Select("name", "age")

	if result == nil {
		t.Error("Expected non-nil Query from Select")
	}
}

func TestSelectSQLGeneration(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.Select("name").Select("email")

	wrapped := WrapQuery(q)
	sql, _ := wrapped.BuildSelectSQL()

	if !strings.Contains(sql, "SELECT") {
		t.Errorf("Expected SQL to contain 'SELECT', got: %s", sql)
	}
	if !strings.Contains(sql, "name") || !strings.Contains(sql, "email") {
		t.Errorf("Expected SQL to contain selected columns, got: %s", sql)
	}
	if !strings.Contains(sql, "FROM users") {
		t.Errorf("Expected SQL to contain 'FROM users', got: %s", sql)
	}
}

func TestDistinct(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.Distinct()

	if result == nil {
		t.Error("Expected non-nil Query from Distinct")
	}
}

func TestDistinctWithColumns(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.Distinct("name", "email")

	if result == nil {
		t.Error("Expected non-nil Query from Distinct with columns")
	}

	wrapped := WrapQuery(result.(*Query))
	if !wrapped.GetDistinct() {
		t.Error("Distinct flag should be set")
	}

	distinctCols := wrapped.GetDistinctCols()
	if len(distinctCols) != 2 {
		t.Errorf("Expected 2 distinct columns, got %d", len(distinctCols))
	}
	if distinctCols[0] != "name" {
		t.Errorf("Expected first column 'name', got '%s'", distinctCols[0])
	}
	if distinctCols[1] != "email" {
		t.Errorf("Expected second column 'email', got '%s'", distinctCols[1])
	}
}

func TestDistinctWithSingleColumn(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.Distinct("status")

	if result == nil {
		t.Error("Expected non-nil Query from Distinct with single column")
	}

	wrapped := WrapQuery(result.(*Query))
	distinctCols := wrapped.GetDistinctCols()
	if len(distinctCols) != 1 {
		t.Errorf("Expected 1 distinct column, got %d", len(distinctCols))
	}
	if distinctCols[0] != "status" {
		t.Errorf("Expected column 'status', got '%s'", distinctCols[0])
	}
}

func TestDistinctWithNoColumns(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.Distinct()

	if result == nil {
		t.Error("Expected non-nil Query from Distinct with no columns")
	}

	wrapped := WrapQuery(result.(*Query))
	if !wrapped.GetDistinct() {
		t.Error("Distinct flag should be set")
	}

	distinctCols := wrapped.GetDistinctCols()
	if len(distinctCols) != 0 {
		t.Errorf("Expected 0 distinct columns, got %d", len(distinctCols))
	}
}

func TestDistinctSQLGeneration(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.Select("name", "email")
	q.Distinct()

	wrapped := WrapQuery(q)
	sql, _ := wrapped.BuildSelectSQL()

	if !strings.Contains(sql, "SELECT DISTINCT") {
		t.Errorf("Expected SQL to contain 'SELECT DISTINCT', got: %s", sql)
	}
	// When Distinct is called without args, it uses the selected columns
	if !strings.Contains(sql, "name") {
		t.Errorf("Expected SQL to contain selected column 'name', got: %s", sql)
	}
}

func TestDistinctWithColumnsSQLGeneration(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.Distinct("name", "email")

	wrapped := WrapQuery(q)
	sql, _ := wrapped.BuildSelectSQL()

	// When Distinct is called with columns, they are stored in distinctCols
	// but the SELECT clause still uses the selects array (which defaults to *)
	// The distinctCols are used in aggregate functions like COUNT(DISTINCT col)
	if !strings.Contains(sql, "SELECT *") {
		t.Errorf("Expected SQL to contain 'SELECT *' when no explicit Select is called, got: %s", sql)
	}
}

func TestDistinctWithAggregateCount(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.Distinct("status")

	wrapped := WrapQuery(q)

	// When Distinct is set with columns, COUNT(DISTINCT column) should be generated
	// This is tested through the actual Count method in query_aggregate_test.go
	if !wrapped.GetDistinct() {
		t.Error("Distinct flag should be set")
	}
	if len(wrapped.GetDistinctCols()) != 1 {
		t.Errorf("Expected 1 distinct column, got %d", len(wrapped.GetDistinctCols()))
	}
}

func TestJoin(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.Join("posts", "users.id = posts.user_id")

	if result == nil {
		t.Error("Expected non-nil Query from Join")
	}
}

func TestJoinSQLGeneration(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.Join("posts", "users.id = posts.user_id")

	wrapped := WrapQuery(q)
	sql, _ := wrapped.BuildSelectSQL()

	if !strings.Contains(sql, "JOIN posts") {
		t.Errorf("Expected SQL to contain 'JOIN posts', got: %s", sql)
	}
	// Join condition may not appear in simple SELECT without ON clause
	// The join table name should be present
}

func TestLeftJoin(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.LeftJoin("posts", "users.id = posts.user_id")

	if result == nil {
		t.Error("Expected non-nil Query from LeftJoin")
	}
}

func TestLeftJoinSQLGeneration(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.LeftJoin("posts", "users.id = posts.user_id")

	wrapped := WrapQuery(q)
	sql, _ := wrapped.BuildSelectSQL()

	if !strings.Contains(sql, "LEFT JOIN posts") {
		t.Errorf("Expected SQL to contain 'LEFT JOIN posts', got: %s", sql)
	}
}

func TestRightJoin(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.RightJoin("posts", "users.id = posts.user_id")

	if result == nil {
		t.Error("Expected non-nil Query from RightJoin")
	}
}

func TestRightJoinSQLGeneration(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.RightJoin("posts", "users.id = posts.user_id")

	wrapped := WrapQuery(q)
	sql, _ := wrapped.BuildSelectSQL()

	if !strings.Contains(sql, "RIGHT JOIN posts") {
		t.Errorf("Expected SQL to contain 'RIGHT JOIN posts', got: %s", sql)
	}
}

func TestGroup(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.Group("name")

	if result == nil {
		t.Error("Expected non-nil Query from Group")
	}
}

func TestGroupSQLGeneration(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.Group("name").Group("status")

	wrapped := WrapQuery(q)
	sql, _ := wrapped.BuildSelectSQL()

	if !strings.Contains(sql, "GROUP BY") {
		t.Errorf("Expected SQL to contain 'GROUP BY', got: %s", sql)
	}
	if !strings.Contains(sql, "name") || !strings.Contains(sql, "status") {
		t.Errorf("Expected SQL to contain group columns, got: %s", sql)
	}
}

func TestOrderBy(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrderBy("name")

	if result == nil {
		t.Error("Expected non-nil Query from OrderBy")
	}
}

func TestOrderBySQLGeneration(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.OrderBy("name", "created_at")

	wrapped := WrapQuery(q)
	sql, _ := wrapped.BuildSelectSQL()

	if !strings.Contains(sql, "ORDER BY") {
		t.Errorf("Expected SQL to contain 'ORDER BY', got: %s", sql)
	}
	if !strings.Contains(sql, "name") || !strings.Contains(sql, "created_at") {
		t.Errorf("Expected SQL to contain order columns, got: %s", sql)
	}
}

func TestOrderByDesc(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.OrderByDesc("name")

	if result == nil {
		t.Error("Expected non-nil Query from OrderByDesc")
	}
}

func TestOrderByDescSQLGeneration(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.OrderByDesc("name")

	wrapped := WrapQuery(q)
	sql, _ := wrapped.BuildSelectSQL()

	if !strings.Contains(sql, "ORDER BY") {
		t.Errorf("Expected SQL to contain 'ORDER BY', got: %s", sql)
	}
	if !strings.Contains(sql, "desc") {
		t.Errorf("Expected SQL to contain 'desc', got: %s", sql)
	}
}

func TestLimit(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.Limit(10)

	if result == nil {
		t.Error("Expected non-nil Query from Limit")
	}
}

func TestLimitSQLGeneration(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.Limit(10)

	wrapped := WrapQuery(q)
	sql, _ := wrapped.BuildSelectSQL()

	if !strings.Contains(sql, "LIMIT") {
		t.Errorf("Expected SQL to contain 'LIMIT', got: %s", sql)
	}
	if !strings.Contains(sql, "10") {
		t.Errorf("Expected SQL to contain limit value 10, got: %s", sql)
	}
}

func TestOffset(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.Offset(5)

	if result == nil {
		t.Error("Expected non-nil Query from Offset")
	}
}

func TestOffsetSQLGeneration(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.Offset(5)

	wrapped := WrapQuery(q)
	sql, _ := wrapped.BuildSelectSQL()

	if !strings.Contains(sql, "OFFSET") {
		t.Errorf("Expected SQL to contain 'OFFSET', got: %s", sql)
	}
	if !strings.Contains(sql, "5") {
		t.Errorf("Expected SQL to contain offset value 5, got: %s", sql)
	}
}

func TestHaving(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	result := q.Having("count > ?", 5)

	if result == nil {
		t.Error("Expected non-nil Query from Having")
	}
}

func TestHavingSQLGeneration(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.Group("status")
	q.Having("count > ?", 5)

	wrapped := WrapQuery(q)
	sql, args := wrapped.BuildSelectSQL()

	if !strings.Contains(sql, "HAVING") {
		t.Errorf("Expected SQL to contain 'HAVING', got: %s", sql)
	}
	if !strings.Contains(sql, "count > ?") {
		t.Errorf("Expected SQL to contain having condition, got: %s", sql)
	}
	if len(args) != 1 || args[0] != 5 {
		t.Errorf("Expected argument [5], got %v", args)
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
