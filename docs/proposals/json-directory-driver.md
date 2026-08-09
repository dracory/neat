# JSONDB Driver

**Date**: August 9, 2026
**Status**: Proposal
**Priority**: Medium

## Problem Statement

The `csvdb` driver lets you point at a directory of CSV files and query them as database tables. There is no equivalent for JSON files — users with JSON or JSONL/NDJSON data (API exports, NoSQL dumps, log files, configuration datasets) must still write boilerplate parsing code or use the per-file `NewJsonFileSource` helper for each file individually.

The existing `support/jsonsource` package provides `NewJsonFileSource(filePath)` for single-file use with the array driver, but:
- You must call it once per file and pass each result to `Model()` separately
- There is no "point at a directory and query everything" experience
- Joins across JSON files require manually creating multiple sources and registering them

### Desired Experience

```go
config := neat.DBConfig{
    Default: "json_db",
    Connections: map[string]neat.ConnectionConfig{
        "json_db": {
            Driver:   "jsondb",
            Database: "data/",   // directory path
        },
    },
}

database, _ := neat.New(config)
defer database.Close()

// data/users.json → "users" table
var users []User
err := database.Query().Model(&User{}).Where("active = ?", true).Get(&users)

// data/events.jsonl → "events" table (JSONL auto-detected from extension)
var events []Event
err = database.Query().Model(&Event{}).Where("type = ?", "purchase").Get(&events)

// JOIN across JSON files — works because all tables are in one SQLite DB
var orders []OrderWithUser
err = database.Query().
    Table("orders").
    LeftJoin("users", "orders.user_id = users.id").
    Get(&orders)
```

The directory is the database. Each `.json`, `.jsonl`, or `.ndjson` file is a table. The filename (without the extension) is the table name. Object keys define the columns. No per-file configuration, no `ArraySource` structs, no manual parsing.

## Proposed Solution

Implement a new `JSONDB` driver that mirrors the `CSVDB` driver's architecture:

1. **Embeds `*SQLite`** for all `Driver` interface methods (Open, Close, Ping, BeginTx, Placeholder, Dialect)
2. **Overrides `Open`** to scan the directory for JSON files, parse each one, and populate an in-memory SQLite database with one table per file
3. **Returns `Dialect() == "sqlite"`** so the query builder generates SQLite-compatible SQL and uses SQLite placeholders — no query builder changes needed

The driver is stateless. All state lives in the in-memory SQLite database. When the connection is closed, everything is gone.

### Mental Model

```
Directory (database)          SQLite (in-memory)
┌──────────────────┐         ┌──────────────────┐
│ data/            │         │                  │
│   users.json  ───┼────────▶│ "users" table    │
│   events.jsonl ──┼────────▶│ "events" table   │
│   orders.json ───┼────────▶│ "orders" table   │
└──────────────────┘         └──────────────────┘
```

- `data/users.json` → table `"users"` (filename without extension)
- `.json` files: parsed as a JSON array of objects `[{"id":1,...},...]`
- `.jsonl` / `.ndjson` files: parsed as JSON Lines (one object per line)
- Object keys across all rows define the column schema (union of keys)
- JSON native types are preserved: `int64`, `float64`, `bool`, `string`, `nil`
- Nested objects and arrays are stored as JSON strings (queryable via SQLite JSON functions)
- String values matching RFC3339 are converted to `time.Time` → `DATETIME` columns
- All tables populated at `Open` time (eager, not lazy)

### Architecture

```
database/
  driver/
    jsondb.go              // JSONDB driver (embeds *SQLite, overrides Open)
    jsondb_json.go         // JSON parsing + schema inference helper
    jsondb_test.go         // Driver unit tests
    jsondb_json_test.go    // JSON parser tests

integration_tests/
  jsondb/
    helper.go              // SetupJSONDBTest, JSON fixtures, model types
    jsondb_query_test.go   // Query, WHERE, ORDER BY, LIMIT, First, Find, Count
    jsondb_join_test.go    // LEFT JOIN, INNER JOIN, 3-table JOIN, JOIN+WHERE
    jsondb_aggregate_test.go  // COUNT, SUM, AVG, MIN/MAX, GROUP BY
    jsondb_type_inference_test.go  // Native type preservation, nested objects, JSONL

examples/
  jsondb-driver/
    main.go                // Example usage
    main_test.go
    README.md
    data/
      users.json           // Sample JSON array
      products.json        // Sample JSON array
      orders.jsonl         // Sample JSONL
```

