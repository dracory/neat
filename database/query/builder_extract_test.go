package query

import (
	"context"
	"reflect"
	"testing"
	"time"

	_ "modernc.org/sqlite"
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

	// Zero values for strings and integers are included (only boolean, time.Time, and some types are skipped)
	if len(cols) != 3 {
		t.Errorf("Expected 3 columns (strings and integers included even when zero), got %d", len(cols))
	}
	if len(vals) != 3 {
		t.Errorf("Expected 3 values (strings and integers included even when zero), got %d", len(vals))
	}
}

func TestExtractStructColumnsAndValuesWithTime(t *testing.T) {
	q := NewQuery(context.TODO(), nil, &FakeDriver{DialectName: "sqlite"}, "", nil, nil)
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

	// For SQLite, time.Time value should be converted to a datetime string
	if _, ok := vals[1].(string); !ok {
		t.Errorf("Expected time.Time to be converted to string for SQLite, got %T", vals[1])
	}
}

func TestExtractStructColumnsAndValuesTimeConvertedToString(t *testing.T) {
	q := NewQuery(context.TODO(), nil, &FakeDriver{DialectName: "sqlite"}, "", nil, nil)
	b := NewBuilder(q)

	type User struct {
		Name      string    `db:"name"`
		CreatedAt time.Time `db:"created_at"`
	}

	ts := time.Date(2026, 6, 20, 12, 34, 56, 0, time.UTC)
	user := User{Name: "Alice", CreatedAt: ts}
	_, vals := b.extractStructColumnsAndValues(reflect.ValueOf(user))

	if s, ok := vals[1].(string); !ok || s != "2026-06-20 12:34:56" {
		t.Errorf("Expected time to be converted to '2026-06-20 12:34:56', got %v (%T)", vals[1], vals[1])
	}
}

func TestExtractStructColumnsAndValuesTimePassedAsIsForNonSQLite(t *testing.T) {
	q := NewQuery(context.TODO(), nil, &FakeDriver{DialectName: "mysql"}, "", nil, nil)
	b := NewBuilder(q)

	type User struct {
		Name      string    `db:"name"`
		CreatedAt time.Time `db:"created_at"`
	}

	ts := time.Date(2026, 6, 20, 12, 34, 56, 0, time.UTC)
	user := User{Name: "Alice", CreatedAt: ts}
	_, vals := b.extractStructColumnsAndValues(reflect.ValueOf(user))

	if tt, ok := vals[1].(time.Time); !ok || !tt.Equal(ts) {
		t.Errorf("Expected time.Time to be passed as-is for MySQL, got %v (%T)", vals[1], vals[1])
	}
}

func TestExtractMapValuesTimeConvertedToString(t *testing.T) {
	q := NewQuery(context.TODO(), nil, &FakeDriver{DialectName: "sqlite"}, "", nil, nil)
	b := NewBuilder(q)

	ts := time.Date(2026, 6, 20, 12, 34, 56, 0, time.UTC)
	data := map[string]any{"created_at": ts}
	_, vals, err := b.extractColumnsAndValues(data)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if s, ok := vals[0].(string); !ok || s != "2026-06-20 12:34:56" {
		t.Errorf("Expected time.Time in map to be converted to '2026-06-20 12:34:56', got %v (%T)", vals[0], vals[0])
	}
}

func TestExtractMapValuesTimePassedAsIsForNonSQLite(t *testing.T) {
	q := NewQuery(context.TODO(), nil, &FakeDriver{DialectName: "oracle"}, "", nil, nil)
	b := NewBuilder(q)

	ts := time.Date(2026, 6, 20, 12, 34, 56, 0, time.UTC)
	data := map[string]any{"created_at": ts}
	_, vals, err := b.extractColumnsAndValues(data)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if tt, ok := vals[0].(time.Time); !ok || !tt.Equal(ts) {
		t.Errorf("Expected time.Time in map to be passed as-is for Oracle, got %v (%T)", vals[0], vals[0])
	}
}

func TestExtractMapValuesPtrTimeConvertedToString(t *testing.T) {
	q := NewQuery(context.TODO(), nil, &FakeDriver{DialectName: "sqlite"}, "", nil, nil)
	b := NewBuilder(q)

	ts := time.Date(2026, 6, 20, 12, 34, 56, 0, time.UTC)
	data := map[string]any{"deleted_at": &ts}
	_, vals, err := b.extractColumnsAndValues(data)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if s, ok := vals[0].(string); !ok || s != "2026-06-20 12:34:56" {
		t.Errorf("Expected *time.Time in map to be converted to '2026-06-20 12:34:56', got %v (%T)", vals[0], vals[0])
	}
}

func TestExtractMapValuesNilPtrTimeRemainsNil(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	b := NewBuilder(q)

	var nilTime *time.Time
	data := map[string]any{"deleted_at": nilTime}
	_, vals, err := b.extractColumnsAndValues(data)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// Nil *time.Time should not be converted to a string; verify it stays as *time.Time nil
	if _, ok := vals[0].(string); ok {
		t.Errorf("Expected nil *time.Time in map NOT to be converted to string, got %v", vals[0])
	}
	if ptr, ok := vals[0].(*time.Time); !ok || ptr != nil {
		t.Errorf("Expected nil *time.Time to remain as (*time.Time)(nil), got %v (%T)", vals[0], vals[0])
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
