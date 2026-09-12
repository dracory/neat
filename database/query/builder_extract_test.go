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

// TestExtractStructColumnNamesWithNamedTimeWrapper reproduces a bug where
// a named struct field that wraps a single time.Time (e.g. orm.CreatedAt,
// orm.UpdatedAt) is excluded from the extracted column list. These
// wrapper structs have a json tag on the inner field but no db tag on the
// outer field, so extractStructColumnNames skips them because the field
// type is a struct (not time.Time). This causes SELECT queries built via
// Model() to omit created_at/updated_at columns, leaving them at the Go
// zero value after scanning.
func TestExtractStructColumnNamesWithNamedTimeWrapper(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	b := NewBuilder(q)

	// Mirrors orm.CreatedAt / orm.UpdatedAt from database/orm/model.go
	type CreatedAtWrapper struct {
		CreatedAt time.Time `json:"created_at"`
	}
	type UpdatedAtWrapper struct {
		UpdatedAt time.Time `json:"updated_at"`
	}

	type User struct {
		Name    string           `db:"name"`
		Email   string           `db:"email"`
		Created CreatedAtWrapper `json:"created_at"`
		Updated UpdatedAtWrapper `json:"updated_at"`
	}

	user := User{Name: "Alice", Email: "alice@example.com"}
	cols := b.extractStructColumnNames(reflect.ValueOf(user))

	// Should extract 4 columns: name, email, created_at, updated_at
	if len(cols) != 4 {
		t.Errorf("Expected 4 columns (name, email, created_at, updated_at), got %d: %v", len(cols), cols)
	}

	// Verify created_at and updated_at are present
	foundCreatedAt := false
	foundUpdatedAt := false
	for _, col := range cols {
		if col == "created_at" {
			foundCreatedAt = true
		}
		if col == "updated_at" {
			foundUpdatedAt = true
		}
	}
	if !foundCreatedAt {
		t.Errorf("Expected 'created_at' in columns, got %v", cols)
	}
	if !foundUpdatedAt {
		t.Errorf("Expected 'updated_at' in columns, got %v", cols)
	}
}

// TestExtractStructColumnsAndValuesWithNamedTimeWrapper reproduces Finding 2:
// the INSERT path (extractStructColumnsAndValues) skips named struct fields
// that wrap a single time.Time (e.g., orm.CreatedAt, orm.UpdatedAt), while
// the SELECT path (extractStructColumnNames via unwrapTimeColumn) now
// includes them. This creates an INSERT/SELECT asymmetry: a struct-based
// Create(&T{...}) omits created_at/updated_at, silently dropping any
// timestamp the caller set on the wrapper before insert.
func TestExtractStructColumnsAndValuesWithNamedTimeWrapper(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	b := NewBuilder(q)

	// Mirrors orm.CreatedAt from database/orm/model.go
	type CreatedAtWrapper struct {
		CreatedAt time.Time `json:"created_at"`
	}

	type User struct {
		Name    string           `db:"name"`
		Created CreatedAtWrapper `json:"created_at"`
	}

	want := time.Date(2026, 9, 12, 10, 0, 0, 0, time.UTC)
	user := User{
		Name:    "Alice",
		Created: CreatedAtWrapper{CreatedAt: want},
	}

	cols, vals := b.extractStructColumnsAndValues(reflect.ValueOf(user))

	// Should extract 2 columns: name, created_at
	if len(cols) != 2 {
		t.Fatalf("Expected 2 columns (name, created_at), got %d: %v", len(cols), cols)
	}
	if len(vals) != 2 {
		t.Fatalf("Expected 2 values, got %d: %v", len(vals), vals)
	}

	// Verify created_at is present
	foundCreatedAt := false
	var createdAtVal any
	for i, col := range cols {
		if col == "created_at" {
			foundCreatedAt = true
			createdAtVal = vals[i]
		}
	}
	if !foundCreatedAt {
		t.Fatalf("Expected 'created_at' in columns, got %v", cols)
	}

	// Verify the value is the inner time.Time (or convertible to it)
	tv, ok := createdAtVal.(time.Time)
	if !ok {
		t.Fatalf("Expected created_at value to be time.Time, got %T (%v)", createdAtVal, createdAtVal)
	}
	if !tv.Equal(want) {
		t.Errorf("Expected created_at value %v, got %v", want, tv)
	}
}

// TestUnwrapTimeColumnRespectsJsonTag reproduces Finding 3: the comment on
// extractStructColumnNames says "Extract the column name from the inner
// field's db/json tag", but structFieldColumnName (via
// structref.FieldColumnName) only checks db/neat/gorm tags — never json.
// The existing tests pass by coincidence because the inner field is named
// CreatedAt, which snake-cases to "created_at" matching the json tag.
//
// This test uses an inner field named "When" with json:"created_at" to
// prove the json tag is ignored: unwrapTimeColumn returns "when" (from
// the field name) instead of "created_at" (from the json tag).
func TestUnwrapTimeColumnRespectsJsonTag(t *testing.T) {
	// Inner field named "When" — CamelToSnake("When") = "when"
	// json tag says "created_at" — the column name should come from the tag
	type CreatedAtWrapper struct {
		When time.Time `json:"created_at"`
	}

	col, ok := unwrapTimeColumn(reflect.TypeOf(CreatedAtWrapper{}))
	if !ok {
		t.Fatal("Expected unwrapTimeColumn to detect the wrapper")
	}
	if col != "created_at" {
		t.Errorf("Expected column %q (from json tag), got %q (from field name) "+
			"— json tags are not checked by structFieldColumnName",
			"created_at", col)
	}
}
