# GODB Driver

**Date**: August 9, 2026
**Status**: Proposal
**Priority**: Medium

## Problem Statement

The `csvdb`, `jsondb`, and `xmldb` drivers let you point at a directory of data files and query them as database tables. All three share a common limitation: **runtime parsing**. At `Open()` time, files are read from disk, parsed (CSV/JSON/XML), type-inferred, and loaded into SQLite. This adds startup latency, introduces type inference ambiguity, and requires the data files to ship with the binary.

For **internal data you control** — reference tables, lookup data, configuration constants, test fixtures, enum metadata — the file-based drivers are the wrong tool. The data rarely changes, is always deployed with the code, and would benefit from compile-time validation. Yet there is no way to query compiled-in Go data through the ORM without manually wiring up `arraysource.NewArraySourceFrom()` for each dataset and passing each to `Model()` individually.

GODB solves this by treating **Go data arrays as database tables**. Think of it as `NewArraySource` — but multiple arrays, each array is a table, all the tables form a database. The data is already in the Go binary (compiled at build time), already typed, already in memory. No file I/O, no parsing, no type inference. The driver creates an in-memory SQLite database and populates it from the Go variables at `Open()` time.

### Desired Experience

```go
// pkg/blogs/blogs.go — normal Go source, compiled into the binary
package blogs

type Blog struct {
    ID         int64  `db:"id"`
    Title      string `db:"title"`
    CategoryID int64  `db:"category_id"`
}

type Category struct {
    ID   int64  `db:"id"`
    Name string `db:"name"`
}

var Blogs = []Blog{
    {ID: 1, Title: "Hello World", CategoryID: 1},
    {ID: 2, Title: "Go Tips", CategoryID: 2},
    {ID: 3, Title: "Advanced Patterns", CategoryID: 2},
}

var Categories = []Category{
    {ID: 1, Name: "General"},
    {ID: 2, Name: "Programming"},
}
```

```go
// main.go
package main

import (
    "myapp/pkg/blogs"

    "github.com/dracory/neat"
    "github.com/dracory/neat/database/driver/godb"
)

func main() {
    config := neat.DBConfig{
        Default: "go_db",
        Connections: map[string]neat.ConnectionConfig{
            "go_db": {
                Driver: "godb",
                Tables: godb.Tables{
                    "blogs":      blogs.Blogs,
                    "categories": blogs.Categories,
                },
            },
        },
    }

    database, _ := neat.New(config)
    defer database.Close()

    // Query blogs — works like any other database
    var blogs []Blog
    database.Query().Model(&Blog{}).Where("category_id = ?", 2).Get(&blogs)

    // JOIN across tables — both are in one SQLite DB
    type BlogWithCategory struct {
        ID           int64
        Title        string
        CategoryName string
    }
    var results []BlogWithCategory
    database.Query().
        Table("blogs").
        LeftJoin("categories", "blogs.category_id = categories.id").
        Select("blogs.id", "blogs.title", "categories.name AS category_name").
        Get(&results)

    // Aggregates work too
    var count int
    database.Query().Table("blogs").Count(&count)
}
```

The data is Go source code, compiled into the binary at build time. It is human-readable during development, version-controlled, and type-checked by the compiler. At runtime, the data exists as native Go variables in memory. The driver populates an in-memory SQLite database from these variables — the only runtime work is SQLite table creation and row insertion. No file I/O, no parsing, no type inference.

### Multiple GODB Databases

Each GODB connection has its own `Tables` — completely isolated. No global registry, no `init()`, no side-effect imports. You can have multiple GODB databases in the same application:

```go
config := neat.DBConfig{
    Default: "blog_db",
    Connections: map[string]neat.ConnectionConfig{
        "blog_db": {
            Driver: "godb",
            Tables: godb.Tables{
                "blogs":      blogs.Blogs,
                "categories": blogs.Categories,
            },
        },
        "shop_db": {
            Driver: "godb",
            Tables: godb.Tables{
                "products": products.Products,
                "orders":   orders.Orders,
            },
        },
    },
}
```

Each connection gets its own in-memory SQLite database with only its own tables.

### Alternative Config Style

