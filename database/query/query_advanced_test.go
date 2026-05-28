package query

import (
	"context"
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/database/driver"
	_ "modernc.org/sqlite"
)

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

// --- Lock Clause Tests ---

func TestLockForUpdate(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	result := q.LockForUpdate()

	if result == nil {
		t.Error("Expected non-nil Query from LockForUpdate")
	}

	if !result.(*Query).lockForUpdate {
		t.Error("Expected lockForUpdate flag to be set")
	}
}

func TestSharedLock(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	result := q.SharedLock()

	if result == nil {
		t.Error("Expected non-nil Query from SharedLock")
	}

	if !result.(*Query).sharedLock {
		t.Error("Expected sharedLock flag to be set")
	}
}

func TestLockForUpdate_SqlGeneration(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewMySQL(), "", nil, nil)
	q.Table("users")
	q = q.LockForUpdate().(*Query)

	// Verify driver is set
	if q.driver == nil {
		t.Fatal("Driver should not be nil")
	}
	if q.driver.Dialect() != "mysql" {
		t.Errorf("Expected mysql dialect, got: %s", q.driver.Dialect())
	}

	// Verify lock flag is set
	if !q.lockForUpdate {
		t.Error("lockForUpdate flag should be set")
	}

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "FOR UPDATE") {
		t.Errorf("Expected FOR UPDATE in SQL, got: %s", sql)
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 arguments, got %d", len(args))
	}
}

func TestSharedLock_SqlGeneration(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewMySQL(), "", nil, nil)
	q.Table("users")
	q = q.SharedLock().(*Query)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "LOCK IN SHARE MODE") {
		t.Errorf("Expected LOCK IN SHARE MODE in SQL, got: %s", sql)
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 arguments, got %d", len(args))
	}
}

func TestLockForUpdate_WithWhereClause(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewMySQL(), "", nil, nil)
	q.Table("users")
	q.Where("id = ?", 1)
	q = q.LockForUpdate().(*Query)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "FOR UPDATE") {
		t.Errorf("Expected FOR UPDATE in SQL, got: %s", sql)
	}

	if !contains(sql, "WHERE") {
		t.Error("Expected WHERE clause in SQL")
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(args))
	}
}

func TestSharedLock_WithWhereClause(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewMySQL(), "", nil, nil)
	q.Table("users")
	q.Where("status = ?", "active")
	q = q.SharedLock().(*Query)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "LOCK IN SHARE MODE") {
		t.Errorf("Expected LOCK IN SHARE MODE in SQL, got: %s", sql)
	}

	if !contains(sql, "WHERE") {
		t.Error("Expected WHERE clause in SQL")
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(args))
	}
}

func TestLockForUpdate_WithLimit(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewMySQL(), "", nil, nil)
	q.Table("users")
	q.Limit(10)
	q = q.LockForUpdate().(*Query)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "FOR UPDATE") {
		t.Errorf("Expected FOR UPDATE in SQL, got: %s", sql)
	}

	if !contains(sql, "LIMIT") {
		t.Error("Expected LIMIT clause in SQL")
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 arguments, got %d", len(args))
	}
}

func TestSharedLock_WithOrderBy(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewMySQL(), "", nil, nil)
	q.Table("users")
	q.OrderBy("created_at", "DESC")
	q = q.SharedLock().(*Query)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	if !contains(sql, "LOCK IN SHARE MODE") {
		t.Errorf("Expected LOCK IN SHARE MODE in SQL, got: %s", sql)
	}

	if !contains(sql, "ORDER BY") {
		t.Error("Expected ORDER BY clause in SQL")
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 arguments, got %d", len(args))
	}
}

func TestLockForUpdate_Clone(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	lockedQ := q.LockForUpdate()

	// Original query should not have lock flag
	if q.lockForUpdate {
		t.Error("Original query should not have lockForUpdate flag set")
	}

	// Cloned query should have lock flag
	if !lockedQ.(*Query).lockForUpdate {
		t.Error("Cloned query should have lockForUpdate flag set")
	}
}

func TestSharedLock_Clone(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "users", nil, nil)
	lockedQ := q.SharedLock()

	// Original query should not have lock flag
	if q.sharedLock {
		t.Error("Original query should not have sharedLock flag set")
	}

	// Cloned query should have lock flag
	if !lockedQ.(*Query).sharedLock {
		t.Error("Cloned query should have sharedLock flag set")
	}
}

func TestLockForUpdate_PrecedenceOverSharedLock(t *testing.T) {
	q := NewQuery(context.TODO(), nil, driver.NewMySQL(), "", nil, nil)
	q.Table("users")
	q = q.SharedLock().(*Query)
	q = q.LockForUpdate().(*Query)

	builder := NewBuilder(q)
	sql, _ := builder.BuildSelect()

	// LockForUpdate should take precedence
	if !contains(sql, "FOR UPDATE") {
		t.Errorf("Expected FOR UPDATE (LockForUpdate takes precedence), got: %s", sql)
	}

	if contains(sql, "LOCK IN SHARE MODE") {
		t.Error("Should not have LOCK IN SHARE MODE when LockForUpdate is set")
	}
}

func TestLockForUpdate_DialectDifferences(t *testing.T) {
	dialects := []struct {
		name   string
		driver driver.Driver
	}{
		{"MySQL", driver.NewMySQL()},
		{"PostgreSQL", driver.NewPostgreSQL()},
	}

	for _, tc := range dialects {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQuery(context.TODO(), nil, tc.driver, "", nil, nil)
			q.Table("users")
			q = q.LockForUpdate().(*Query)

			builder := NewBuilder(q)
			sql, _ := builder.BuildSelect()

			// All dialects use FOR UPDATE syntax
			if !contains(sql, "FOR UPDATE") {
				t.Errorf("Expected FOR UPDATE in %s, got: %s", tc.name, sql)
			}
		})
	}
}

func TestSharedLock_DialectDifferences(t *testing.T) {
	dialects := []struct {
		name   string
		driver driver.Driver
	}{
		{"MySQL", driver.NewMySQL()},
		{"PostgreSQL", driver.NewPostgreSQL()},
	}

	for _, tc := range dialects {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQuery(context.TODO(), nil, tc.driver, "", nil, nil)
			q.Table("users")
			q = q.SharedLock().(*Query)

			builder := NewBuilder(q)
			sql, _ := builder.BuildSelect()

			// All dialects use LOCK IN SHARE MODE syntax
			if !contains(sql, "LOCK IN SHARE MODE") {
				t.Errorf("Expected LOCK IN SHARE MODE in %s, got: %s", tc.name, sql)
			}
		})
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
