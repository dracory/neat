# CSVDB Driver

**Date**: August 9, 2026
**Status**: Completed
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
            Driver:   "csvdb",
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

## Solution

A `CSVDB` driver that:

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
    csvdb.go              // CSVDB driver (embeds *SQLite, overrides Open)
    csvdb_csv.go          // CSV parsing + type inference helper
    csvdb_test.go         // Driver unit tests
    csvdb_csv_test.go     // CSV parser tests

integration_tests/
  csvdb/
    helper.go             // SetupCSVDBTest, CSV fixtures, model types
    csvdb_query_test.go   // Query, WHERE, ORDER BY, LIMIT, First, Find, Count
    csvdb_join_test.go    // LEFT JOIN, INNER JOIN, 3-table JOIN, JOIN+WHERE
    csvdb_aggregate_test.go  // COUNT, SUM, AVG, MIN/MAX, GROUP BY
    csvdb_type_inference_test.go  // Type inference verification via queries

examples/
  csvdb-driver/
    main.go               // Example usage
    main_test.go
    README.md
    data/
      users.csv           // Sample CSV
      products.csv        // Sample CSV
      orders.csv          // Sample CSV
```

No new contracts. No new interfaces. No changes to `contracts/database/orm/`. No changes to `database/query/`. The driver plugs into the existing driver registration system.

## Driver Implementation

```go
// database/driver/csvdb.go

package driver

import (
    "database/sql"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    _ "modernc.org/sqlite"
)

// CSVDB implements the Driver interface for CSV-directory-backed storage.
// It embeds *SQLite for all standard Driver methods and overrides Open to
// scan a directory of CSV files and populate an in-memory SQLite database.
type CSVDB struct {
    *SQLite
}

func NewCSVDB() *CSVDB {
    return &CSVDB{SQLite: NewSQLite()}
}

func (c *CSVDB) Dialect() string {
    return "sqlite"
}

func (c *CSVDB) Open(dirPath string) (*sql.DB, error) {
    db, err := sql.Open("sqlite", ":memory:")
    if err != nil {
        return nil, fmt.Errorf("csvdb: failed to open in-memory SQLite: %w", err)
    }

    if dirPath == "" || dirPath == ":memory:" {
        return db, nil
    }

    info, err := os.Stat(dirPath)
    if err != nil {
        db.Close()
        return nil, fmt.Errorf("csvdb: cannot access directory %s: %w", dirPath, err)
    }
    if !info.IsDir() {
        db.Close()
        return nil, fmt.Errorf("csvdb: %s is not a directory", dirPath)
    }

    entries, err := os.ReadDir(dirPath)
    if err != nil {
        db.Close()
        return nil, fmt.Errorf("csvdb: cannot read directory %s: %w", dirPath, err)
    }

    seenTables := make(map[string]string) // lower(tableName) → original filename

    for _, entry := range entries {
        if entry.IsDir() {
            continue
        }
        if !strings.HasSuffix(strings.ToLower(entry.Name()), ".csv") {
            continue
        }

        tableName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
        filePath := filepath.Join(dirPath, entry.Name())

        lowerName := strings.ToLower(tableName)
        if prevFile, exists := seenTables[lowerName]; exists {
            db.Close()
            return nil, fmt.Errorf("csvdb: table name collision: %s and %s produce the same table name (case-insensitive)", prevFile, entry.Name())
        }
        seenTables[lowerName] = entry.Name()

        if err := populateCSVFile(db, tableName, filePath); err != nil {
            db.Close()
            return nil, fmt.Errorf("csvdb: failed to populate table %s from %s: %w", tableName, filePath, err)
        }
    }

    return db, nil
}
```

### CSV Parsing and Type Inference

```go
// database/driver/csvdb_csv.go

// MaxCSVRows limits the number of data rows that can be loaded from a single
// CSV file to prevent unbounded memory/CPU consumption.
const MaxCSVRows = 100000

// parseCSV reads a CSV file and returns column names and rows.
// Strips UTF-8 BOM from the first header field. Reads rows incrementally
// and stops at MaxCSVRows without loading the entire file into memory.
func parseCSV(filePath string) (columns []string, rows [][]string, err error) {
    // ...opens file, reads header (strips BOM), reads rows incrementally
}

