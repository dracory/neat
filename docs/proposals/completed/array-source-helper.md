# Proposal: Array Source Helper with `NewArraySourceFrom[T]`

**Date**: August 8, 2026
**Status**: Completed
**Priority**: Medium

## Problem

Today, to use the `array` driver, users **must** define a custom struct implementing the `ArraySource` interface (`TableName()` + `Rows()`), and hand-write `map[string]any` literals. This is boilerplate-heavy compared to Laravel's `collect([...])` ergonomics.

**Current usage (15+ lines of boilerplate):**

```go
type StatusSource struct{}

func (s *StatusSource) TableName() string { return "statuses" }

func (s *StatusSource) Rows() ([]map[string]any, error) {
    return []map[string]any{
        {"id": 1, "name": "Pending", "color": "yellow"},
        {"id": 2, "name": "Active", "color": "green"},
    }, nil
}

// ... then:
database.Query().Model(&StatusSource{}).Get(&statuses)
```

## Goal

Let users pass **a plain Go slice of structs** — no custom `ArraySource` struct, no `map[string]any` literals, no table name. One function, one line.

**Primary target usage:**

```go
statuses := []Status{
    {ID: 1, Name: "Pending", Color: "yellow"},
    {ID: 2, Name: "Active", Color: "green"},
}

database.Query().Model(neat.NewArraySourceFrom(statuses)).OrderBy("id", "asc").Get(&out)
```

**Secondary usage (map-based, for dynamic rows):**

```go
database.Query().Model(neat.NewArraySource([]map[string]any{
    {"id": 1, "name": "Pending", "color": "yellow"},
})).Get(&out)
```

## Design

### 1. `NewArraySourceFrom[T any]` — the primary entry point

A generic function that accepts any slice — structs or `map[string]any` — and returns an `*Model`. Panics on unsupported types (programmer error, same convention as `regexp.MustCompile` which is used 16 times across this codebase).

**File:** `support/arraysource/from_slice.go` (new)

```go
package arraysource

import (
    "fmt"
    "reflect"
    "sync/atomic"

    "github.com/dracory/neat/support/structref"
)

// NewArraySourceFrom creates an array-backed data source from a slice of
// structs or []map[string]any, with zero setup: no custom struct, no table
// name, no schema. This is the primary, day-to-day entry point for the
// array driver — the third constructor in the NewArraySource family.
//
//	statuses := []Status{{ID: 1, Name: "Pending"}, {ID: 2, Name: "Active"}}
//	database.Query().Model(neat.NewArraySourceFrom(statuses)).Get(&out)
//
// For struct slices, field names are resolved to column names using the same
// tag convention as the ORM (db > neat > gorm > snake_case). Embedded structs
// are flattened. Association fields (slices, struct pointers) are skipped.
// Nullable pointer fields (*string, *int) are dereferenced; nil pointers
// become nil (NULL in SQLite). time.Time fields are included as-is.
//
// For []map[string]any, rows are shallow-copied (snapshot at call time).
//
// Panics if T is not a struct, pointer-to-struct, or map[string]any — this
// is a programmer error, not a runtime condition.
// Panics if the slice is empty — use NewWithSchema for empty datasets.
func NewArraySourceFrom[T any](items []T) *Model {
    if len(items) == 0 {
        panic("arraysource: NewArraySourceFrom() requires non-empty items; use NewWithSchema() for empty datasets")
    }

    // Type assertion on the slice — avoids reflect.TypeOf(zero) which panics
    // on nil interface zero values
    if rows, ok := any(items).([]map[string]any); ok {
        return &Model{
            table: nextTableName[T](),
            data:  copyRows(rows),
        }
    }

    rows := structsToRows(items)
    if len(rows) == 0 {
        panic("arraysource: NewArraySourceFrom() produced no rows; check that the struct has exported fields")
    }

    return &Model{
        table: nextTableName[T](),
        data:  rows,
    }
}

// copyRows creates a shallow copy of each map so the source is a snapshot
// at call time — mutating the caller's maps after NewArraySourceFrom doesn't
// affect the array source. Both the struct path and map path now have the
// same "snapshot at call time" contract.
func copyRows(rows []map[string]any) []map[string]any {
    out := make([]map[string]any, len(rows))
    for i, row := range rows {
        m := make(map[string]any, len(row))
        for k, v := range row {
            m[k] = v
        }
        out[i] = m
    }
    return out
}
```

