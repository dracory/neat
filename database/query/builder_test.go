package query

import (
	"context"
	"strings"
	"testing"
)

func TestNewBuilder(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	b := NewBuilder(q)
	if b == nil {
		t.Error("Expected non-nil Builder")
	}
	if b.query != q {
		t.Error("Expected Builder to have the provided query")
	}
}

func TestBuilderPlaceholderDefault(t *testing.T) {
	// Without a driver, quoteIdentifier returns unquoted names.
	// Binding placeholders in the built SQL should be '?'.
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.Where("id = ?", 42)

	sql, args := NewBuilder(q).BuildSelect()

	if !strings.Contains(sql, "?") {
		t.Errorf("Expected '?' placeholder in SQL, got: %s", sql)
	}
	if len(args) != 1 || args[0] != 42 {
		t.Errorf("Expected args [42], got %v", args)
	}
}

func TestBuilderPlaceholderPostgres(t *testing.T) {
	// PostgreSQL driver replaces '?' with '$N' during execution,
	// but BuildSelect itself stores '?' — replacement is a driver concern.
	// This test verifies the builder still accepts and passes args correctly.
	drv := &FakeDriver{DialectName: "postgres"}
	q := NewQuery(context.TODO(), nil, drv, "", nil, nil)
	q.Table("users")
	q.Where("id = ?", 7)

	_, args := NewBuilder(q).BuildSelect()

	if len(args) != 1 || args[0] != 7 {
		t.Errorf("Expected args [7] for postgres builder, got %v", args)
	}
}

func TestBuilderTableNameWrapping(t *testing.T) {
	drv := &FakeDriver{DialectName: "postgres"}
	q := NewQuery(context.TODO(), nil, drv, "", nil, nil)
	q.Table("users")

	sql, _ := NewBuilder(q).BuildSelect()

	if !strings.Contains(sql, `"users"`) {
		t.Errorf("Expected quoted table name \"users\", got: %s", sql)
	}
}

func TestBuilderTableNameWrappingMySQL(t *testing.T) {
	drv := &FakeDriver{DialectName: "mysql"}
	q := NewQuery(context.TODO(), nil, drv, "", nil, nil)
	q.Table("orders")

	sql, _ := NewBuilder(q).BuildSelect()

	if !strings.Contains(sql, "`orders`") {
		t.Errorf("Expected backtick-quoted table name `orders`, got: %s", sql)
	}
}

func TestBuilderColumnNameWrapping(t *testing.T) {
	drv := &FakeDriver{DialectName: "postgres"}
	q := NewQuery(context.TODO(), nil, drv, "", nil, nil)
	q.Table("users")
	q.WhereIn("id", []any{1, 2, 3})

	sql, _ := NewBuilder(q).BuildSelect()

	if !strings.Contains(sql, `"id"`) {
		t.Errorf("Expected quoted column name \"id\" in WHERE clause, got: %s", sql)
	}
}

func TestBuilderIdentifierEscapingSpecialChars(t *testing.T) {
	// Dotted identifiers (table.column) should quote each part separately.
	drv := &FakeDriver{DialectName: "postgres"}
	q := NewQuery(context.TODO(), nil, drv, "", nil, nil)
	b := NewBuilder(q)

	result := b.quoteIdentifier("orders.user_id")

	if !strings.Contains(result, `"orders"`) || !strings.Contains(result, `"user_id"`) {
		t.Errorf("Expected dotted identifier to quote each part, got: %s", result)
	}
}

func TestBuilderIdentifierAlreadyQuoted(t *testing.T) {
	drv := &FakeDriver{DialectName: "postgres"}
	q := NewQuery(context.TODO(), nil, drv, "", nil, nil)
	b := NewBuilder(q)

	already := `"name"`
	result := b.quoteIdentifier(already)

	if result != already {
		t.Errorf("Already-quoted identifier should be returned as-is, got: %s", result)
	}
}

func TestBuilderQueryCompilationSelectFromWhere(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("products")
	q.Select("id")
	q.Where("price > ?", 100)
	q.OrderBy("price", "desc")
	q.Limit(5)

	sql, args := NewBuilder(q).BuildSelect()

	for _, expected := range []string{"SELECT", "FROM", "WHERE", "ORDER BY", "LIMIT"} {
		if !strings.Contains(sql, expected) {
			t.Errorf("Expected compiled SQL to contain %q, got: %s", expected, sql)
		}
	}
	if len(args) != 1 || args[0] != 100 {
		t.Errorf("Expected args [100], got %v", args)
	}
}

func TestBuilderMultipleBindingArgs(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")
	q.Where("age > ?", 18)
	q.Where("age < ?", 65)
	q.Where("status = ?", "active")

	_, args := NewBuilder(q).BuildSelect()

	if len(args) != 3 {
		t.Errorf("Expected 3 binding args, got %d: %v", len(args), args)
	}
}
