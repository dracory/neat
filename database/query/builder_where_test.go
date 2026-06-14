package query

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/dracory/neat/database/schema/constants"
)

// whereTestSoftModel is a package-level type used by builder_where tests.
// It implements SoftDeleteColumnNamer so the query builder applies soft-delete filtering.
type whereTestSoftModel struct {
	ID        int
	Name      string
	DeletedAt *time.Time
}

func (m *whereTestSoftModel) SoftDeletedAtColumn() string { return constants.DeletedAtColumnName }

func TestBuildWheres(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.wheres = []whereClause{
		{_type: "", query: "name = ?", args: []any{"Alice"}},
		{_type: "AND", query: "age > ?", args: []any{18}},
	}
	b := NewBuilder(q)

	where, args := b.buildWheres()

	if where == "" {
		t.Error("Expected non-empty WHERE clause")
	}
	if !strings.Contains(where, "name = ?") {
		t.Error("Expected 'name = ?' in WHERE clause")
	}
	if !strings.Contains(where, "age > ?") {
		t.Error("Expected 'age > ?' in WHERE clause")
	}
	if args == nil {
		t.Error("Expected non-nil args")
	}
}

func TestBuildWheresWithIN(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.wheres = []whereClause{
		{_type: "", query: "id IN (?)", args: []any{[]any{1, 2, 3}}},
	}
	b := NewBuilder(q)

	where, args := b.buildWheres()

	if where == "" {
		t.Error("Expected non-empty WHERE clause")
	}
	// IN (?) should be expanded to IN (?, ?, ?)
	if !strings.Contains(where, "IN (?, ?, ?)") {
		t.Error("Expected IN clause to be expanded")
	}
	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d", len(args))
	}
}

func TestBuildWheresEmpty(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	b := NewBuilder(q)

	where, args := b.buildWheres()

	if where != "" {
		t.Error("Expected empty WHERE clause")
	}
	if len(args) != 0 {
		t.Error("Expected empty args")
	}
}

func TestLaravelStyleWhere(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Where("name", "Alice")
	b := NewBuilder(q)

	where, args := b.buildWheres()

	if !strings.Contains(where, "name = ?") {
		t.Errorf("Expected 'name = ?' in WHERE clause, got %s", where)
	}
	if len(args) != 1 || args[0] != "Alice" {
		t.Errorf("Expected args [Alice], got %v", args)
	}
}

func TestLaravelStyleOrWhere(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.OrWhere("name", "Bob")
	b := NewBuilder(q)

	where, args := b.buildWheres()

	if !strings.Contains(where, "name = ?") {
		t.Errorf("Expected 'name = ?' in WHERE clause, got %s", where)
	}
	if len(args) != 1 || args[0] != "Bob" {
		t.Errorf("Expected args [Bob], got %v", args)
	}
}

func TestExplicitOperatorWhere(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Where("age > ?", 18)
	b := NewBuilder(q)

	where, args := b.buildWheres()

	if !strings.Contains(where, "age > ?") {
		t.Errorf("Expected 'age > ?' in WHERE clause, got %s", where)
	}
	if len(args) != 1 || args[0] != 18 {
		t.Errorf("Expected args [18], got %v", args)
	}
}

func TestMixedWhereStyles(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Where("name", "Alice")
	q.Where("age > ?", 18)
	b := NewBuilder(q)

	where, args := b.buildWheres()

	if !strings.Contains(where, "name = ?") {
		t.Errorf("Expected 'name = ?' in WHERE clause, got %s", where)
	}
	if !strings.Contains(where, "age > ?") {
		t.Errorf("Expected 'age > ?' in WHERE clause, got %s", where)
	}
	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}
}

func TestContainsOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"name = ?", true},
		{"age > ?", true},
		{"age < ?", true},
		{"age >= ?", true},
		{"age <= ?", true},
		{"age != ?", true},
		{"age <> ?", true},
		{"name LIKE ?", true},
		{"name NOT LIKE ?", true},
		{"id IN (?)", true},
		{"id NOT IN (?)", true},
		{"age BETWEEN ? AND ?", true},
		{"age NOT BETWEEN ? AND ?", true},
		{"name=?", true},
		{"age>?", true},
		{"age<?", true},
		{"age>=?", true},
		{"age<=?", true},
		{"age!=?", true},
		{"age<>?", true},
		{"name", false},
		{"internal", false},
		{"between", false},
		{"like", false},
		{"column_name", false},
		{"= value", true},
		{"value =", true},
		{"> value", true},
		{"value >", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := containsOperator(tt.input)
			if result != tt.expected {
				t.Errorf("containsOperator(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestWhereWithNoArgs(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Where("name IS NULL")
	b := NewBuilder(q)

	where, args := b.buildWheres()

	if !strings.Contains(where, "name IS NULL") {
		t.Errorf("Expected 'name IS NULL' in WHERE clause, got %s", where)
	}
	if len(args) != 0 {
		t.Errorf("Expected 0 args, got %d", len(args))
	}
}

func TestWhereWithMultipleArgs(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Where("BETWEEN", 1, 10)
	b := NewBuilder(q)

	where, args := b.buildWheres()

	// Should not transform since it has 2 args
	if !strings.Contains(where, "BETWEEN") {
		t.Errorf("Expected 'BETWEEN' in WHERE clause, got %s", where)
	}
	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}
}

func TestWhereWithColumnContainingOperatorLikeString(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Where("internal", "value")
	b := NewBuilder(q)

	where, args := b.buildWheres()

	// Should transform to "internal = ?" since "internal" doesn't contain operator
	if !strings.Contains(where, "internal = ?") {
		t.Errorf("Expected 'internal = ?' in WHERE clause, got %s", where)
	}
	if len(args) != 1 || args[0] != "value" {
		t.Errorf("Expected args [value], got %v", args)
	}
}

func TestWhereWithMixedCaseOperator(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Where("name like ?", "test")
	b := NewBuilder(q)

	where, args := b.buildWheres()

	// Should not transform since it contains operator
	if !strings.Contains(where, "like ?") {
		t.Errorf("Expected 'like ?' in WHERE clause, got %s", where)
	}
	if len(args) != 1 || args[0] != "test" {
		t.Errorf("Expected args [test], got %v", args)
	}
}

func TestWhereWithNilArgs(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Where("name")
	b := NewBuilder(q)

	where, args := b.buildWheres()

	// Should not transform since args is nil/empty
	if !strings.Contains(where, "name") {
		t.Errorf("Expected 'name' in WHERE clause, got %s", where)
	}
	if len(args) != 0 {
		t.Errorf("Expected 0 args, got %d", len(args))
	}
}

func TestBuildWheresWithSoftDelete(t *testing.T) {
	model := &whereTestSoftModel{}
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.model = model
	q.Where("name = ?", "Alice")
	b := NewBuilder(q)

	where, args := b.buildWheresWithSoftDelete()

	// Should include soft-delete filter
	if !strings.Contains(where, "deleted_at IS NULL") {
		t.Errorf("Expected 'deleted_at IS NULL' in WHERE clause, got %s", where)
	}
	if !strings.Contains(where, "name = ?") {
		t.Errorf("Expected 'name = ?' in WHERE clause, got %s", where)
	}
	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(args))
	}
}

func TestBuildWheresWithSoftDeleteOnlySoftDeleted(t *testing.T) {
	model := &whereTestSoftModel{}
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.model = model
	q = q.OnlySoftDeleted().(*Query)
	q.Where("name = ?", "Alice")
	b := NewBuilder(q)

	where, args := b.buildWheresWithSoftDelete()

	// Should include only trashed filter
	if !strings.Contains(where, "deleted_at IS NOT NULL") {
		t.Errorf("Expected 'deleted_at IS NOT NULL' in WHERE clause, got %s", where)
	}
	if !strings.Contains(where, "name = ?") {
		t.Errorf("Expected 'name = ?' in WHERE clause, got %s", where)
	}
	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(args))
	}
}

func TestBuildWheresWithSoftDeleteWithSoftDeleted(t *testing.T) {
	model := &whereTestSoftModel{}
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.model = model
	q = q.WithSoftDeleted().(*Query)
	q.Where("name = ?", "Alice")
	b := NewBuilder(q)

	where, args := b.buildWheresWithSoftDelete()

	// Should NOT include soft-delete filter
	if strings.Contains(where, constants.DeletedAtColumnName) {
		t.Errorf("Expected no soft-delete filter in WHERE clause, got %s", where)
	}
	if !strings.Contains(where, "name = ?") {
		t.Errorf("Expected 'name = ?' in WHERE clause, got %s", where)
	}
	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(args))
	}
}

func TestBuildWheresWithSoftDeleteNoModel(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Where("name = ?", "Alice")
	b := NewBuilder(q)

	where, args := b.buildWheresWithSoftDelete()

	// Should NOT include soft-delete filter (no model)
	if strings.Contains(where, constants.DeletedAtColumnName) {
		t.Errorf("Expected no soft-delete filter in WHERE clause, got %s", where)
	}
	if !strings.Contains(where, "name = ?") {
		t.Errorf("Expected 'name = ?' in WHERE clause, got %s", where)
	}
	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(args))
	}
}

func TestBuildWheresWithIndex(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.wheres = []whereClause{
		{_type: "", query: "name = ?", args: []any{"Alice"}},
		{_type: "AND", query: "age > ?", args: []any{18}},
	}
	b := NewBuilder(q)

	where, args := b.buildWheresWithIndex(5)

	if where == "" {
		t.Error("Expected non-empty WHERE clause")
	}
	if !strings.Contains(where, "name = ?") {
		t.Error("Expected 'name = ?' in WHERE clause")
	}
	if !strings.Contains(where, "age > ?") {
		t.Error("Expected 'age > ?' in WHERE clause")
	}
	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}
}

func TestBuildWheresWithSoftDeleteIndex(t *testing.T) {
	model := &whereTestSoftModel{}
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.model = model
	q.Where("name = ?", "Alice")
	b := NewBuilder(q)

	where, args := b.buildWheresWithSoftDeleteIndex(10)

	// Should include soft-delete filter
	if !strings.Contains(where, "deleted_at IS NULL") {
		t.Errorf("Expected 'deleted_at IS NULL' in WHERE clause, got %s", where)
	}
	if !strings.Contains(where, "name = ?") {
		t.Errorf("Expected 'name = ?' in WHERE clause, got %s", where)
	}
	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(args))
	}
}