### 2. `structsToRows` — struct→map conversion via shared reflection

This is the critical piece. It **reuses the ORM's existing `structFieldColumnName` logic** — not a second implementation — by extracting that function to `support/structref/`.

**File:** `support/structref/structref.go` (new — extracted from `database/query/utils.go`)

```go
package structref

import (
    "reflect"
    "strings"
)

// FieldColumnName returns the column name for a struct field by checking
// db, neat, gorm tags (in that order), then falling back to snake_case
// of the field name.
//
// This is the single source of truth for struct→column name resolution,
// shared by the query builder and the array source helper.
func FieldColumnName(f reflect.StructField) string {
    for _, tag := range []string{"db", "neat", "gorm"} {
        if v := f.Tag.Get(tag); v != "" && v != "-" {
            parts := strings.SplitN(v, ";", 2)
            if len(parts) == 0 {
                continue
            }
            part := parts[0]
            if strings.HasPrefix(part, "column:") {
                return strings.TrimPrefix(part, "column:")
            }
            if tag == "db" || tag == "neat" {
                return part
            }
        }
    }
    return CamelToSnake(f.Name)
}

// CamelToSnake converts CamelCase to snake_case.
func CamelToSnake(s string) string {
    // ... (moved verbatim from database/query/utils.go)
}
```

**Refactor in `database/query/utils.go`:** Replace the local `structFieldColumnName` and `camelToSnake` with calls to `structref.FieldColumnName` and `structref.CamelToSnake`. This is a mechanical rename — behavior is identical. The existing `export_test.go` shim continues to work (it can delegate to `structref` or be removed in favor of direct `structref` tests).

**`structsToRows` implementation:**

```go
func structsToRows[T any](items []T) []map[string]any {
    rows := make([]map[string]any, 0, len(items))
    for i := range items {
        v := reflect.ValueOf(items[i])
        if v.Kind() == reflect.Pointer {
            if v.IsNil() {
                continue // skip nil elements
            }
            v = v.Elem()
        }
        if v.Kind() != reflect.Struct {
            panic(fmt.Sprintf("arraysource: NewArraySourceFrom requires struct or map[string]any, got %s", v.Kind()))
        }
        rows = append(rows, structToMap(v))
    }
    return rows
}

func structToMap(v reflect.Value) map[string]any {
    row := make(map[string]any)
    t := v.Type()
    for i := 0; i < v.NumField(); i++ {
        field := t.Field(i)
        fieldValue := v.Field(i)

        // Flatten embedded structs
        if field.Anonymous && fieldValue.Kind() == reflect.Struct {
            for k, val := range structToMap(fieldValue) {
                row[k] = val
            }
            continue
        }

        // Skip unexported fields
        if !fieldValue.CanInterface() {
            continue
        }

        // Skip association fields — mirrors builder_extract.go logic:
        // - Slices (except []byte, json.RawMessage) → associations
        // - Structs (except time.Time) → associations
        // - Pointers to structs → associations
        // - Pointers to basic types (*string, *int) → KEEP, dereference
        if isAssociationField(fieldValue) {
            continue
        }

        col := structref.FieldColumnName(field)
        if col == "" || col == "-" {
            continue
        }

        // Dereference non-nil pointers to basic types
        // Nil pointers → nil (becomes NULL in SQLite)
        if fieldValue.Kind() == reflect.Pointer {
            if fieldValue.IsNil() {
                row[col] = nil
            } else {
                row[col] = fieldValue.Elem().Interface()
            }
            continue
        }

        row[col] = fieldValue.Interface()
    }
    return row
}

// isAssociationField returns true for fields that represent associations
// rather than scalar columns. Mirrors the logic in
// database/query/builder_extract.go:155-176.
func isAssociationField(fieldValue reflect.Value) bool {
    switch fieldValue.Kind() {
    case reflect.Slice:
        // []byte and json.RawMessage are scalar-like, keep them
        if fieldValue.Type() == reflect.TypeOf([]byte(nil)) ||
            fieldValue.Type() == reflect.TypeOf(json.RawMessage(nil)) {
            return false
        }
        return true
    case reflect.Struct:
        // time.Time is a scalar column type, keep it
        if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
            return false
        }
        return true
    case reflect.Pointer:
        if fieldValue.IsNil() {
            // Nil pointer: skip if it points to a struct (association),
            // keep if it points to a basic type (nullable column)
            return fieldValue.Type().Elem().Kind() == reflect.Struct
        }
        // Non-nil pointer: skip if it points to a struct (association),
        // keep if it points to a basic type (nullable column)
        return fieldValue.Elem().Kind() == reflect.Struct
    default:
        return false
    }
}
```

