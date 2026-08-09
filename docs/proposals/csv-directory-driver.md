# CSV Directory Driver

**Date**: August 9, 2026
**Status**: Proposal
**Priority**: Medium

## Problem Statement

The array driver lets you query in-memory data via SQLite, but you must supply the rows yourself. There is no built-in way to point at a directory of CSV files and query them as if they were database tables.

Users who have CSV data (exports, reports, test fixtures, datasets) must write boilerplate parsing code before they can use the ORM's query builder.

### Desired Experience

```go
config := neat.DBConfig{
    Default: "csv_db",
    Connections: map[string]neat.ConnectionConfig{
        "csv_db": {
            Driver:   "csvdir",
            Database: "data/",   // directory path
        },
    },
}

database, _ := neat.New(config)
defer database.Close()

// data/users.csv → "users" table
var users []User
err := database.Query().Model(&User{}).Where("active = ?", true).Get(&users)

// data/products.csv → "products" table
var products []Product
err = database.Query().Model(&Product{}).Where("price > ?", 50).Get(&products)
```

The directory is the database. Each `.csv` file is a table. The filename (without `.csv`) is the table name. The CSV header row defines the columns. No per-file configuration, no `ArraySource` structs, no manual parsing.

## Proposed Solution

Implement a new `csvdir` driver that:

1. **Embeds `*SQLite`** for all `Driver` interface methods (Open, Close, Ping, BeginTx, Placeholder, Dialect)
2. **Overrides `Open`** to scan the directory for `.csv` files, parse each one, and populate an in-memory SQLite database with one table per file
3. **Returns `Dialect() == "sqlite"`** so the query builder generates SQLite-compatible SQL and uses SQLite placeholders — no query builder changes needed

The driver is stateless. All state lives in the in-memory SQLite database. When the connection is closed, everything is gone.

### Mental Model

```
Directory (database)          SQLite (in-memory)
┌──────────────────┐         ┌──────────────────┐
│ data/            │         │                  │
│   users.csv  ────┼────────▶│ "users" table    │
│   products.csv ──┼────────▶│ "products" table │
│   orders.csv  ───┼────────▶│ "orders" table   │
└──────────────────┘         └──────────────────┘
```

- `data/users.csv` → table `"users"` (filename without `.csv`)
- First row of each CSV is the header → column names
- Remaining rows → data, with type inference (int, float, bool, time, string)
- All tables populated at `Open` time (eager, not lazy)

### Architecture

```
database/
  driver/
    csvdir.go              // CSVDir driver (embeds *SQLite, overrides Open)
    csvdir_csv.go          // CSV parsing + type inference helper
    csvdir_test.go         // Driver unit tests
    csvdir_csv_test.go     // CSV parser tests
  csvdir_integration_test.go  // Integration test via Database.Query()

examples/
  csvdir-driver/
    main.go                // Example usage
    main_test.go
    README.md
    data/
      users.csv            // Sample CSV
      products.csv         // Sample CSV
      orders.csv           // Sample CSV
```

No new contracts. No new interfaces. No changes to `contracts/database/orm/`. No changes to `database/query/`. The driver plugs into the existing driver registration system.

## Driver Implementation