Two config styles are proposed for passing data to the driver. The final choice is an open question (see below).

**Style A — `godb.Tables` map (recommended):**

```go
"go_db": {
    Driver: "godb",
    Tables: godb.Tables{
        "blogs":      blogs.Blogs,
        "categories": blogs.Categories,
    },
},
```

`godb.Tables` is a `map[string]any` where the key is the table name and the value is the data slice. The value can be `[]map[string]any`, `[]SomeStruct`, or any slice — the driver converts it to rows using the same logic as `arraysource.NewArraySourceFrom`. Simplest and cleanest.

**Style B — `[]godb.Table` slice:**

```go
"go_db": {
    Driver: "godb",
    Tables: []godb.Table{
        {Name: "blogs",      Data: blogs.Blogs},
        {Name: "categories", Data: blogs.Categories},
    },
},
```

`godb.Table` is a struct with `Name string` and `Data any` fields. More verbose but preserves table declaration order (useful if table creation order matters for foreign key constraints in future extensions).

## Proposed Solution

Implement a new `GODB` driver that accepts Go data arrays directly via config and populates an in-memory SQLite database at `Open()` time:

1. **Embeds `*SQLite`** for all `Driver` interface methods (Open, Close, Ping, BeginTx, Placeholder, Dialect)
2. **Overrides `Open`** to iterate the `Tables` from config, converting each data slice to rows and creating one SQLite table per entry
3. **Returns `Dialect() == "sqlite"`** so the query builder generates SQLite-compatible SQL — no query builder changes needed

The driver is stateless. All state lives in the in-memory SQLite database. When the connection is closed, everything is gone. The data itself is immutable — it lives in compiled Go variables and cannot be modified at runtime.

### Mental Model

```
Build time (go build):
  pkg/blogs/blogs.go ──── compiled ────▶  Go binary (Blogs, Categories vars in memory)

Runtime (Open()):
  Config                           SQLite (in-memory)
  ┌──────────────────┐             ┌──────────────────┐
  │ Tables:          │             │                  │
  │   "blogs" ───────┼────────────▶│ "blogs" table    │
  │   → blogs.Blogs  │             │                  │
  │   "categories" ──┼────────────▶│ "categories" tbl │
  │   → blogs.Cats   │             │                  │
  └──────────────────┘             └──────────────────┘
```

- User passes Go data slices directly in config via `Tables` field
- Each entry becomes one SQLite table (struct slices converted to `[]map[string]any` via reflection)
- Types are exact Go types — `int64`, `float64`, `bool`, `time.Time`, `string`
- No file I/O, no parsing, no type inference — data is already in memory as Go variables
- The only runtime work is SQLite table creation and row insertion

### Architecture

```
database/
  driver/
    godb.go               // GODB driver (embeds *SQLite, overrides Open)
    godb_tables.go        // Tables type, Table struct, data conversion
    godb_test.go          // Driver unit tests

integration_tests/
  godb/
    helper.go             // SetupGODBTest, model types, sample data
    godb_query_test.go    // Query, WHERE, ORDER BY, LIMIT, First, Find, Count
    godb_join_test.go     // LEFT JOIN, INNER JOIN, 3-table JOIN, JOIN+WHERE
    godb_aggregate_test.go   // COUNT, SUM, AVG, MIN/MAX, GROUP BY
    godb_type_test.go     // Native type preservation, nested structs, NULL

examples/
  godb-driver/
    main.go               // Example usage
    main_test.go
    README.md
    data/
      data.go             // Sample data (struct slices)
```

No new contracts. No new interfaces. No changes to `contracts/database/orm/`. No changes to `database/query/`. The driver plugs into the existing driver registration system — same integration points as `csvdb`, `jsondb`, and `xmldb`. The only new type is `godb.Tables` (or `[]godb.Table`) passed via the config `Tables` field.

## Data Types

### Tables Type