No new contracts. No new interfaces. No changes to `contracts/database/orm/`. No changes to `database/query/`. The driver plugs into the existing driver registration system — same integration points as `csvdb`.

## Driver Implementation

```go
// database/driver/jsondb.go

package driver

import (
    "database/sql"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    _ "modernc.org/sqlite"
)

// JSONDB implements the Driver interface for JSON-directory-backed storage.
// It embeds *SQLite for all standard Driver methods and overrides Open to
// scan a directory of JSON/JSONL files and populate an in-memory SQLite
// database.
//
// The directory path is passed as the DSN (from ConnectionConfig.Database).
// Each .json, .jsonl, or .ndjson file becomes a table named after the
// filename (without the extension). Object keys define the columns.
//
// The driver is stateless — all state lives in the in-memory SQLite database.
type JSONDB struct {
    *SQLite
}

func NewJSONDB() *JSONDB {
    return &JSONDB{SQLite: NewSQLite()}
}

func (j *JSONDB) Dialect() string {
    return "sqlite"
}

func (j *JSONDB) Open(dirPath string) (*sql.DB, error) {
    db, err := sql.Open("sqlite", ":memory:")
    if err != nil {
        return nil, fmt.Errorf("jsondb: failed to open in-memory SQLite: %w", err)
    }

    if dirPath == "" || dirPath == ":memory:" {
        return db, nil
    }

    info, err := os.Stat(dirPath)
    if err != nil {
        db.Close()
        return nil, fmt.Errorf("jsondb: cannot access directory %s: %w", dirPath, err)
    }
    if !info.IsDir() {
        db.Close()
        return nil, fmt.Errorf("jsondb: %s is not a directory", dirPath)
    }

    entries, err := os.ReadDir(dirPath)
    if err != nil {
        db.Close()
        return nil, fmt.Errorf("jsondb: cannot read directory %s: %w", dirPath, err)
    }

    seenTables := make(map[string]string) // lower(tableName) → original filename

    for _, entry := range entries {
        if entry.IsDir() {
            continue
        }
        ext := strings.ToLower(filepath.Ext(entry.Name()))
        if ext != ".json" && ext != ".jsonl" && ext != ".ndjson" {
            continue
        }

        tableName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
        filePath := filepath.Join(dirPath, entry.Name())

        lowerName := strings.ToLower(tableName)
        if prevFile, exists := seenTables[lowerName]; exists {
            db.Close()
            return nil, fmt.Errorf("jsondb: table name collision: %s and %s produce the same table name (case-insensitive)", prevFile, entry.Name())
        }
        seenTables[lowerName] = entry.Name()

        if err := populateJSONFile(db, tableName, filePath); err != nil {
            db.Close()
            return nil, fmt.Errorf("jsondb: failed to populate table %s from %s: %w", tableName, filePath, err)
        }
    }

    return db, nil
}
```

### JSON Parsing and Schema Inference