```go
// database/driver/csvdir.go

package driver

import (
    "context"
    "database/sql"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    _ "modernc.org/sqlite"
)

// CSVDir implements the Driver interface for CSV-directory-backed storage.
// It embeds *SQLite for all standard Driver methods and overrides Open to
// scan a directory of CSV files and populate an in-memory SQLite database.
//
// The directory path is passed as the DSN (from ConnectionConfig.Database).
// Each .csv file in the directory becomes a table named after the filename
// (without the .csv extension). The first row of each CSV is the header.
//
// The driver is stateless — all state lives in the in-memory SQLite database.
type CSVDir struct {
    *SQLite
}

// NewCSVDir creates a new CSVDir driver.
func NewCSVDir() *CSVDir {
    return &CSVDir{SQLite: NewSQLite()}
}

// Dialect returns "sqlite" so the query builder generates SQLite-compatible
// SQL and uses SQLite placeholders. The query builder's isSQLite() check
// returns true, and no ArraySource Model() hook fires (tables are already
// populated at Open time).
func (c *CSVDir) Dialect() string {
    return "sqlite"
}

// Open opens an in-memory SQLite database, scans the directory at dirPath
// for .csv files, and populates one table per file. The dirPath is the DSN,
// which comes from ConnectionConfig.Database via BuildDSN.
//
// If dirPath is empty or ":memory:", no directory is scanned and an empty
// in-memory SQLite database is returned (useful for testing).
func (c *CSVDir) Open(dirPath string) (*sql.DB, error) {
    db, err := sql.Open("sqlite", ":memory:")
    if err != nil {
        return nil, fmt.Errorf("csvdir: failed to open in-memory SQLite: %w", err)
    }

    // If no directory path, return empty in-memory DB
    if dirPath == "" || dirPath == ":memory:" {
        return db, nil
    }

    // Verify directory exists
    info, err := os.Stat(dirPath)
    if err != nil {
        db.Close()
        return nil, fmt.Errorf("csvdir: cannot access directory %s: %w", dirPath, err)
    }
    if !info.IsDir() {
        db.Close()
        return nil, fmt.Errorf("csvdir: %s is not a directory", dirPath)
    }

    // Scan for .csv files and populate tables
    entries, err := os.ReadDir(dirPath)
    if err != nil {
        db.Close()
        return nil, fmt.Errorf("csvdir: cannot read directory %s: %w", dirPath, err)
    }

    for _, entry := range entries {
        if entry.IsDir() {
            continue
        }
        if !strings.HasSuffix(strings.ToLower(entry.Name()), ".csv") {
            continue
        }

        tableName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
        filePath := filepath.Join(dirPath, entry.Name())

        if err := populateCSVFile(db, tableName, filePath); err != nil {
            db.Close()
            return nil, fmt.Errorf("csvdir: failed to populate table %s from %s: %w", tableName, filePath, err)
        }
    }

    return db, nil
}

// populateCSVFile reads a CSV file, infers column types, creates a SQLite
// table, and inserts all rows.
func populateCSVFile(db *sql.DB, tableName string, filePath string) error {
    // 1. Open and parse CSV file
    // 2. First row = header (column names)
    // 3. Remaining rows = data
    // 4. Infer column types from data (int → float → bool → time → string)
    // 5. CREATE TABLE with inferred schema
    // 6. INSERT rows in batches
    // ...
}
```

### CSV Parsing and Type Inference