```go
// database/driver/godb_tables.go

package driver

// Tables is a map of table names to data slices.
// Each value can be []map[string]any, []SomeStruct, or any slice.
// The driver converts struct slices to []map[string]any using the same
// reflection logic as arraysource.NewArraySourceFrom.
//
// Usage in config:
//
//   Tables: godb.Tables{
//       "blogs":      blogs.Blogs,
//       "categories": blogs.Categories,
//   }
type Tables map[string]any

// Table is an alternative config style that preserves declaration order.
// Useful if table creation order matters (e.g., for foreign key constraints).
//
// Usage in config:
//
//   Tables: []godb.Table{
//       {Name: "blogs",      Data: blogs.Blogs},
//       {Name: "categories", Data: blogs.Categories},
//   }
type Table struct {
    Name string
    Data any
}
```

### Supported Data Types

The `Tables` values can be:

| Go type | Behavior |
|---------|----------|
| `[]map[string]any` | Used directly — each map is a row, keys are columns |
| `[]SomeStruct` | Converted to `[]map[string]any` via reflection (same as `arraysource.NewArraySourceFrom`) |
| `[]*SomeStruct` | Same as `[]SomeStruct` — nil elements skipped |
| `[]map[string]any` with nested values | Nested maps/slices stored as JSON strings (queryable via SQLite JSON functions) |

### Struct to Row Conversion

Struct slices are converted to `[]map[string]any` using the same logic as `arraysource.NewArraySourceFrom`:
- Field names resolved to column names using `db` > `neat` > `gorm` > snake_case tags
- Embedded structs flattened
- Association fields (slices, struct pointers) skipped
- Nullable pointer fields (`*string`, `*int`) dereferenced; nil pointers → NULL
- `time.Time` fields included as-is

This means the user's existing struct definitions work unchanged — no special tags or modifications needed.

## Driver Implementation

```go
// database/driver/godb.go

package driver

import (
    "database/sql"
    "fmt"
    "sort"
    "strings"

    _ "modernc.org/sqlite"
)

// GODB implements the Driver interface for Go-data-backed storage.
// It embeds *SQLite for all standard Driver methods and overrides Open
// to populate an in-memory SQLite database from Go data slices passed
// via the config Tables field.
//
// The data is compiled into the binary at build time by the Go compiler.
// At runtime, the driver converts each data slice to rows and inserts
// them into SQLite. No file I/O, no parsing.
type GODB struct {
    *SQLite
    tables Tables
}

func NewGODB() *GODB {
    return &GODB{SQLite: NewSQLite()}
}

// SetTables configures the data tables before Open is called.
// Called by the ORM during connection setup, reading from config.
func (g *GODB) SetTables(tables Tables) {
    g.tables = tables
}

func (g *GODB) Dialect() string {
    return "sqlite"
}

func (g *GODB) Open(_ string) (*sql.DB, error) {
    // The DSN (dirPath) is ignored — data comes from the Tables config.
    db, err := sql.Open("sqlite", ":memory:")
    if err != nil {
        return nil, fmt.Errorf("godb: failed to open in-memory SQLite: %w", err)
    }

    // Sort table names for deterministic creation order
    tableNames := make([]string, 0, len(g.tables))
    for name := range g.tables {
        tableNames = append(tableNames, name)
    }
    sort.Strings(tableNames)

    seen := make(map[string]bool)

    for _, tableName := range tableNames {
        if seen[strings.ToLower(tableName)] {
            _ = db.Close()
            return nil, fmt.Errorf("godb: table name collision (case-insensitive): %s", tableName)
        }
        seen[strings.ToLower(tableName)] = true

        data := g.tables[tableName]
        if err := populateGODBTable(db, tableName, data); err != nil {
            _ = db.Close()
            return nil, fmt.Errorf("godb: failed to populate table %s: %w", tableName, err)
        }
    }

    return db, nil
}
```

### Table Population