```go
// database/driver/jsondb_json.go

// MaxJSONRows limits the number of data rows (objects) that can be loaded
// from a single JSON/JSONL file to prevent unbounded memory/CPU consumption.
const MaxJSONRows = 100000

// parseJSONFile reads a JSON or JSONL file and returns rows as []map[string]any.
// Format is auto-detected from the file extension:
//   .json    → JSON array of objects
//   .jsonl   → JSON Lines (one object per line)
//   .ndjson  → JSON Lines (one object per line)
//
// Reads rows incrementally (streaming) so MaxJSONRows is enforced without
// loading the entire file into memory.
func parseJSONFile(filePath string) ([]map[string]any, error) {
    // ...opens file, dispatches to parseJSONArray or parseJSONL based on extension
}

// inferSchema examines all rows and returns:
//   - sorted column names (union of all object keys across all rows)
//   - SQLite type for each column (inferred from native JSON types)
//
// JSON has native types, so inference is simpler than CSV:
//   int64 → INTEGER, float64 → REAL, bool → INTEGER (0/1),
//   time.Time → DATETIME, string → TEXT, nil → skipped (doesn't affect type)
//   nested objects/arrays → TEXT (stored as JSON strings)
//
// Type widening applies when a column has mixed types across rows:
//   INTEGER + REAL → REAL, any incompatible mix → TEXT
func inferSchema(rows []map[string]any) (columns []string, types []string) {
    // 1. Collect union of all keys across all rows
    // 2. For each column, infer type from non-nil values
    // 3. Widen types as needed (INTEGER → REAL → TEXT)
    // 4. Sort columns alphabetically for deterministic ordering
    // 5. Default to TEXT for columns with only nil values
}

// normalizeValue converts a JSON-native value into a form suitable for
// database/sql insertion:
//   float64 with integer value → int64 (JSON has no int type)
//   map[string]any → JSON string (via json.Marshal)
//   []any → JSON string (via json.Marshal)
//   string matching RFC3339 → time.Time
//   everything else → unchanged
func normalizeValue(v any) any { /* ... */ }

// populateJSONFile reads a JSON/JSONL file, infers schema, creates a table,
// and inserts all rows in a transaction. Validates:
// - Table name is a simple identifier (SQL injection prevention)
// - Column names are simple identifiers
// - No duplicate column names (case-insensitive)
// - Row count does not exceed MaxJSONRows
func populateJSONFile(db *sql.DB, tableName string, filePath string) error {
    // 1. parseJSONFile → rows
    // 2. Validate table name, collect & validate column names
    // 3. inferSchema → columns, types
    // 4. CREATE TABLE with inferred schema
    // 5. BEGIN TRANSACTION
    // 6. INSERT rows in batches (SQLite parameter limit ~999)
    // 7. COMMIT
}
```

### Key Difference from CSVDB: Native Types

CSV requires string-to-type inference (`strconv.ParseInt`, `strconv.ParseFloat`, date format guessing). JSON has native types — `encoding/json` already produces `int64` (via `float64` normalization), `float64`, `bool`, `string`, and `nil`. This means:

| Concern | CSVDB | JSONDB |
|---------|-------|--------|
| Type inference | String parsing (int → float → bool → time → string) | Native JSON types (int64, float64, bool, string, nil) |
| Date detection | String format guessing (RFC3339, `2006-01-02`, etc.) | String values matching RFC3339 → `time.Time` |
| Nested data | Not supported (flat rows only) | Nested objects/arrays → JSON strings (queryable via SQLite JSON functions) |
| Missing columns | Ragged rows (fewer fields → NULL) | Missing keys → NULL (natural in JSON) |
| Schema source | First row (header) | Union of all object keys across all rows |
| BOM stripping | Required (Excel exports) | Not needed (JSON files don't have BOM) |
| Format variants | One format (CSV) | Three extensions: `.json`, `.jsonl`, `.ndjson` |

### Type Inference

| JSON value | SQLite type |
|------------|-------------|
| `42` (integer-valued float) | INTEGER |
| `19.99` | REAL |
| `true`, `false` | INTEGER (0/1) |
| `"hello"` | TEXT |
| `"2024-01-15T10:30:00Z"` | DATETIME (RFC3339 → `time.Time`) |
| `{"city": "NYC"}` (nested object) | TEXT (JSON string) |
| `[1, 2, 3]` (nested array) | TEXT (JSON string) |
| `null` | NULL (doesn't affect column type) |

Mixed types in a column widen: INTEGER → REAL → TEXT. A column with `null` and `42` → INTEGER. A column with `42` and `19.99` → REAL. A column with `"hello"` and `42` → TEXT.

### JSONL Format Detection

| Extension | Format | Parsing |
|-----------|--------|---------|
| `.json` | JSON array of objects | `json.Decoder` with `[` token stream |
| `.jsonl` | JSON Lines (one object per line) | `bufio.Scanner` + `json.Unmarshal` per line |
| `.ndjson` | JSON Lines (one object per line) | `bufio.Scanner` + `json.Unmarshal` per line |

JSONL files skip empty lines. The buffer is sized to 1MB per line to handle long records.

## Integration Points

The driver requires one-line additions to the same 6 files touched by `csvdb`. No new contracts, no query builder changes, no new interfaces.

### 1. Driver Constant

```go
// contracts/database/config.go
DriverJSONDB    Driver = "jsondb"
```

### 2. Driver Registration

```go
// database/orm/orm.go — createDriver()
case "jsondb":
    return driver.NewJSONDB()
```

### 3. DSN Builder

```go
// database/db/config_builder.go — BuildDSN()
case "jsondb":
    return b.buildJSONDBDSN()
```

`buildJSONDBDSN` returns `ConnectionConfig.Database` if non-empty, else `:memory:`. For JSONDB, `Database` holds the directory path.

### 4. Config Validation

```go
// database/db/config_builder.go — ConnectionConfig.Validate()
case "sqlite", "array", "csvdb", "jsondb":
    // database path is optional; empty defaults to :memory:
    // For JSONDB, the Database field holds the directory path.
    return nil
```

### 5. Connection Pool Configuration

```go
// database/orm/orm.go — configureConnectionPool()
pinSingleConn := connConfig.Driver == "sqlite" || connConfig.Driver == "array" || connConfig.Driver == "csvdb" || connConfig.Driver == "jsondb"
```

JSONDB uses in-memory SQLite, so it shares the same single-connection constraint.

### 6. Database Name Detection

```go
// database/db.go — detectDatabaseName()
case "sqlite", "turso", "array", "csvdb", "jsondb":
    return "main"
```

### 7. Schema Builder

```go
// database/schema/schema.go — New()
case contractsdatabase.DriverJSONDB:
    // JSONDB uses SQLite grammar since Dialect() returns "sqlite".
    sqliteGrammar := grammars.NewSqlite(log, prefix)
    driverSchema = NewSqliteSchema(sqliteGrammar, orm, prefix)
    grammar = sqliteGrammar
    processor = processors.NewSqlite()
```

## Robustness Features

All robustness features from `csvdb` carry over to `jsondb`, adapted for JSON:

- **Table name collision detection**: Files that produce colliding table names (e.g., `Users.json` and `users.json` on case-sensitive filesystems) are detected and rejected.
- **MaxJSONRows limit**: 100,000-row limit per JSON/JSONL file to prevent unbounded memory/CPU consumption. Rows are read incrementally (streaming via `json.Decoder` for arrays, `bufio.Scanner` for JSONL) so the limit is enforced without loading the entire file.
- **Duplicate column detection**: If two object keys differ only in case (e.g., `"Name"` and `"name"`), this is detected as a duplicate since SQLite column names are case-insensitive.
- **Invalid identifier detection**: Table names derived from filenames and column names from object keys are validated with `isSimpleIdentifier` to prevent SQL injection.
- **Transaction wrapping**: Batch INSERTs are wrapped in a `BEGIN`/`COMMIT` transaction for atomicity and performance.
- **Empty file handling**: A JSON file with `[]` (empty array) or a JSONL file with only empty lines produces an empty table (CREATE TABLE with no rows, schema inferred as a single `TEXT` placeholder column or no columns — see Open Questions).
- **Trailing data rejection**: JSON array parsing rejects trailing data after the closing `]` (e.g., `][]`).

## Testing

### Unit Tests (`database/driver/`)
- JSON array parsing, JSONL parsing, NDJSON parsing
- Schema inference: native types, type widening, union of keys, nil-only columns
- Nested objects → JSON strings, nested arrays → JSON strings
- RFC3339 string → `time.Time` conversion
- Integer-valued float → `int64` normalization
- Table name collision, case-insensitive extension matching
- MaxJSONRows limit enforcement (streaming)
- Invalid identifiers, duplicate columns (case-insensitive)
- Empty array, empty JSONL file, non-JSON/JSONL files skipped
- Transaction wrapping, batch insertion

### Integration Tests (`integration_tests/jsondb/`)
- **Query**: basic query, WHERE equals, WHERE bool, WHERE on nested JSON string, ORDER BY asc/desc, LIMIT, LIMIT+OFFSET, First, Find, Count
- **JOIN**: LEFT JOIN, INNER JOIN, 3-table JOIN (JSON + JSONL), JOIN with WHERE
- **Aggregates**: COUNT, COUNT with WHERE, SUM, AVG, MIN/MAX, GROUP BY, SUM with JOIN+GROUP BY
- **Type inference**: integer, float, bool, string, datetime (RFC3339), nested object as JSON string, NULL handling, JSONL format

### Example (`examples/jsondb-driver/`)
- Working example with sample JSON and JSONL data (users.json, products.json, orders.jsonl)
- Demonstrates WHERE, JOIN (across JSON and JSONL files), and aggregate queries
- Shows nested object querying via SQLite JSON functions

## Open Questions

### 1. Empty JSON Array — Schema Inference

A JSON file containing `[]` has no objects to infer column names from. Options:
- **A**: Create a table with no columns (`CREATE TABLE "name" ()`) — SQLite allows this but it's unusual
- **B**: Skip the file entirely (no table created) — consistent with "no data = no table"
- **C**: Return an error — "cannot infer schema from empty JSON array"

**Recommendation**: Option B (skip the file). This is consistent with the `csvdb` behavior where a CSV with only a header row and no data rows still creates a table (because the header defines columns). But JSON has no header — an empty array has no schema information. Skipping is the least surprising behavior. A warning could be logged.

### 2. Heterogeneous Objects — Conflicting Key Types

If `users.json` contains `[{"id": 1, "name": "Alice"}, {"id": "x", "name": "Bob"}]`, the `id` column has mixed `int64` and `string` values. The type widens to TEXT, and the integer `1` is stored as `"1"`. This is consistent with `csvdb`'s widening behavior. No special handling needed — just documented.

### 3. Column Name Collisions with JSON Nested Keys

If an object has both `"name"` and `"Name"` as keys, they collide in SQLite (case-insensitive column names). The driver should detect this and return an error: `duplicate column name (case-insensitive): name`. This is the same approach as `csvdb`'s duplicate column detection, extended to case-insensitive comparison.

### 4. Reuse of `support/jsonsource` Package

The existing `support/jsonsource` package has `parseJSONArrayReader`, `parseJSONLReader`, `normalizeValue`, and `normalizeRows` functions. The `jsondb` driver could either:
- **A**: Reuse `support/jsonsource` functions directly (import the package)
- **B**: Copy the logic into `database/driver/jsondb_json.go` (self-contained, like `csvdb` copied CSV logic)

**Recommendation**: Option A (reuse). The `support/jsonsource` functions are already tested and production-quality. The `csvdb` driver copied CSV logic because there was no existing `support/csvsource` parsing function that returned `[]string` rows (the existing one returns `*arraysource.Model`). For JSON, the parsing functions return `[]map[string]any` which is exactly what `jsondb` needs. However, `support/jsonsource` panics on parse errors (it's designed for the `Model()` API where panics are acceptable). The `jsondb` driver needs errors, not panics. So either:
- Refactor `support/jsonsource` to return errors instead of panicking (breaking change to the public API)
- Extract the internal parsing functions (`parseJSONArrayReader`, `parseJSONLReader`, `normalizeValue`) into a shared internal package or make them exported
- Copy the logic (Option B)

**Final recommendation**: Extract the parsing functions into `support/jsonsource` as exported functions that return errors (`ParseJSONArray`, `ParseJSONL`, `NormalizeValue`), and have the existing `NewJsonSource` / `NewJsonFileSource` wrappers call them and panic on error. This avoids code duplication while preserving the existing API.

## Relationship to Existing Work

| Component | Relationship |
|-----------|-------------|
| `csvdb` driver | Architectural template — same embedding pattern, same integration points, same robustness features |
| `support/jsonsource` | Parsing logic to reuse (or extract from) — `parseJSONArrayReader`, `parseJSONLReader`, `normalizeValue` |
| `support/arraysource` | Not used — `jsondb` populates SQLite directly via `database/sql`, not via the array driver's `Model()` hook |
| `flat-file-driver.md` proposal | The v3 revision notes that "directory-as-database mode is covered by a separate proposal" — this is that proposal for JSON (alongside `csv-directory-driver.md` for CSV) |

## Future Extensions

- **YAMLDB**: Same pattern for `.yaml`/`.yml` files (requires a YAML parser dependency)
- **XMLDB**: Same pattern for `.xml` files (reuses `support/xmlsource` parsing)
- **Mixed-format directory**: A single `filedb` driver that handles all extensions (`.csv`, `.json`, `.jsonl`, `.xml`) in one directory — would unify `csvdb` and `jsondb` into one driver
- **Streaming queries**: For directories with many large files, populate tables lazily (on first query) instead of eagerly at `Open` time
- **Persistent caching**: Save the populated SQLite database to a file on disk and reuse it on subsequent runs if the source files haven't changed (mirrors the `array-driver-enhancement` proposal's `ArrayCache` interface)