```go
// database/driver/csvdir_csv.go

package driver

import (
    "database/sql"
    "encoding/csv"
    "fmt"
    "os"
    "sort"
    "strconv"
    "strings"
    "time"
)

// parseCSV reads a CSV file and returns column names, rows, and inferred
// column types.
func parseCSV(filePath string) (columns []string, rows [][]string, err error) {
    f, err := os.Open(filePath)
    if err != nil {
        return nil, nil, err
    }
    defer f.Close()

    reader := csv.NewReader(f)
    reader.LazyQuotes = true

    allRecords, err := reader.ReadAll()
    if err != nil {
        return nil, nil, fmt.Errorf("CSV parse error: %w", err)
    }
    if len(allRecords) == 0 {
        return nil, nil, fmt.Errorf("CSV file is empty")
    }

    // First row is the header
    columns = allRecords[0]
    rows = allRecords[1:]
    return columns, rows, nil
}

// inferColumnTypes examines all rows and infers the SQLite type for each
// column. For each column, it tries int → float → bool → time → string,
// widening to a more general type if any value doesn't fit.
func inferColumnTypes(columns []string, rows [][]string) []string {
    types := make([]string, len(columns))
    for i := range types {
        types[i] = "INTEGER" // start optimistic, widen as needed
    }

    for _, row := range rows {
        for i, val := range row {
            if i >= len(types) {
                break
            }
            if val == "" {
                continue // skip empty strings, don't affect type
            }

            valType := inferValueType(val)
            types[i] = widenType(types[i], valType)
        }
    }

    return types
}

// inferValueType tries to determine the Go-native type of a string value.
func inferValueType(val string) string {
    if _, err := strconv.ParseInt(val, 10, 64); err == nil {
        return "INTEGER"
    }
    if _, err := strconv.ParseFloat(val, 64); err == nil {
        return "REAL"
    }
    if val == "true" || val == "false" || val == "True" || val == "False" {
        return "INTEGER" // SQLite stores bools as 0/1
    }
    if _, err := time.Parse(time.RFC3339, val); err == nil {
        return "DATETIME"
    }
    // Try common date formats
    for _, layout := range []string{"2006-01-02", "2006-01-02 15:04:05", "01/02/2006"} {
        if _, err := time.Parse(layout, val); err == nil {
            return "DATETIME"
        }
    }
    return "TEXT"
}

// widenType returns the wider of two SQLite types.
// INTEGER → REAL → TEXT is the widening chain.
// DATETIME is compatible only with itself and TEXT.
func widenType(current, new string) string {
    if current == new {
        return current
    }
    if current == "INTEGER" && new == "REAL" {
        return "REAL"
    }
    if current == "REAL" && new == "INTEGER" {
        return "REAL"
    }
    // Any incompatible mix → TEXT
    return "TEXT"
}

// convertValue converts a string value to the appropriate Go type for
// the given SQLite column type, for insertion via database/sql.
func convertValue(val string, sqlType string) any {
    if val == "" {
        return nil
    }

    switch sqlType {
    case "INTEGER":
        if val == "true" || val == "True" {
            return 1
        }
        if val == "false" || val == "False" {
            return 0
        }
        if n, err := strconv.ParseInt(val, 10, 64); err == nil {
            return n
        }
        // Fallback if inference was wrong for this specific row
        return val
    case "REAL":
        if f, err := strconv.ParseFloat(val, 64); err == nil {
            return f
        }
        return val
    case "DATETIME":
        if t, err := time.Parse(time.RFC3339, val); err == nil {
            return t
        }
        for _, layout := range []string{"2006-01-02", "2006-01-02 15:04:05", "01/02/2006"} {
            if t, err := time.Parse(layout, val); err == nil {
                return t
            }
        }
        return val
    default:
        return val
    }
}

// populateCSVFile reads a CSV file, infers types, creates a table, and
// inserts all rows. This is the core logic called by CSVDir.Open for each
// .csv file in the directory.
func populateCSVFile(db *sql.DB, tableName string, filePath string) error {
    columns, rows, err := parseCSV(filePath)
    if err != nil {
        return err
    }

    // Validate table and column names (SQL injection prevention)
    if !isSimpleIdentifier(tableName) {
        return fmt.Errorf("invalid table name derived from filename: %s", tableName)
    }
    for _, col := range columns {
        if !isSimpleIdentifier(col) {
            return fmt.Errorf("invalid column name in CSV header: %s", col)
        }
    }

    // Infer column types
    colTypes := inferColumnTypes(columns, rows)

    // Sort columns for deterministic ordering
    // (not strictly necessary, but consistent with the array driver)
    indices := make([]int, len(columns))
    for i := range indices {
        indices[i] = i
    }
    sort.Slice(indices, func(a, b int) bool {
        return columns[indices[a]] < columns[indices[b]]
    })

    // CREATE TABLE
    var colDefs []string
    for _, idx := range indices {
        colDefs = append(colDefs, fmt.Sprintf("\"%s\" %s", columns[idx], colTypes[idx]))
    }
    createSQL := fmt.Sprintf("CREATE TABLE \"%s\" (%s)", tableName, strings.Join(colDefs, ", "))
    if _, err := db.Exec(createSQL); err != nil {
        return fmt.Errorf("failed to create table: %w", err)
    }

    if len(rows) == 0 {
        return nil
    }

    // INSERT rows in batches (SQLite parameter limit is ~999)
    colCount := len(columns)
    batchSize := 500 / colCount
    if batchSize == 0 {
        batchSize = 1
    }

    // Build column list (sorted)
    var colNames []string
    var placeholders []string
    for _, idx := range indices {
        colNames = append(colNames, fmt.Sprintf("\"%s\"", columns[idx]))
        placeholders = append(placeholders, "?")
    }
    insertPrefix := fmt.Sprintf("INSERT INTO \"%s\" (%s) VALUES ",
        tableName, strings.Join(colNames, ", "))

    for i := 0; i < len(rows); i += batchSize {
        end := i + batchSize
        if end > len(rows) {
            end = len(rows)
        }

        batch := rows[i:end]
        var values []any
        var placeholderGroups []string

        for _, row := range batch {
            placeholderGroups = append(placeholderGroups, "("+strings.Join(placeholders, ", ")+")")
            for _, idx := range indices {
                if idx < len(row) {
                    values = append(values, convertValue(row[idx], colTypes[idx]))
                } else {
                    values = append(values, nil)
                }
            }
        }

        insertSQL := insertPrefix + strings.Join(placeholderGroups, ", ")
        if _, err := db.Exec(insertSQL, values...); err != nil {
            return fmt.Errorf("failed to insert rows: %w", err)
        }
    }

    return nil
}
```