```go
// database/driver/godb_tables.go (continued)

// populateGODBTable converts a data slice to rows, infers schema,
// creates the table, and inserts all rows in a transaction.
func populateGODBTable(db *sql.DB, tableName string, data any) error {
    // 1. Validate table name (SQL injection prevention)
    if !isSimpleIdentifier(tableName) {
        return fmt.Errorf("invalid table name: %s", tableName)
    }

    // 2. Convert data to []map[string]any
    //    - []map[string]any → used directly
    //    - []SomeStruct    → converted via reflection (same as NewArraySourceFrom)
    //    - []*SomeStruct   → converted, nil elements skipped
    rows := convertToRows(data)
    if len(rows) == 0 {
        return nil // no data = no table
    }

    // 3. Infer schema from Go native types
    //    Go types map directly: int64→INTEGER, float64→REAL, bool→INTEGER,
    //    time.Time→DATETIME, string→TEXT, []byte→BLOB
    schema := inferSchemaFromRows(rows)
    if len(schema) == 0 {
        return nil // 0 columns = skip table
    }

    // 4. Validate column names (simple identifiers, no case-insensitive dupes)
    // 5. CREATE TABLE with inferred schema
    // 6. BEGIN TRANSACTION
    // 7. INSERT rows in batches (SQLite parameter limit ~999)
    // 8. COMMIT
    // (Same logic as Array driver — can be shared via a helper)
    return nil
}

// convertToRows converts a data slice to []map[string]any.
// If data is already []map[string]any, it's used directly.
// If data is a struct slice, it's converted via reflection using the
// same logic as arraysource.NewArraySourceFrom.
func convertToRows(data any) []map[string]any {
    // Check if it's already []map[string]any
    if rows, ok := data.([]map[string]any); ok {
        return rows
    }
    // Otherwise, use reflection to convert struct slices
    // (delegates to the same structToMap logic used by NewArraySourceFrom)
    return structsToRows(data)
}
```

### Key Difference from File-Based Drivers: No Parsing

| Concern | CSVDB / JSONDB / XMLDB | GODB |
|---------|------------------------|------|
| Data source | Files on disk | Go variables (compiled into binary) |
| How data reaches driver | Directory scan at Open() time | Passed directly via config `Tables` field |
| Parsing at startup | Yes (CSV/JSON/XML → Go values) | No (already Go values in memory) |
| Type safety | Runtime (inference may be wrong) | Compile-time (compiler enforces struct fields) |
| Human-readable | Yes (text files) | Yes (Go source code) |
| Mutable at runtime | Yes (change the files) | No (compiled into binary) |
| Data changes | Edit file, restart | Edit `.go`, recompile, redeploy |
| External tools | Any text editor, any language | Go IDE only |
| Binary size | Zero (files on disk) | Data in binary (Go vars) |
| File I/O at startup | Yes (read + parse) | No |
| Type inference | Required (string→type or float64→int64) | Minimal (Go types → SQLite types, direct mapping) |
| Best for | External data you don't control | Internal data you control |

### Relationship to Array Driver

GODB is essentially "Array driver with multiple tables and config-based wiring":

| Aspect | Array Driver | GODB |
|--------|-------------|------|
| Tables | One per `Model()` call | Multiple, all at `Open()` time |
| Data source | `ArraySource` interface (Rows() method) | Go data slices via config `Tables` |
| Wiring | Manual: `database.Query().Model(arraysource.NewArraySourceFrom(data))` | Config: `Tables: godb.Tables{"name": data}` |
| JOINs | Not possible (each source is a separate connection) | Yes — all tables in one SQLite DB |
| Schema inference | Same `goTypeToSQLite`, `inferSchema` | Same logic, reused |
| Batch insert | Same `insertRows` logic | Same logic, reused |

### Type Mapping

Go types map directly to SQLite types — no string-to-type inference needed:

| Go type | SQLite type |
|---------|-------------|
| `int`, `int8`, `int16`, `int32`, `int64` | INTEGER |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64` | INTEGER |
| `float32`, `float64` | REAL |
| `bool` | INTEGER (0/1) |
| `string` | TEXT |
| `time.Time` | DATETIME |
| `[]byte`, `json.RawMessage` | BLOB |
| `nil` | NULL (doesn't affect column type) |
| Nested `map[string]any` | TEXT (JSON string, queryable via SQLite JSON functions) |
| Nested `[]any` | TEXT (JSON string) |

This is identical to the Array driver's `goTypeToSQLite` function — GODB reuses the same type mapping.

### Schema Inference

Schema inference is simpler than file-based drivers because Go types are exact. The driver examines all rows, collects the union of all keys, and maps each value's Go type to its SQLite type. Type widening (INTEGER → REAL → TEXT) applies when a column has mixed types across rows — same `widenType` function used by all drivers.

Columns are sorted alphabetically for deterministic ordering, same as JSONDB and CSVDB.

## Integration Points

The driver requires additions to the same files touched by `jsondb`, plus a new `Tables` field in `ConnectionConfig`. No new contracts beyond the `Tables` type, no query builder changes, no new interfaces.

### 1. Driver Constant

```go
// contracts/database/config.go
DriverGODB      Driver = "godb"
```

### 2. Tables Field in ConnectionConfig

```go
// database/db/config_builder.go — ConnectionConfig
type ConnectionConfig struct {
    Driver   string
    Dsn      string
    Host     string
    Port     int
    Database string
    Username string
    Password string
    Schema   string
    Tables   any  // GODB: godb.Tables or []godb.Table; ignored by other drivers
}
```

The `Tables` field is `any` to avoid importing the driver package from the config package. The GODB driver type-asserts it to `Tables` (or `[]Table`) at `Open()` time. Other drivers ignore it.

### 3. Driver Registration

```go
// database/orm/orm.go — createDriver()
case "godb":
    return driver.NewGODB()
```

### 4. Tables Wiring

```go
// database/orm/orm.go — connection setup
// After creating the driver, if it's a GODB driver, pass Tables from config
if gd, ok := d.(*driver.GODB); ok {
    gd.SetTables(config.Tables.(driver.Tables))
}
```

### 5. DSN Builder

```go
// database/db/config_builder.go — BuildDSN()
case "godb":
    return ":memory:", nil
```

GODB always uses in-memory SQLite. The `Database` field is not used — data comes from `Tables`.

### 6. Config Validation

```go
// database/db/config_builder.go — ConnectionConfig.Validate()
case "sqlite", "array", "csvdb", "jsondb", "xmldb", "godb":
    // database path is optional; empty defaults to :memory:
    // For GODB, the Database field is not used — data comes from Tables.
    return nil
```

### 7. Connection Pool Configuration

```go
// database/orm/orm.go — configureConnectionPool()
pinSingleConn := connConfig.Driver == "sqlite" || connConfig.Driver == "array" || connConfig.Driver == "csvdb" || connConfig.Driver == "jsondb" || connConfig.Driver == "xmldb" || connConfig.Driver == "godb"
```

GODB uses in-memory SQLite, so it shares the same single-connection constraint.

### 8. Database Name Detection

```go
// database/db.go — detectDatabaseName()
case "sqlite", "turso", "array", "csvdb", "jsondb", "xmldb", "godb":
    return "main"
```

### 9. Schema Builder

```go
// database/schema/schema.go — New()
case contractsdatabase.DriverGODB:
    // GODB uses SQLite grammar since Dialect() returns "sqlite".
    sqliteGrammar := grammars.NewSqlite(log, prefix)
    driverSchema = NewSqliteSchema(sqliteGrammar, orm, prefix)
    grammar = sqliteGrammar
    processor = processors.NewSqlite()