// inferValueType tries to determine the type of a string value.
// Order: int → float → bool → time → string.
// Rejects Inf/NaN (accepted by strconv.ParseFloat but almost always
// literal text in a CSV).
func inferValueType(val string) string {
    if _, err := strconv.ParseInt(val, 10, 64); err == nil {
        return "INTEGER"
    }
    if f, err := strconv.ParseFloat(val, 64); err == nil && !math.IsNaN(f) && !math.IsInf(f, 0) {
        return "REAL"
    }
    if val == "true" || val == "false" || val == "True" || val == "False" {
        return "INTEGER" // SQLite stores bools as 0/1
    }
    if _, err := time.Parse(time.RFC3339, val); err == nil {
        return "DATETIME"
    }
    for _, layout := range []string{"2006-01-02", "2006-01-02 15:04:05", "01/02/2006"} {
        if _, err := time.Parse(layout, val); err == nil {
            return "DATETIME"
        }
    }
    return "TEXT"
}

// widenType returns the wider of two SQLite types.
// INTEGER → REAL → TEXT is the widening chain.
// DATETIME is compatible only with itself and TEXT; any mix of DATETIME
// with INTEGER or REAL widens to TEXT.
func widenType(current, new string) string { /* ... */ }

// populateCSVFile reads a CSV file, infers types, creates a table, and
// inserts all rows in a transaction. Validates:
// - Table and column names are simple identifiers (SQL injection prevention)
// - No duplicate column names
// - No data row has more fields than the header (ragged row check)
func populateCSVFile(db *sql.DB, tableName string, filePath string) error {
    // 1. parseCSV → columns, rows
    // 2. Validate table name, column names, duplicate columns
    // 3. Validate ragged rows (more fields than header → error)
    // 4. Infer column types
    // 5. CREATE TABLE with inferred schema
    // 6. BEGIN TRANSACTION
    // 7. INSERT rows in batches (SQLite parameter limit ~999)
    // 8. COMMIT
}
```

### Type Inference

| CSV content              | SQLite type |
|--------------------------|-------------|
| `123`, `-456`            | INTEGER     |
| `19.99`, `-0.5`          | REAL        |
| `true`, `false`          | INTEGER (0/1) |
| `2024-01-15T10:30:00Z`   | DATETIME    |
| `2024-01-15`             | DATETIME    |
| `2024-01-15 10:30:00`    | DATETIME    |
| `01/02/2006`             | DATETIME    |
| `hello`, `abc123`        | TEXT        |
| `Inf`, `NaN`             | TEXT (not REAL) |

Mixed types in a column widen: INTEGER → REAL → TEXT. DATETIME mixed with INTEGER or REAL widens to TEXT.

## Integration Points

The driver required one-line additions to 5 existing files plus one case in the schema builder. No new contracts, no query builder changes, no new interfaces.

### 1. Driver Registration

```go
// database/orm/orm.go — createDriver()
case "csvdb":
    return driver.NewCSVDB()
```

### 2. DSN Builder

```go
// database/db/config_builder.go — BuildDSN()
case "csvdb":
    return b.buildCSVDBDSN()
```

`buildCSVDBDSN` returns `ConnectionConfig.Database` if non-empty, else `:memory:`. For CSVDB, `Database` holds the directory path. The directory path becomes the DSN, which `CSVDB.Open` receives.

### 3. Config Validation

```go
// database/db/config_builder.go — ConnectionConfig.Validate()
case "sqlite", "array", "csvdb":
    // database path is optional; empty defaults to :memory:
    // For CSVDB, the Database field holds the directory path.
    return nil
```

### 4. Connection Pool Configuration

```go
// database/orm/orm.go — configureConnectionPool()
pinSingleConn := connConfig.Driver == "sqlite" || connConfig.Driver == "array" || connConfig.Driver == "csvdb"
```

CSVDB uses in-memory SQLite, so it shares the same single-connection constraint as SQLite and the array driver.

### 5. Database Name Detection

```go
// database/db.go — detectDatabaseName()
case "sqlite", "turso", "array", "csvdb":
    return "main"