## Integration Points

The driver requires one-line additions to 5 existing files. No new contracts, no query builder changes, no new interfaces.

### 1. Driver Registration

```go
// database/orm/orm.go — createDriver()
case "csvdir":
    return driver.NewCSVDir()
```

### 2. DSN Builder

```go
// database/db/config_builder.go — BuildDSN()
case "csvdir":
    return b.buildSQLiteDSN()
```

`buildSQLiteDSN` returns `ConnectionConfig.Database` if non-empty, else `:memory:`. For `csvdir`, `Database` holds the directory path. The directory path becomes the DSN, which `CSVDir.Open` receives.

### 3. Config Validation

```go
// database/db/config_builder.go — ConnectionConfig.Validate()
case "csvdir":
    // directory path is optional; empty returns empty in-memory DB
    return nil
```

### 4. Connection Pool Configuration

```go
// database/orm/orm.go — configureConnectionPool()
pinSingleConn := connConfig.Driver == "sqlite" || connConfig.Driver == "array" || connConfig.Driver == "csvdir"
```

In-memory SQLite requires single-connection pinning (same as the array driver).

### 5. detectDatabaseName

```go
// database/db.go — detectDatabaseName()
case "sqlite", "turso", "array", "csvdir":
    return "main"
```

### What is NOT needed

- **Placeholder funcs**: `Dialect()` returns `"sqlite"`, so `GetPlaceholderFunc("sqlite")` is used automatically. No new entry in `PlaceholderFuncs`.
- **Model() hook**: `Dialect()` returns `"sqlite"` (not `"array"`), so the `ArraySource` population hook at `query_model.go:18` does not fire. Tables are already populated at `Open` time — no lazy population needed.
- **isSQLite() check**: `Dialect()` returns `"sqlite"`, so `isSQLite()` at `query_constructors.go:86` returns `true` automatically. No change needed.
- **Close() cleanup**: The driver is stateless (no `sync.Map`). The `Close()` code at `orm.go:532-535` checks for `*driver.Array` — `*CSVDir` is not `*Array`, so it skips cleanup. Correct behavior: closing the `*sql.DB` drops the in-memory database entirely.
- **New contracts**: No new interfaces in `contracts/database/orm/`. The driver implements the existing `driver.Driver` interface.

## How It Works — Execution Flow

### 1. `neat.New(config)` — startup

```
neat.New(config)
  → BuildOrm(ctx, dbConfig, "csv_db", ...)
    → buildQuery(...)
      → createDriver("csvdir")           → *CSVDir{SQLite: &SQLite{}}
      → BuildDSN(connConfig)              → "data/" (the Database field)
      → CSVDir.Open("data/")
        → sql.Open("sqlite", ":memory:")  → in-memory SQLite *sql.DB
        → os.ReadDir("data/")
        → for each .csv file:
            populateCSVFile(db, "users", "data/users.csv")
              → parseCSV("data/users.csv")
                  → header: ["id", "name", "email", "active", "created"]
                  → rows: [["1", "Alice", ...], ["2", "Bob", ...], ...]
              → inferColumnTypes(columns, rows)
                  → ["INTEGER", "TEXT", "TEXT", "INTEGER", "DATETIME"]
              → CREATE TABLE "users" ("active" INTEGER, "created" DATETIME, "email" TEXT, "id" INTEGER, "name" TEXT)
              → INSERT INTO "users" ("active", "created", "email", "id", "name") VALUES (?, ?, ?, ?, ?), ...
            populateCSVFile(db, "products", "data/products.csv")
              → same process → "products" table
            populateCSVFile(db, "orders", "data/orders.csv")
              → same process → "orders" table
        → return *sql.DB (all tables populated)
      → configureConnectionPool(...)       → SetMaxOpenConns(1), PRAGMAs
      → NewQuery(ctx, sqlDB, csvDirDriver, ...)
```

### 2. `database.Query().Model(&User{}).Where("active = ?", true).Get(&users)` — query

```
Query()
  → returns *Query with q.driver = *CSVDir (Dialect() = "sqlite")
Model(&User{})
  → resolveTableName(&User{}) → "users" (from User.TableName() or struct name)
  → Dialect() == "array"? NO → no ArraySource hook fires
  → table "users" already exists in SQLite (populated at Open time)
Where("active = ?", true)
  → adds WHERE clause
Get(&users)
  → generates: SELECT * FROM "users" WHERE "active" = ?
  → SQLite executes against the pre-populated table
  → scans results into []User
```