```

## Robustness Features

- **Table name collision detection**: If two entries in `Tables` produce colliding table names (case-insensitive), the driver returns an error at `Open()` time.
- **Invalid identifier detection**: Table names and column names are validated with `isSimpleIdentifier` to prevent SQL injection.
- **Transaction wrapping**: Batch INSERTs are wrapped in a `BEGIN`/`COMMIT` transaction for atomicity and performance.
- **Empty dataset handling**: A data slice with 0 rows is skipped — no table created. Consistent with JSONDB's empty-array behavior.
- **0-column skip**: If rows exist but no columns could be inferred (e.g., all maps are empty `{}`), the table is skipped. Consistent with JSONDB's `[{}]` handling.
- **MaxGODBRows limit**: 100,000-row limit per table to prevent unbounded memory consumption. Same limit as `MaxArrayRows` and `MaxJSONRows`.
- **Immutable data**: Data is compiled into the binary — no concern about file changes, no need for file watching or cache invalidation.
- **Type safety**: Struct field types are validated by the Go compiler at build time. A field typed as `int64` cannot accidentally contain a string.

## Testing

### Unit Tests (`database/driver/`)
- Driver open, dialect, empty Tables
- Table creation from `[]map[string]any` data
- Table creation from `[]Struct` data (struct-to-row conversion)
- Table creation from `[]*Struct` data (nil elements skipped)
- Type mapping: int64, float64, bool, string, time.Time, nil
- Type widening (INTEGER + REAL → REAL, etc.)
- Table name collision (case-insensitive)
- Invalid identifiers
- Empty dataset → skip table
- 0-column rows → skip table
- MaxGODBRows limit enforcement
- Transaction wrapping, batch insertion
- Both config styles: `Tables` map and `[]Table` slice

### Integration Tests (`integration_tests/godb/`)
- **Query**: basic query, WHERE equals, WHERE bool, ORDER BY asc/desc, LIMIT, LIMIT+OFFSET, First, Find, Count
- **JOIN**: LEFT JOIN, INNER JOIN, 3-table JOIN, JOIN with WHERE
- **Aggregates**: COUNT, COUNT with WHERE, SUM, AVG, MIN/MAX, GROUP BY, SUM with JOIN+GROUP BY
- **Type preservation**: integer, float, bool, string, datetime, NULL handling, nested map as JSON string
- **Struct-based data**: verify struct slices work end-to-end with `db`/`neat`/`gorm` tags
- **Both config styles**: verify `Tables` map and `[]Table` slice both work

### Example (`examples/godb-driver/`)
- Working example with sample data (blogs, categories) as struct slices
- Demonstrates WHERE, JOIN (across struct-based tables), and aggregate queries
- Shows both `Tables` map and `[]Table` config styles
- Shows nested struct querying via SQLite JSON functions

## Open Questions

### 1. Config Style: `Tables` Map vs `[]Table` Slice

Two config styles are proposed:

**Style A — `godb.Tables` map:**
```go
Tables: godb.Tables{
    "blogs":      blogs.Blogs,
    "categories": blogs.Categories,
},
```
- Simplest, cleanest
- No table creation order guarantee (Go maps are unordered — driver sorts alphabetically)
- Recommended for v1

**Style B — `[]godb.Table` slice:**
```go
Tables: []godb.Table{
    {Name: "blogs",      Data: blogs.Blogs},
    {Name: "categories", Data: blogs.Categories},
},
```
- More verbose but preserves declaration order
- Useful if table creation order matters (e.g., foreign key constraints in future extensions)
- The `Tables` config field is `any`, so both styles are accepted

**Recommendation**: Support both. The driver type-asserts the `Tables` field: if it's `Tables` (map), use it directly; if it's `[]Table`, iterate in order. The `Tables` map style is recommended for documentation examples; the `[]Table` style is available for advanced cases.

### 2. `Tables` Field on `ConnectionConfig`

Adding a `Tables any` field to `ConnectionConfig` is a new addition. It's `any` to avoid circular imports (config package can't import driver package). The GODB driver type-asserts it at `Open()` time. Other drivers ignore it.

**Alternative**: Pass `Tables` via a separate mechanism (e.g., a `SetTables()` method called after driver creation but before `Open()`). This avoids modifying `ConnectionConfig` but requires special handling in the ORM setup code.

**Recommendation**: Add the `Tables` field to `ConnectionConfig`. It's a clean, config-driven approach consistent with how other drivers receive their configuration (e.g., `Database` for CSVDB directory path, `Dsn` for MySQL).

### 3. Sharing Schema Inference Logic with Array Driver

The GODB driver's `populateGODBTable` function performs the same steps as `Array.Populate` — schema inference, table creation, batch insertion. The logic could be:
- **A**: Duplicate the code in `godb.go` (self-contained, like CSVDB/JSONDB copied their logic)
- **B**: Extract shared helpers into a `populate.go` file in the driver package
- **C**: Have GODB use the Array driver's `Populate` method directly

**Recommendation**: Option B (shared helpers). The schema inference, table creation, and batch insert logic is identical across Array, JSONDB, and GODB. Extracting it into a shared `populate.go` would reduce duplication. However, this is a refactoring concern — for v1, duplicating the logic (Option A) is acceptable and consistent with the existing CSVDB/JSONDB approach.

### 4. Struct Slice Conversion: Reuse `arraysource` or Inline?

The `convertToRows` function needs to convert struct slices to `[]map[string]any`. This is the same logic as `arraysource.NewArraySourceFrom`. Options:
- **A**: Import `arraysource` and call `NewArraySourceFrom` internally
- **B**: Copy the `structsToRows` / `structToMap` logic into `godb_tables.go`
- **C**: Extract the struct-to-row logic into a shared package

**Recommendation**: Option A (import `arraysource`). The logic already exists and is tested. Duplicating reflection code is error-prone. The `arraysource` package is lightweight with no heavy dependencies.

### 5. Should `Tables` Support Dynamic Addition After Open?

Should users be able to add tables after `Open()` (e.g., `database.AddTable("new_table", newData)`)? This would require the driver to hold a reference to the `*sql.DB` and populate new tables on demand.

**Recommendation**: Not in v1. All tables are defined at config time and populated at `Open()` time. Dynamic addition can be added later if needed — the Array driver's `Populate` method already supports this pattern.

## Relationship to Existing Work

| Component | Relationship |
|-----------|-------------|
| `csvdb` driver | Architectural template — same embedding pattern, same integration points |
| `jsondb` driver | Architectural template — same 0-column skip, same empty-data skip behavior |
| `array` driver | Direct predecessor — GODB is "Array driver with multiple tables, config-based wiring, and JOIN support". Same `goTypeToSQLite`, `inferSchema`, `insertRows` logic |
| `support/arraysource` | Used for struct-to-`[]map[string]any` conversion (imported, not duplicated) |
| `flat-file-driver.md` proposal | GODB is not a file-based driver — it's a data-as-code driver. It complements the file-based drivers (CSVDB, JSONDB, XMLDB) by serving the "internal data" use case |

## Ideal Use Cases

- **Reference/lookup data** — country codes, currency codes, status enums, category taxonomies, ISO standards
- **Test fixtures** — compile test data directly into test binaries, no external files to manage, no `t.TempDir()` boilerplate
- **Configuration data** — feature flags, default settings, seed data for migrations, environment-specific constants
- **Small-to-medium immutable datasets** — anything that changes rarely and is always deployed with the code
- **Embedded applications** — Go binaries running on devices or in containers where external data files are undesirable
- **Application data with relationships** — blogs with categories, orders with products, users with roles — JOINs work because all tables are in one SQLite DB

## Comparison with File-Based Drivers

| Aspect | File-Based (CSVDB/JSONDB/XMLDB) | GODB |
|--------|----------------------------------|------|
| Data changes | Edit file, restart app | Edit `.go`, recompile, redeploy |
| Data source | External files on disk | Go variables compiled into binary |
| How data reaches driver | Directory scan at Open() time | Passed directly via config `Tables` field |
| Type safety | Runtime inference | Compile-time enforcement (struct fields) |
| Startup cost | File I/O + parsing | SQLite table creation only (data already in memory) |
| Human-readable | Yes (text files) | Yes (Go source) |
| Cross-language | Yes (JSON, CSV, XML are universal) | No (Go only) |
| External data | Yes (exports, API responses, feeds) | No |
| Binary size | Zero (files separate) | Data in binary (Go vars) |
| Best for | Data you don't control | Data you control |

## Future Extensions

- **Dynamic table addition**: `database.AddTable("name", data)` after `Open()` — for runtime-constructed datasets
- **Mixed-mode driver**: A `filedb` driver that unifies CSVDB, JSONDB, XMLDB, and GODB — file-based tables and compiled-in tables in one SQLite database
- **Data validation**: Pre-flight validation of table names, column names, and type consistency before `Open()` — fail fast on config errors
- **Persistent caching**: Save the populated SQLite database to a file on disk and reuse it on subsequent runs if the data hasn't changed
- **Foreign key constraints**: Support declaring relationships between tables in config (e.g., `godb.Table{Name: "blogs", Data: Blogs, FK: []godb.FK{{Column: "category_id", References: "categories(id)"}}}`)