**Nullable pointer handling** (addresses review point #4):
- `*string`, `*int`, `*float64`, etc. → dereferenced; nil → `nil` (NULL in SQLite)
- `*sql.NullString`, `*sql.NullInt64`, etc. → dereferenced; nil → `nil`
- `*time.Time` → dereferenced; nil → `nil`
- `*SomeStruct` (pointer to struct) → skipped (association)

This mirrors `builder_extract.go:155-176` which includes nil basic-type pointers as NULL and skips struct pointers as associations.

**Why this is safe and not a second implementation:**
- Column name resolution is delegated to `structref.FieldColumnName` — the exact same function the query builder uses.
- The field-skipping logic (associations, embedded structs, `time.Time` special case, nullable pointer handling) mirrors `extractStructColumnsAndValues` in `builder_extract.go`, but is intentionally simpler: for `NewArraySourceFrom`, we include ALL basic fields with their actual values (no zero-value skipping, no auto-increment ID skipping, no dialect-specific time handling). The user is providing static data — every value is intentional.

### 3. Type-name-based table naming

Instead of opaque `array_sa4rc789wxg`, derive a readable prefix from the element type:

```go
var arrayCounter uint64

func nextTableName[T any]() string {
    n := atomic.AddUint64(&arrayCounter, 1)
    var name string
    t := reflect.TypeFor[T]()
    if t != nil {
        // Handle pointer types: []*Status → "status", not ""
        if t.Kind() == reflect.Pointer {
            t = t.Elem()
        }
        // Handle slice types (shouldn't happen in practice, but be safe)
        if t.Kind() == reflect.Slice {
            t = t.Elem()
            if t.Kind() == reflect.Pointer {
                t = t.Elem()
            }
        }
        name = strings.ToLower(t.Name())
    }
    if name == "" {
        name = "array"
    }
    return fmt.Sprintf("array_%s_%d", name, n)
}
```

Result: `array_status_1`, `array_status_2`, `array_map_3` (for `map[string]any` slices). `[]*Status` also produces `array_status_4` (pointer unwrapped). Nicer in error messages and SQL debugging, reduces the need for `.Table()`.

### 4. `Model` struct

**File:** `support/arraysource/model.go` (new)

```go
package arraysource

type Model struct {
    table  string
    data   []map[string]any
    schema map[string]string
}

func (m *Model) TableName() string              { return m.table }
func (m *Model) Rows() ([]map[string]any, error) { return m.data, nil }
func (m *Model) Schema() map[string]string       { return m.schema }

// Table sets a custom table name. Must be called before passing to Model().
// Not safe to call concurrently with a query using this source.
func (m *Model) Table(name string) *Model {
    m.table = name
    return m
}
```

### 5. Map-based constructors (lower-level API)

**File:** `support/arraysource/model.go` (same file)

```go
// New creates an ArraySource from []map[string]any rows. Table name is
// auto-generated. Rows are shallow-copied (snapshot at call time).
// Panics if rows is nil or empty — use NewWithSchema for empty datasets.
func New(rows []map[string]any) *Model {
    if len(rows) == 0 {
        panic("arraysource: New() requires non-empty rows; use NewWithSchema() for empty datasets")
    }
    return &Model{
        table:  "array_map_" + strconv.FormatUint(atomic.AddUint64(&arrayCounter, 1), 10),
        data:   copyRows(rows),
    }
}

// NewWithSchema creates an ArraySource with an explicit column schema.
// Rows are shallow-copied (snapshot at call time).
func NewWithSchema(rows []map[string]any, schema map[string]string) *Model {
    return &Model{
        table:  "array_map_" + strconv.FormatUint(atomic.AddUint64(&arrayCounter, 1), 10),
        data:   copyRows(rows),
        schema: schema,
    }
}
```

### 6. Re-export from root `neat` package

**File:** `array_source.go` (new, root `neat` package)

```go
package neat

import (
    "github.com/dracory/neat/support/arraysource"
)

type ArraySourceModel = arraysource.Model

// NewArraySourceFrom creates an array-backed data source from a slice of
// structs or map[string]any. This is the primary entry point for array-backed
// queries — the third constructor in the NewArraySource family.
func NewArraySourceFrom[T any](items []T) *arraysource.Model {
    return arraysource.NewArraySourceFrom(items)
}

func NewArraySource(rows []map[string]any) *arraysource.Model {
    return arraysource.New(rows)
}

func NewArraySourceWithSchema(rows []map[string]any, schema map[string]string) *arraysource.Model {
    return arraysource.NewWithSchema(rows, schema)
}
```

## API Summary

| Function | Args | When |
|----------|------|------|
| `neat.NewArraySourceFrom(structSlice)` | `[]Status`, `[]Product`, etc. | **Daily default** — one line, no ceremony |
| `neat.NewArraySourceFrom([]map[string]any{...})` | map slice | Dynamic rows, still one function |
| `neat.NewArraySource(rows).Table("x")` | `[]map[string]any` + chained setter | Need a specific table name (JOINs, raw SQL) |
| `neat.NewArraySourceWithSchema(rows, schema)` | rows + schema | Empty data or ambiguous types |
| Custom struct implementing `ArraySource` | struct + 2 methods | Legacy / full control |

## Panic vs Error Convention

`NewArraySourceFrom` and `New` panic on misuse (empty input, unsupported type). This is consistent with the codebase's existing convention:
- `regexp.MustCompile` is used 16 times across `support/str`, `database/query`, `database/schema`, `database/association` — all panic on bad input known at authorship time.
- No error-returning `Must`-style helpers exist in the library.

The distinction: **wrong type** (`NewArraySourceFrom([]int{...})`) and **empty input without schema** are programmer errors, not runtime conditions. They surface immediately at construction with a clear message, not deep inside the array driver with a cryptic error.

If empty input can legitimately arise at runtime (e.g., loading from a config file that happens to be empty), the user should check before calling `NewArraySourceFrom` and use `NewWithSchema` with an explicit schema for the empty case.

## Package Layout

```
support/
  structref/
    structref.go          (new — extracted from database/query/utils.go)
    structref_test.go     (new — tests for FieldColumnName, CamelToSnake)
  arraysource/
    model.go              (new — Model struct, New, NewWithSchema, Table, copyRows)
    from_slice.go         (new — NewArraySourceFrom[T], structsToRows, structToMap, isAssociationField, nextTableName)
    from_slice_test.go    (new — unit tests for NewArraySourceFrom with structs + maps)
    model_test.go         (new — unit tests for New, NewWithSchema, Table)
    concurrent_test.go    (new — concurrent unique-name test)
database/
  query/
    utils.go              (modified — delegate to support/structref)
    utils_test.go         (unchanged — tests still pass via delegation)
    export_test.go        (modified — delegate to structref or remove)
array_source.go           (new — root neat package re-exports)
```

## Sequencing: Two PRs

### PR 1: `structref` extraction (refactor, no new features)

The one piece that touches a live, well-tested production code path. Land it independently:

1. Write characterization tests against the **current** `structFieldColumnName` and `camelToSnake` behavior (before moving anything) — verify the `db > neat > gorm` tag priority and acronym-aware snake_case are what we think they are.
2. Extract to `support/structref/`.
3. Update `database/query/utils.go` to delegate.
4. Run `go test ./database/query/...` — all existing tests must pass.
5. Merge.

This keeps the higher-risk refactor small and independently revertible.

### PR 2: `NewArraySourceFrom` / `arraysource` (purely additive)

Build on top of the merged `structref` package:

1. Implement `support/arraysource/` (model.go, from_slice.go).
2. Implement `array_source.go` (root neat re-exports).
3. Update examples and docs.
4. Run full test suite.
5. Merge.

## Files to Change

### PR 1 (structref extraction)

| File | Action |
|------|--------|
| `support/structref/structref.go` (new) | Extract `FieldColumnName` + `CamelToSnake` from `database/query/utils.go` |
| `support/structref/structref_test.go` (new) | Characterization tests (verify existing behavior before codifying) |
| `database/query/utils.go` (modified) | Replace local `structFieldColumnName`/`camelToSnake` with `structref` calls |
| `database/query/export_test.go` (modified) | Delegate to `structref` or remove |

### PR 2 (NewArraySourceFrom + arraysource)

| File | Action |
|------|--------|
| `support/arraysource/model.go` (new) | `Model` struct, `New()`, `NewWithSchema()`, `Table()`, `copyRows()` |
| `support/arraysource/from_slice.go` (new) | `NewArraySourceFrom[T]`, `structsToRows`, `structToMap`, `isAssociationField`, `nextTableName[T]` |
| `support/arraysource/*_test.go` (new) | Unit + concurrent + integration tests |
| `array_source.go` (new, root `neat` package) | Re-export `NewArraySourceFrom`, `NewArraySource`, `NewArraySourceWithSchema`, `ArraySourceModel` |
| `examples/array-driver/main.go` | Add `NewArraySourceFrom` example (keep old struct example for comparison) |
| `examples/array-driver/main_test.go` | Ensure existing test still passes |
| `docs/models.html` | Document `NewArraySourceFrom` as the primary array-source API (hand-authored HTML) |
| `docs/api-reference.html` | Add `NewArraySourceFrom`, `NewArraySource`, `NewArraySourceWithSchema` to API reference (hand-authored HTML) |

## What Does NOT Change

- `contracts/database/orm/array_source.go` — interfaces stay as-is (per ADR-002)
- `ArraySource` / `ArraySchema` / `ArrayPopulator` interfaces — unchanged
- `database/driver/array.go` — no changes needed (already handles nil schema via inference)
- `database/query/query_model.go` — no changes needed (already does type assertion on `ArraySource`)
- `database/query/builder_extract.go` — no changes needed (uses `structFieldColumnName` which now delegates to `structref`)

## Test Coverage

### PR 1: structref characterization tests

1. `TestFieldColumnName_DbTag` — `db:"custom_name"` → `"custom_name"`
2. `TestFieldColumnName_NeatTag` — `neat:"custom_name"` → `"custom_name"`
3. `TestFieldColumnName_GormTag` — `gorm:"column:custom_name"` → `"custom_name"`
4. `TestFieldColumnName_TagPriority` — `db:"a" neat:"b" gorm:"column:c"` → `"a"` (db wins)
5. `TestFieldColumnName_DashOptOut` — `db:"-"` → `"-"` (skip)
6. `TestFieldColumnName_Fallback` — no tag → snake_case of field name
7. `TestCamelToSnake` — various inputs including acronyms (`HTTPServer` → `http_server`)
8. `TestCamelToSnake_Simple` — `UserName` → `user_name`

### PR 2: arraysource unit tests

9. `TestNewArraySourceFrom_Structs` — `NewArraySourceFrom([]Status{...})` produces correct `[]map[string]any` with snake_case keys
10. `TestNewArraySourceFrom_StructsWithTags` — struct with `db:"custom_name"` tag uses the tag value
11. `TestNewArraySourceFrom_EmbeddedStructs` — embedded struct fields are flattened
12. `TestNewArraySourceFrom_SkipsAssociations` — slice fields and struct pointers are excluded
13. `TestNewArraySourceFrom_TimeTime` — `time.Time` fields are included as-is
14. `TestNewArraySourceFrom_NullablePointer_NonNil` — `*string` with value → dereferenced string
15. `TestNewArraySourceFrom_NullablePointer_Nil` — `*string` nil → `nil` (NULL)
16. `TestNewArraySourceFrom_NullablePointer_ToStruct` — `*SomeStruct` → skipped (association)
17. `TestNewArraySourceFrom_MapSlice` — `NewArraySourceFrom([]map[string]any{...})` passes through with shallow copy
18. `TestNewArraySourceFrom_MapSlice_SnapshotSemantics` — mutating caller's map after NewArraySourceFrom doesn't affect the source
19. `TestNewArraySourceFrom_PointerSlice` — `NewArraySourceFrom([]*Status{...})` works, table name is `array_status_N`
20. `TestNewArraySourceFrom_TableName` — auto-generated name contains type name (e.g. `array_status_1`)
21. `TestNewArraySourceFrom_PanicsOnEmpty` — `NewArraySourceFrom([]Status{})` panics
22. `TestNewArraySourceFrom_PanicsOnUnsupportedType` — `NewArraySourceFrom([]int{1,2})` panics
23. `TestNew_PanicsOnEmpty` — `New(nil)` and `New([]map[string]any{})` panic
24. `TestNew_ShallowCopiesRows` — mutating caller's map after New doesn't affect the source
25. `TestTable_Setter` — `.Table("custom")` overrides auto-generated name
26. `TestNewArraySourceFrom_UniqueNames_Concurrent` — 100 goroutines calling `NewArraySourceFrom` produce 100 unique table names

### PR 2: integration tests

27. End-to-end: `database.Query().Model(neat.NewArraySourceFrom(statuses)).OrderBy("id","asc").Get(&out)` — full path
28. End-to-end with `.Table()`: same but with custom table name
29. End-to-end with `neat.NewArraySourceFrom([]map[string]any{...})`: map slice path
30. End-to-end with `neat.NewArraySourceWithSchema(rows, schema)` and empty rows
31. End-to-end with nullable pointer fields: `NewArraySourceFrom([]UserWithNullableFields{...})` → query → scan back
32. Query builder tests still pass after `structref` extraction (regression check — already covered by PR 1)

## Backward Compatibility

Fully backward compatible:
- Existing custom structs implementing `ArraySource` continue to work unchanged
- The `structref` extraction is a mechanical rename — `database/query/utils.go` delegates to the same logic
- All existing `database/query/` tests continue to pass (verified by PR 1 regression tests)
- The helper is purely additive — new packages, new file in root, minimal modification to `query/utils.go` (function body swap, same behavior)

## Table Registry Lifecycle

The array driver maintains a `sync.Map` (`populated`) keyed by `dbPointer-tableName`. Entries persist until `Cleanup()` is called (via `Database.Close()`). There is no per-table eviction.

**Implication for auto-generated names:** Each `NewArraySourceFrom()` / `New()` call produces a unique table name, so each call creates a new SQLite table and a new `sync.Map` entry. In a long-running service that calls `NewArraySourceFrom` per-request, these accumulate until `Database.Close()`.

**This is acceptable because:**
- The array driver is designed for **static reference data** (statuses, config tables, enums), not per-request dynamic data. Per-request usage is an anti-pattern regardless of this helper.
- The `sync.Map` entries are tiny (a string key → `true`).
- The SQLite tables are in-memory (default `:memory:` DSN), so they're freed when the connection closes.
- The existing [array-driver-enhancement proposal](array-driver-enhancement.md) already discusses persistent caching for heavier workloads.

**Caveat:** This reasoning assumes the in-memory `:memory:` DSN. If the array driver is ever pointed at a persistent backing store (file-based SQLite), auto-generated tables would accumulate on disk and not be cleaned up. The doc comment on `NewArraySourceFrom()` will note this.

**Documentation:** This will be noted in the doc comment on `NewArraySourceFrom()` and in `docs/models.html`.

## Verification

### PR 1

1. `go build ./...` — compiles
2. `go test ./support/structref/...` — extracted reflection utils pass
3. `go test ./database/query/...` — regression: all existing query builder tests pass

### PR 2

4. `go build ./...` — compiles
5. `go test ./support/arraysource/...` — unit + concurrent tests pass
6. `go test ./examples/array-driver/...` — example works (old + new)
7. `go test ./database/...` — integration tests pass
8. `go test ./...` — full test suite passes