No flat-file code runs during the query. The driver is just SQLite from the query builder's perspective.

### 3. `database.Close()` — shutdown

```
Close()
  → drv.(*driver.Array)? NO → no Cleanup call (correct — no sync.Map state)
  → db.Close() → in-memory SQLite dropped → all tables gone
```

## Example Usage

### Basic

```go
package main

import (
    "fmt"
    "log"

    "github.com/dracory/neat"
    _ "modernc.org/sqlite"
)

type User struct {
    ID     int
    Name   string
    Email  string
    Active bool
}

func (u *User) TableName() string { return "users" }

func main() {
    config := neat.DBConfig{
        Default: "csv_db",
        Connections: map[string]neat.ConnectionConfig{
            "csv_db": {
                Driver:   "csvdir",
                Database: "data/",
            },
        },
    }

    database, err := neat.New(config)
    if err != nil {
        log.Fatal(err)
    }
    defer database.Close()

    var users []User
    err = database.Query().
        Model(&User{}).
        Where("active = ?", true).
        OrderBy("name", "asc").
        Get(&users)
    if err != nil {
        log.Fatal(err)
    }

    for _, u := range users {
        fmt.Printf("%s <%s>\n", u.Name, u.Email)
    }
}
```

### Directory structure

```
data/
  users.csv
  products.csv
  orders.csv
```

### Sample CSV files

#### data/users.csv

```csv
id,name,email,active,created
1,Alice,alice@example.com,true,2024-01-15T10:30:00Z
2,Bob,bob@example.com,false,2024-02-20T14:45:00Z
3,Charlie,charlie@example.com,true,2024-03-10T09:00:00Z
```

#### data/products.csv

```csv
id,name,price,category
1,Widget,19.99,Hardware
2,Gadget,49.99,Electronics
3,Gizmo,99.99,Electronics
```

#### data/orders.csv

```csv
id,user_id,product_id,quantity,total
1,1,2,3,149.97
2,3,1,5,99.95
3,1,3,1,99.99
```

### JOINs across CSV files

Since all CSV files are populated as tables in the same in-memory SQLite database, JOINs work:

```go
type OrderWithUser struct {
    ID       int
    UserName string
    Total    float64
}

var results []OrderWithUser
err := database.Query().
    Table("orders AS o").
    Join("LEFT JOIN users AS u ON o.user_id = u.id", "").
    Select("o.id, u.name AS user_name, o.total").
    Get(&results)
```

This generates:
```sql
SELECT o.id, u.name AS user_name, o.total
FROM "orders" AS o
LEFT JOIN "users" AS u ON o.user_id = u.id
```

## Implementation Plan

### Phase 1: Driver Core
1. Create `database/driver/csvdir.go` — `CSVDir` struct, `NewCSVDir()`, `Dialect()`, `Open()`
2. Create `database/driver/csvdir_csv.go` — `parseCSV()`, `inferColumnTypes()`, `inferValueType()`, `widenType()`, `convertValue()`, `populateCSVFile()`
3. Add `isSimpleIdentifier` access (already a method on `*Array` — extract to package-level function or duplicate)

### Phase 2: Integration Points
1. Add `case "csvdir"` to `createDriver()` in `database/orm/orm.go`
2. Add `case "csvdir"` to `BuildDSN()` in `database/db/config_builder.go`
3. Add `case "csvdir"` to `Validate()` in `database/db/config_builder.go`
4. Add `"csvdir"` to `pinSingleConn` check in `configureConnectionPool()` in `database/orm/orm.go`
5. Add `"csvdir"` to `detectDatabaseName()` in `database/db.go`

### Phase 3: Tests
1. Driver unit tests in `database/driver/csvdir_test.go`:
   - Open with valid directory → all CSV files become tables
   - Open with empty directory → empty in-memory DB, no error
   - Open with non-existent directory → error
   - Open with file path (not directory) → error
   - Open with empty string → empty in-memory DB
   - Dialect() returns "sqlite"
   - Table names match filenames (without .csv)
   - Column names from CSV header
   - Type inference: int, float, bool, time, string
   - Type widening: mixed int/float column → REAL
   - Type widening: mixed int/string column → TEXT
   - Empty CSV (header only) → table created, zero rows
   - Empty cells → NULL
   - Invalid column name in header → error
   - Non-.csv files in directory → skipped
   - Subdirectories → skipped
   - Case-insensitive .csv extension (.CSV, .Csv)