```

### 6. Schema Builder

```go
// database/schema/schema.go — New()
case contractsdatabase.DriverCSVDB:
    // CSVDB uses SQLite grammar since Dialect() returns "sqlite".
    sqliteGrammar := grammars.NewSqlite(log, prefix)
    driverSchema = NewSqliteSchema(sqliteGrammar, orm, prefix)
    grammar = sqliteGrammar
    processor = processors.NewSqlite()
```

The schema builder is rarely used for CSVDB (tables are created at Open time from CSV files), but it must not error during `New()`.

## Completed Features

All features from the original proposal plus the following enhancements added during implementation and code review:

### Robustness
- **UTF-8 BOM stripping**: CSV files exported from Excel commonly include a BOM (`\xEF\xBB\xBF`). The parser strips it from the first header field to avoid corrupting the first column name.
- **Duplicate column detection**: A CSV header like `id,name,id` produces a clear error (`duplicate column name in CSV header: id`) instead of a SQLite error at CREATE TABLE time.
- **Ragged row validation**: Data rows with more fields than the header are rejected with a clear error. Rows with fewer fields are allowed (missing fields become NULL).
- **Table name collision detection**: On case-sensitive filesystems (Linux), `Users.csv` and `users.csv` would produce colliding SQLite table names (SQLite table names are case-insensitive). The driver detects this and returns an error.
- **MaxCSVRows limit**: 100,000-row limit per CSV file to prevent unbounded memory/CPU consumption. Rows are read incrementally so the limit is enforced without loading the entire file.
- **Inf/NaN rejection**: `strconv.ParseFloat` accepts `"Inf"`, `"-Inf"`, and `"NaN"`, but these are almost always literal text in a CSV. The type inferencer rejects them, classifying them as TEXT instead of REAL.

### Performance
- **Transaction wrapping**: Batch INSERTs are wrapped in a `BEGIN`/`COMMIT` transaction for atomicity and performance (single commit vs. N commits).
- **Incremental CSV reading**: Uses `reader.Read()` in a loop instead of `reader.ReadAll()`, so the `MaxCSVRows` limit is enforced without parsing the entire file first.

### Type Inference
- **Empty initial type**: Columns start with an empty type (not `INTEGER`), so a time-only column is correctly inferred as `DATETIME` instead of widening from `INTEGER` to `TEXT`.
- **Explicit DATETIME widening**: `widenType` explicitly handles `DATETIME` mixed with `INTEGER`/`REAL` → `TEXT`, instead of relying on the fallback.
- **Consistent int64 typing**: `convertValue` returns `int64` for both integers and bools (previously returned `int` for bools, causing inconsistent typing).

### Code Quality
- **`isSimpleIdentifier` refactoring**: Extracted to a package-level function shared by both the CSVDB and Array drivers, avoiding code duplication.
- **Dedicated `buildCSVDBDSN`**: Semantically distinct from `buildSQLiteDSN` (directory path vs. file path), even though the logic is identical.

## Testing

### Unit Tests (`database/driver/`)
- CSV parsing, type inference, type widening, value conversion
- BOM stripping, duplicate columns, ragged rows, MaxCSVRows limit
- Inf/NaN rejection, table name collision, case-insensitive extension
- Empty CSV (header only), empty cells become NULL, invalid column names
- Non-CSV files skipped, subdirectories skipped

### Integration Tests (`integration_tests/csvdb/`)
- **Query**: basic query, WHERE equals, WHERE bool, ORDER BY asc/desc, LIMIT, LIMIT+OFFSET, WHERE greater than, First, Find, Count
- **JOIN**: LEFT JOIN, INNER JOIN, 3-table JOIN, JOIN with WHERE
- **Aggregates**: COUNT, COUNT with WHERE, SUM, AVG, MIN/MAX, GROUP BY, SUM with JOIN+GROUP BY
- **Type inference**: integer, float, bool, string inference via queries, PRAGMA column type verification, NULL cell handling

### Example (`examples/csvdb-driver/`)
- Working example with sample CSV data (users, products, orders)
- Demonstrates WHERE, JOIN, and aggregate queries
