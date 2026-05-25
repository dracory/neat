package query

import (
	"context"
	"database/sql"
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

func TestRawWithSimpleSQL(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	rawQ := q.Raw("SELECT * FROM users WHERE id = ?", 1)

	if rawQ.(*Query).rawSQL != "SELECT * FROM users WHERE id = ?" {
		t.Errorf("Expected rawSQL to be 'SELECT * FROM users WHERE id = ?', got '%s'", rawQ.(*Query).rawSQL)
	}
	if len(rawQ.(*Query).rawArgs) != 1 {
		t.Errorf("Expected 1 rawArg, got %d", len(rawQ.(*Query).rawArgs))
	}
	if rawQ.(*Query).rawArgs[0] != 1 {
		t.Errorf("Expected rawArg[0] to be 1, got %v", rawQ.(*Query).rawArgs[0])
	}
}

func TestRawWithMultipleParameters(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	rawQ := q.Raw("SELECT * FROM users WHERE name = ? AND age > ?", "John", 25)

	if rawQ.(*Query).rawSQL != "SELECT * FROM users WHERE name = ? AND age > ?" {
		t.Errorf("Expected rawSQL to match, got '%s'", rawQ.(*Query).rawSQL)
	}
	if len(rawQ.(*Query).rawArgs) != 2 {
		t.Errorf("Expected 2 rawArgs, got %d", len(rawQ.(*Query).rawArgs))
	}
	if rawQ.(*Query).rawArgs[0] != "John" {
		t.Errorf("Expected rawArg[0] to be 'John', got %v", rawQ.(*Query).rawArgs[0])
	}
	if rawQ.(*Query).rawArgs[1] != 25 {
		t.Errorf("Expected rawArg[1] to be 25, got %v", rawQ.(*Query).rawArgs[1])
	}
}

func TestRawWithoutParameters(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	rawQ := q.Raw("SELECT * FROM users")

	if rawQ.(*Query).rawSQL != "SELECT * FROM users" {
		t.Errorf("Expected rawSQL to be 'SELECT * FROM users', got '%s'", rawQ.(*Query).rawSQL)
	}
	if len(rawQ.(*Query).rawArgs) != 0 {
		t.Errorf("Expected 0 rawArgs, got %d", len(rawQ.(*Query).rawArgs))
	}
}

func TestExecMethod(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	result, err := q.Exec("INSERT INTO test (name) VALUES (?)", "test_name")
	if err != nil {
		t.Errorf("Exec failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result to be returned")
	}
	if result.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", result.RowsAffected)
	}
}

func TestExecWithParameterBinding(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	result, err := q.Exec("INSERT INTO test (name, age) VALUES (?, ?)", "john", 30)
	if err != nil {
		t.Errorf("Exec with parameter binding failed: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", result.RowsAffected)
	}

	// Verify the data was inserted correctly
	var name string
	var age int
	err = db.QueryRow("SELECT name, age FROM test WHERE id = 1").Scan(&name, &age)
	if err != nil {
		t.Errorf("Failed to query inserted data: %v", err)
	}
	if name != "john" {
		t.Errorf("Expected name 'john', got '%s'", name)
	}
	if age != 30 {
		t.Errorf("Expected age 30, got %d", age)
	}
}

func TestExecWithMultipleParameters(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT, value INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	result, err := q.Exec("INSERT INTO test (name, value) VALUES (?, ?)", "test", 100)
	if err != nil {
		t.Errorf("Exec with multiple parameters failed: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", result.RowsAffected)
	}
}

func TestExecInTransaction(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.tx = tx
	q.inTransaction = true

	result, err := q.Exec("INSERT INTO test (name) VALUES (?)", "transaction_test")
	if err != nil {
		t.Errorf("Exec in transaction failed: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", result.RowsAffected)
	}

	// Verify the data is visible within the transaction
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM test").Scan(&count)
	if err != nil {
		t.Errorf("Failed to query within transaction: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 row in transaction, got %d", count)
	}
}

func TestExecWithUpdate(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	_, err = db.Exec("INSERT INTO test (name) VALUES ('original')")
	if err != nil {
		t.Fatalf("Failed to insert initial data: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	result, err := q.Exec("UPDATE test SET name = ? WHERE id = ?", "updated", 1)
	if err != nil {
		t.Errorf("Exec with UPDATE failed: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", result.RowsAffected)
	}

	// Verify the update
	var name string
	err = db.QueryRow("SELECT name FROM test WHERE id = 1").Scan(&name)
	if err != nil {
		t.Errorf("Failed to query updated data: %v", err)
	}
	if name != "updated" {
		t.Errorf("Expected name 'updated', got '%s'", name)
	}
}

func TestExecWithDelete(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	_, err = db.Exec("INSERT INTO test (name) VALUES ('to_delete')")
	if err != nil {
		t.Fatalf("Failed to insert initial data: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	result, err := q.Exec("DELETE FROM test WHERE id = ?", 1)
	if err != nil {
		t.Errorf("Exec with DELETE failed: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", result.RowsAffected)
	}

	// Verify the deletion
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test").Scan(&count)
	if err != nil {
		t.Errorf("Failed to query count: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 rows after delete, got %d", count)
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