2. CSV parser tests in `database/driver/csvdir_csv_test.go`:
   - `parseCSV`: basic parsing, empty file, header only
   - `inferValueType`: int, float, bool, time (RFC3339), time (date only), string
   - `inferColumnTypes`: all same type, mixed types, empty cells
   - `widenType`: INTEGER→REAL, REAL→INTEGER, INTEGER→TEXT, DATETIME→TEXT
   - `convertValue`: int, float, bool true/false, time, empty string→nil
3. Integration test in `database/csvdir_integration_test.go`:
   - Full flow: `neat.New(config)` → `Query().Model().Get()`
   - WHERE, ORDER BY, LIMIT
   - JOIN across two CSV-backed tables
   - Aggregate queries (COUNT, SUM, AVG)
4. Example in `examples/csvdir-driver/` with sample CSV files

### Phase 4: Documentation
1. Create `examples/csvdir-driver/README.md`
2. Add `csvdir` driver to main docs (driver-registration page, API reference)
3. Document type inference rules and directory structure expectations

## Design Decisions

### Why a New Driver Instead of Array Source Adapters?

The `flat-file-driver.md` proposal (revised August 9, 2026) describes array source adapters where each file is a separate `Source` implementing `ArraySource`. That approach requires the user to create a `Source` per file and pass it to `Model()`:

```go
database.Query().Model(flatfile.NewFromCSV("data/users.csv", "users")).Get(&users)
```

The CSV directory driver takes a different, simpler approach for the common case: point at a directory, get a database. No per-file configuration. No `Model()` argument changes. The user passes a regular model struct, and the table is already there.

The two approaches are complementary:
- **CSV directory driver** (this proposal): directory = database, file = table. Best for "I have a folder of CSVs, query them."
- **Array source adapters** (flat-file-driver.md): per-file sources with format-specific options. Best for "I want this one CSV with a custom delimiter and this one JSON file."

### Why Eager Population at Open Time?

All CSV files are parsed and loaded into SQLite when `Open` is called, not when a table is first queried. This means:
- **Startup cost**: proportional to the total size of all CSV files
- **Query speed**: fast — no parsing during queries, pure SQLite
- **Simplicity**: no lazy-loading logic, no `sync.Map`, no `Model()` hook changes

For the target use case (config files, exports, test fixtures, small-to-medium datasets), eager loading is fine. A directory of 10 CSV files with 1,000 rows each loads in milliseconds.

For large directories, lazy loading can be added as a future enhancement (see Future Enhancements).

### Why Dialect() Returns "sqlite"

The query builder uses `Dialect()` to determine SQL syntax, placeholder style, and feature availability. By returning `"sqlite"`:
- `isSQLite()` returns true → SQLite SQL syntax, `?` placeholders, JSON functions available
- `isMySQL()`, `isPostgres()`, etc. return false → no wrong-dialect code paths
- The `ArraySource` `Model()` hook (which checks `Dialect() == "array"`) does not fire → no conflict with pre-populated tables
- No changes to `query_constructors.go`, `builder_quote.go`, `to_sql.go`, or any query builder code

The driver is SQLite with pre-populated tables. The query builder doesn't need to know that the data came from CSV files.

### Why No Primary Key Support (Yet)

CSV files don't have inherent primary key metadata. Adding PK support would require either:
- A convention (e.g., first column named `id` is the PK)
- A sidecar file (e.g., `users.schema.json` alongside `users.csv`)
- A config option

This proposal keeps the initial implementation simple: no PK constraints. All columns are regular columns. PK support can be added as a future enhancement.

### Why No sync.Map or Cleanup

The array driver uses `sync.Map` to track which tables have been populated per `*sql.DB` connection, enabling lazy population and preventing double-population in concurrent scenarios. The CSV directory driver doesn't need this because:
- All tables are populated at `Open` time (eager, not lazy)
- The driver is stateless — no per-connection state outside the SQLite database itself
- When `*sql.DB` is closed, the in-memory database is dropped entirely

This means `Close()` in `orm.go` doesn't need to call `Cleanup` for this driver. The existing type assertion `drv.(*driver.Array)` simply doesn't match `*CSVDir`, and the cleanup step is skipped — which is correct.

