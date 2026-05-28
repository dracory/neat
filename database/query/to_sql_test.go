package query

import (
	"context"
	"strings"
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

func TestToSqlSaveInsert(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	type User struct {
		ID   int
		Name string
	}

	user := &User{Name: "John Doe"}
	toSql := q.ToSql()
	sql := toSql.Save(user)

	// Should generate INSERT since ID is 0
	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	// Check that it contains INSERT
	if !strings.Contains(sql, "INSERT") {
		t.Errorf("Expected INSERT in SQL, got %q", sql)
	}
	// Check that it contains the table name
	if !strings.Contains(sql, "users") {
		t.Errorf("Expected 'users' in SQL, got %q", sql)
	}
}

func TestToSqlSaveUpdate(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	type User struct {
		ID   int
		Name string
	}

	user := &User{ID: 1, Name: "John Doe"}
	toSql := q.ToSql()
	sql := toSql.Save(user)

	// Should generate UPDATE since ID is non-zero
	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	// Check that it contains UPDATE
	if !strings.Contains(sql, "UPDATE") {
		t.Errorf("Expected UPDATE in SQL, got %q", sql)
	}
	// Check that it contains WHERE id = ?
	if !strings.Contains(sql, "WHERE") || !strings.Contains(sql, "id = ?") {
		t.Errorf("Expected WHERE id = ? in SQL, got %q", sql)
	}
}

func TestToSqlSaveUpdateWithValues(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	type User struct {
		ID   int
		Name string
	}

	user := &User{ID: 1, Name: "John Doe"}
	toSql := q.ToRawSql()
	sql := toSql.Save(user)

	// Should generate UPDATE with actual values
	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	// Check that it contains UPDATE
	if !strings.Contains(sql, "UPDATE") {
		t.Errorf("Expected UPDATE in SQL, got %q", sql)
	}
	// Check that it contains WHERE id = 1 (actual value, not placeholder)
	if !strings.Contains(sql, "WHERE") || !strings.Contains(sql, "id = 1") {
		t.Errorf("Expected WHERE id = 1 in SQL, got %q", sql)
	}
}

func TestToSqlSaveInsertWithValues(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"

	type User struct {
		ID   int
		Name string
	}

	user := &User{Name: "John Doe"}
	toSql := q.ToRawSql()
	sql := toSql.Save(user)

	// Should generate INSERT with actual values
	if sql == "" {
		t.Error("Expected non-empty SQL")
	}
	// Check that it contains INSERT
	if !strings.Contains(sql, "INSERT") {
		t.Errorf("Expected INSERT in SQL, got %q", sql)
	}
	// Check that it contains the actual name value
	if !strings.Contains(sql, "'John Doe'") {
		t.Errorf("Expected 'John Doe' in SQL, got %q", sql)
	}
}

// TestToSqlCreate tests SQL generation for Create.
func TestToSqlCreate(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")

	model := struct {
		Name string
	}{Name: "John"}

	toSql := q.ToSql()
	sql := toSql.Create(model)

	if sql == "" {
		t.Error("Expected SQL to be generated for Create")
	}
}

// TestToSqlDelete tests SQL generation for Delete.
func TestToSqlDelete(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")

	toSql := q.ToSql()
	sql := toSql.Delete()

	if sql == "" {
		t.Error("Expected SQL to be generated for Delete")
	}
}

// TestToSqlFirst tests SQL generation for First.
func TestToSqlFirst(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")

	toSql := q.ToSql()
	sql := toSql.First(nil)

	if sql == "" {
		t.Error("Expected SQL to be generated for First")
	}
}

// TestToSqlGet tests SQL generation for Get.
func TestToSqlGet(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")

	toSql := q.ToSql()
	sql := toSql.Get(nil)

	if sql == "" {
		t.Error("Expected SQL to be generated for Get")
	}
}

// TestToSqlUpdate tests SQL generation for Update.
func TestToSqlUpdate(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")

	toSql := q.ToSql()
	sql := toSql.Update("name", "John")

	if sql == "" {
		t.Error("Expected SQL to be generated for Update")
	}
}
