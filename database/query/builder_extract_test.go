package query

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestExtractColumnsAndValuesStruct(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	b := NewBuilder(q)

	type User struct {
		Name  string
		Email string
		Age   int
	}

	user := User{Name: "Alice", Email: "alice@example.com", Age: 30}
	cols, vals, err := b.extractColumnsAndValues(user)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(cols) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(cols))
	}
	if len(vals) != 3 {
		t.Errorf("Expected 3 values, got %d", len(vals))
	}
}

func TestExtractColumnsAndValuesMap(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	b := NewBuilder(q)

	data := map[string]any{"name": "Bob", "age": 25}
	cols, vals, err := b.extractColumnsAndValues(data)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(cols) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(cols))
	}
	if len(vals) != 2 {
		t.Errorf("Expected 2 values, got %d", len(vals))
	}
}

func TestExtractColumnsAndValuesSlice(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	b := NewBuilder(q)

	type User struct {
		Name  string
		Email string
	}

	users := []User{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}

	cols, vals, err := b.extractColumnsAndValues(users)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(cols) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(cols))
	}
	if len(vals) != 4 {
		t.Errorf("Expected 4 values (2 users × 2 fields), got %d", len(vals))
	}
}

func TestExtractColumnsAndValuesEmptySlice(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	b := NewBuilder(q)

	var users []struct {
		Name string
	}

	cols, vals, err := b.extractColumnsAndValues(users)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cols != nil {
		t.Error("Expected nil columns for empty slice")
	}
	if len(vals) != 0 {
		t.Error("Expected empty values for empty slice")
	}
}

func TestExtractSingleColumnsAndValuesStruct(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	b := NewBuilder(q)

	type User struct {
		Name  string
		Email string
	}

	user := User{Name: "Alice", Email: "alice@example.com"}
	cols, vals, err := b.extractSingleColumnsAndValues(user)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(cols) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(cols))
	}
	if len(vals) != 2 {
		t.Errorf("Expected 2 values, got %d", len(vals))
	}
}

func TestExtractSingleColumnsAndValuesMap(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	b := NewBuilder(q)

	data := map[string]any{"name": "Bob", "age": 25}
	cols, vals, err := b.extractSingleColumnsAndValues(data)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(cols) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(cols))
	}
	if len(vals) != 2 {
		t.Errorf("Expected 2 values, got %d", len(vals))
	}
}

func TestExtractStructColumnsAndValues(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	b := NewBuilder(q)

	type User struct {
		Name  string
		Email string
		Age   int
	}

	user := User{Name: "Alice", Email: "alice@example.com", Age: 30}
	cols, vals := b.extractStructColumnsAndValues(reflect.ValueOf(user))

	if len(cols) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(cols))
	}
	if len(vals) != 3 {
		t.Errorf("Expected 3 values, got %d", len(vals))
	}
}

func TestExtractStructColumnsAndValuesWithOmitted(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.omitColumns = []string{"password"}
	b := NewBuilder(q)

	type User struct {
		Name     string
		Email    string
		Password string
	}

	user := User{Name: "Alice", Email: "alice@example.com", Password: "secret"}
	cols, vals := b.extractStructColumnsAndValues(reflect.ValueOf(user))

	if len(cols) != 2 {
		t.Errorf("Expected 2 columns (password omitted), got %d", len(cols))
	}
	if len(vals) != 2 {
		t.Errorf("Expected 2 values (password omitted), got %d", len(vals))
	}
}

func TestExtractStructColumnsAndValuesWithZeroValues(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	b := NewBuilder(q)

	type User struct {
		Name  string
		Email string
		Age   int
	}

	user := User{Name: "Alice", Email: "", Age: 0}
	cols, vals := b.extractStructColumnsAndValues(reflect.ValueOf(user))

	// Zero values should be skipped (except for boolean and time.Time)
	if len(cols) != 1 {
		t.Errorf("Expected 1 column (only non-zero), got %d", len(cols))
	}
	if len(vals) != 1 {
		t.Errorf("Expected 1 value (only non-zero), got %d", len(vals))
	}
}

func TestExtractStructColumnsAndValuesWithTime(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	b := NewBuilder(q)

	type User struct {
		Name      string
		CreatedAt time.Time
	}

	user := User{Name: "Alice", CreatedAt: time.Now()}
	cols, vals := b.extractStructColumnsAndValues(reflect.ValueOf(user))

	// time.Time zero values should be included
	if len(cols) != 2 {
		t.Errorf("Expected 2 columns (time.Time included), got %d", len(cols))
	}
	if len(vals) != 2 {
		t.Errorf("Expected 2 values (time.Time included), got %d", len(vals))
	}
}

func TestExtractColumnNames(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	b := NewBuilder(q)

	type User struct {
		Name  string
		Email string
		Age   int
	}

	user := User{Name: "Alice", Email: "alice@example.com", Age: 30}
	cols := b.extractColumnNames(user)

	if len(cols) != 3 {
		t.Errorf("Expected 3 column names, got %d", len(cols))
	}
}

func TestExtractStructColumnNames(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	b := NewBuilder(q)

	type User struct {
		Name  string
		Email string
		Age   int
	}

	user := User{Name: "Alice", Email: "alice@example.com", Age: 30}
	cols := b.extractStructColumnNames(reflect.ValueOf(user))

	if len(cols) != 3 {
		t.Errorf("Expected 3 column names, got %d", len(cols))
	}
}

func TestExtractColumnsAndValuesUnsupportedType(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	b := NewBuilder(q)

	_, _, err := b.extractSingleColumnsAndValues("invalid")

	if err == nil {
		t.Error("Expected error for unsupported type")
	}
}