## Benefits

1. **Zero Boilerplate**: Point at a directory, query with regular model structs
2. **Full Query Builder**: All ORM features work (WHERE, JOIN, ORDER BY, aggregates, etc.)
3. **JOINs Across Files**: All CSV files are tables in the same SQLite database — JOINs work naturally
4. **Type Inference**: Automatic type detection (int, float, bool, time, string) from CSV content
5. **No New Contracts**: No new interfaces in `contracts/` — the driver implements the existing `Driver` interface
6. **No Query Builder Changes**: `Dialect() == "sqlite"` means the query builder works unchanged
7. **Minimal Integration**: 5 one-line additions to existing files
8. **Stateless Driver**: No `sync.Map`, no cleanup, no per-connection state
9. **Test Fixtures**: Easy to load test data from a directory of CSV files

## Risks and Mitigations

### Risk 1: Memory Usage for Large Directories
- **Issue**: Loading all CSV files into in-memory SQLite at startup could consume significant memory
- **Mitigation**: Target use case is small-to-medium datasets (config, exports, fixtures). For large datasets, lazy loading can be added as a future enhancement. A row limit can be added if needed.

### Risk 2: Type Inference Ambiguity
- **Issue**: CSV values like "123" could be int or string (e.g., zip codes with leading zeros)
- **Mitigation**: Type inference is permissive (tries int first). Users who need specific types can use the array source adapter approach with `WithSchema`. A future enhancement could add sidecar schema files (e.g., `users.schema.json`).

### Risk 3: File Encoding
- **Issue**: CSV files may use non-UTF-8 encodings (Latin-1, Windows-1252)
- **Mitigation**: Phase 1 supports UTF-8 only. Encoding support can be added as a future enhancement.

### Risk 4: Startup Time
- **Issue**: Parsing all CSV files at `Open` time adds startup latency
- **Mitigation**: For the target use case (small-to-medium datasets), this is milliseconds. For large directories, lazy loading (parse on first query to each table) can be added as a future enhancement.

### Risk 5: Header Drift
- **Issue**: A CSV file's header might change (columns added/removed/renamed) without the user noticing
- **Mitigation**: The driver creates the table from the current header at `Open` time. If the header doesn't match the model struct, the query builder's column mapping will fail with a clear error. A future enhancement could add header validation against expected schemas.

## Future Enhancements

1. **Lazy Loading**: Only parse a CSV file when its table is first queried, reducing startup time for large directories
2. **File Watching**: Watch the directory for changes and reload tables when files change (using `fsnotify`)
3. **Primary Key Convention**: Automatically mark a column named `id` or `{tablename}_id` as PRIMARY KEY
4. **Sidecar Schema Files**: Read `users.schema.json` alongside `users.csv` for explicit column types and PK declaration
5. **Custom Delimiters**: Support `.tsv` (tab-separated), `.psv` (pipe-separated) files in the same directory
6. **JSON Files**: Extend to `.json` and `.jsonl` files in the same directory (each becomes a table)
7. **Persistent Cache**: Cache the populated SQLite database to a file on disk, only re-parsing CSVs when files change (using mtime, like Sushi's stale-cache detection)
8. **Write Support**: Allow ORM operations (Create, Update, Delete) to write back to CSV files
9. **Streaming**: For very large CSV files, use SQLite's virtual table mechanism to stream rows on demand
10. **Compressed Files**: Support `.csv.gz` files with automatic decompression
11. **File Encoding**: Support non-UTF-8 encodings via a config option

## References

- SQLite driver implementation: `database/driver/sqlite.go`
- Array driver implementation: `database/driver/array.go`
- Driver interface: `database/driver/driver.go`
- Driver registration: `database/orm/orm.go` — `createDriver()`
- DSN builder: `database/db/config_builder.go` — `BuildDSN()`
- Connection pool config: `database/orm/orm.go` — `configureConnectionPool()`
- Database name detection: `database/db.go` — `detectDatabaseName()`
- Query builder Model() hook: `database/query/query_model.go`
- Query builder dialect checks: `database/query/query_constructors.go` — `isSQLite()`
- Go `encoding/csv` package: https://pkg.go.dev/encoding/csv
- SQLite in-memory databases: https://www.sqlite.org/inmemorydb.html
- Sushi (Laravel array driver): https://github.com/calebporzio/sushi
