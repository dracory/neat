# Flat-File Driver

**Date**: June 24, 2026
**Status**: Proposal
**Priority**: Medium
**Supersedes**: `csv-driver.md`, `json-driver.md`

## Problem Statement

The array driver allows querying in-memory data via SQLite, but it requires the user to manually define `Rows()` as `[]map[string]any`. There is no built-in way to query flat-file data formats — CSV, JSON, JSONL/NDJSON, YAML, or Markdown with frontmatter — without first parsing and converting them into Go data structures.

Users who have flat-file data (exports, config, test fixtures, datasets, content files) must write boilerplate parsing code before they can use the ORM's query builder.

### Current Limitations

- No native flat-file support — users must parse files themselves and feed rows into the array driver
- No automatic type inference from file content (CSV needs string parsing; JSON has native types but no ORM integration)
- No support for format-specific features (CSV delimiters/headers, JSON nested objects/JSONL, YAML frontmatter)
- No directory-based mode (one file per record, like Orbit and Laravel Paper)
- No file-metadata timestamps (e.g., file mtime as `updated_at`)
- No soft delete support for file-backed records
- No primary key declaration for file-backed tables
- No custom column names for headerless files
- No header/schema validation to detect drift
- No streaming for large files

### Example of the Desired Experience

```go
config := neat.DBConfig{
    Default: "ff_db",
    Connections: map[string]neat.ConnectionConfig{
        "ff_db": {
            Driver: "flatfile",
        },
    },
}

database, _ := neat.New(config)
defer database.Close()

// Query a CSV file directly
var users []User
err := database.Query().
    Model(&UserSource{FilePath: "data/users.csv"}).
    Where("country = ?", "US").
    OrderBy("name", "asc").
    Get(&users)

// Query a JSON file directly
var products []Product
err = database.Query().
    Model(&ProductSource{FilePath: "data/products.json"}).
    Where("price > ?", 50).
    Get(&products)

// Query a directory of JSON files (one file per record)
var posts []Post
err = database.Query().
    Model(&PostSource{FilePath: "content/posts"}).
    Where("published = ?", true).
    OrderBy("updated_at", "desc").
    Get(&posts)
```

## Proposed Solution

Implement a single `flatfile` driver that mirrors the array driver architecture: it uses an in-memory SQLite database under the hood. Instead of requiring `Rows()` from the user, it reads and parses flat files at populate time using **pluggable format parsers**.

The driver handles all shared concerns (directory mode, timestamps, soft deletes, primary key, schema inference, batched inserts, concurrency). Format-specific parsing is delegated to `FileParser` implementations — built-in parsers for CSV and JSON, with extensibility for YAML, Markdown, and custom formats.

### Architecture

```
contracts/
  database/
    orm/
      flatfile_source.go       // FlatFileSource, FlatFileSchema, FlatFileDir,
                               // FlatFilePrimaryKey, FlatFileTimestamps,
                               // FlatFileSoftDeletes, FlatFileOptions,
                               // FlatFileConfig, FlatFilePopulator,
                               // FileParser, FileParserRegistry

database/
  driver/
    flatfile.go                // FlatFile driver (embeds *SQLite, orchestrates)
    flatfile_csv.go            // CSV parser (implements FileParser)
    flatfile_json.go           // JSON parser (implements FileParser)
    flatfile_test.go           // Driver unit tests
    flatfile_csv_test.go       // CSV parser tests
    flatfile_json_test.go      // JSON parser tests
  flatfile_integration_test.go // Integration test via Database.Query()

examples/
  flatfile-driver/
    main.go                    // Example usage (CSV, JSON, directory mode)
    main_test.go               // Example test
    README.md                  // Documentation
    data/
      users.csv                // Sample CSV (single-file mode)
      products.json            // Sample JSON (single-file mode)
      events.jsonl             // Sample JSONL (single-file mode)
      posts/                   // Sample directory mode (one file per record)
        hello-world.json
        my-second-post.json
```

### Interface Design

```go
// contracts/database/orm/flatfile_source.go

package orm

import (
    "context"
    "database/sql"
    "io"
)

// ----------------------------------------------------------------------------
// Source Interfaces (implemented by user models)
// ----------------------------------------------------------------------------

// FlatFileSource is implemented by any model that wants flat-file-backed storage.
// The driver detects the format from the file extension (or via FlatFileOptions)
// and delegates parsing to the appropriate FileParser.
//
// In single-file mode, FilePath() points to a data file.
// In directory mode (when FlatFileDir is implemented), FilePath() points to
// a directory containing one file per record.
type FlatFileSource interface {
    TableName() string
    FilePath() string
}

// FlatFileDir is an optional interface that enables directory mode.
// Instead of reading a single file with all rows, the driver reads all
// files matching the configured extension in the specified directory.
// Each file is one record. The filename (without extension) is used as
// the primary key column value.
//
// This is similar to Orbit's file-per-record approach and Laravel Paper's
// slug-as-filename pattern.
type FlatFileDir interface {
    IsDirectory() bool // if true, FilePath() points to a directory
}

// FlatFileSchema is an optional interface for specifying column types explicitly.
// If not implemented, the driver infers types from the parsed data.
type FlatFileSchema interface {
    Schema() map[string]string // column -> type ("string", "int", "float", "bool", "time")
}

// FlatFileColumns is an optional interface for providing custom column names.
// For CSV: used when HasHeader is false (overrides generated col_N names).
// For JSON: overrides JSON keys as column names.
// For directory mode: provides names for flattened nested fields.
type FlatFileColumns interface {
    Columns() []string // custom column names
}

// FlatFilePrimaryKey is an optional interface for declaring a primary key column.
// The driver adds a PRIMARY KEY constraint to the declared column when
// creating the SQLite table.
//
// In directory mode, the filename (without extension) is automatically
// inserted as the value for this column. For example, a file named
// "hello-world.json" with PrimaryKey() = "slug" produces slug = "hello-world".
type FlatFilePrimaryKey interface {
    PrimaryKey() string // column name to mark as PRIMARY KEY
}

// FlatFileTimestamps is an optional interface that enables file-metadata
// timestamps. When enabled, the driver adds two columns:
//   - created_at: set to the file's creation time (or mtime if ctime unavailable)
//   - updated_at: set to the file's modification time (mtime)
//
// These values come from the filesystem, not from the file content.
// This is similar to Laravel Paper's #[Timestamps] attribute.
//
// In directory mode, each record gets its own timestamps from its individual file.
// In single-file mode, all rows share the same timestamp (the file's mtime).
type FlatFileTimestamps interface {
    Timestamps() bool // if true, add created_at and updated_at from file metadata
}

// FlatFileSoftDeletes is an optional interface that enables soft delete
// support. When enabled, the driver checks for a "deleted_at" field/column
// in the parsed data. If present and non-null/non-empty, the record is
// considered soft-deleted and excluded from query results by default.
//
// This is similar to Orbit's SoftDeletes trait.
type FlatFileSoftDeletes interface {
    SoftDeletes() bool // if true, respect deleted_at for soft deletion
}

// FlatFileOptions is an optional interface for customizing parsing behavior.
// The returned FlatFileConfig applies to both the driver and the format parser.
type FlatFileOptions interface {
    Options() FlatFileConfig
}

// FlatFileConfig holds shared and format-specific parsing configuration.
// Format-specific fields are used only by the relevant parser.
type FlatFileConfig struct {
    // Shared
    FileExtension string // File extension for directory mode (default: auto-detect from FilePath)

    // CSV-specific
    Comma         rune   // Field delimiter (default: ',')
    Comment       rune   // Comment character (default: 0 = no comments)
    HasHeader     bool   // Whether the first row is a header (default: true)
    SkipRows      int    // Number of rows to skip before header/data (default: 0)
    NullIf        string // Treat this string as NULL (default: "")
    TrimSpace     bool   // Trim leading/trailing whitespace from fields (default: false)
    LazyQuotes    bool   // Allow quotes to appear in unquoted fields and vice versa (default: false)

    // JSON-specific
    RootPath      string // Dot-separated path to row array in nested JSON (e.g., "data.users")
    IsJSONL       bool   // Parse as JSONL/NDJSON (one object per line)
    Flatten       bool   // Flatten nested objects with dot notation
    FlattenDepth  int    // Max nesting depth for flattening (0 = unlimited; default: 3)
    NullIfMissing bool   // Treat missing keys as NULL (default: false, missing keys default to zero value)
    TrimStrings   bool   // Trim leading/trailing whitespace from string values (default: false)

    // Header validation (CSV)
    ExpectedHeaders []string // Headers the file must contain (CSV only)
    StrictHeaders   bool     // If true, error if headers don't match exactly (CSV only)
}

// ----------------------------------------------------------------------------
// Parser Interface (implemented by format-specific parsers)
// ----------------------------------------------------------------------------

// FileParser is implemented by each format parser (CSV, JSON, YAML, etc.).
// The driver calls these methods to read data from files.
//
// Parsers are registered via FileParserRegistry and selected based on file
// extension or explicit configuration.
type FileParser interface {
    // Format returns the parser's format name (e.g., "csv", "json", "yaml").
    Format() string

    // Extensions returns the file extensions this parser handles (e.g., [".csv", ".tsv"]).
    Extensions() []string

    // ParseFile reads a single file and returns all rows.
    // Each row is a map[string]any where keys are column names and values are
    // Go-native types (int, float64, bool, string, time.Time, nil, or nested maps/slices).
    //
    // config provides format-specific options.
    // In directory mode, this is called once per file (each file = one record).
    // In single-file mode, this is called once for the entire file.
    ParseFile(reader io.Reader, config FlatFileConfig) ([]map[string]any, error)
}

// FileParserRegistry manages available format parsers.
// Parsers are registered at init time or via RegisterParser().
type FileParserRegistry interface {
    Register(parser FileParser)
    Get(format string) (FileParser, bool)
    GetByExtension(ext string) (FileParser, bool)
    Formats() []string
}

// ----------------------------------------------------------------------------
// Populator Interface (implemented by the driver)
// ----------------------------------------------------------------------------

// FlatFilePopulator is implemented by the flatfile driver.
type FlatFilePopulator interface {
    Populate(ctx context.Context, db *sql.DB, source FlatFileSource) error
}
```

### Driver Implementation

```go
// database/driver/flatfile.go

package driver

import (
    "context"
    "database/sql"
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"
    "sync"

    contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// FlatFile implements the Driver interface for flat-file-backed storage using SQLite.
// It delegates format-specific parsing to FileParser implementations.
type FlatFile struct {
    *SQLite
    populated sync.Map // map[string]bool, key is "dbPointer-tableName"
    locks     sync.Map // map[string]*sync.Mutex, key is "dbPointer-tableName"
    locksMu   sync.Mutex
    parsers   FileParserRegistry
}

// NewFlatFile creates a new FlatFile driver with built-in CSV and JSON parsers.
func NewFlatFile() *FlatFile {
    registry := NewParserRegistry()
    registry.Register(&csvParser{})
    registry.Register(&jsonParser{})
    return &FlatFile{
        SQLite:  NewSQLite(),
        parsers: registry,
    }
}

// Dialect returns the dialect name.
func (f *FlatFile) Dialect() string {
    return "flatfile"
}

// MaxFlatFileRows limits the number of rows that can be populated from a single
// file (or directory) to prevent unbounded memory/CPU consumption.
const MaxFlatFileRows = 100000

// Populate reads flat file(s), creates an in-memory SQLite table, and inserts
// the parsed rows. It follows the same pattern as Array.Populate.
func (f *FlatFile) Populate(ctx context.Context, db *sql.DB, source contractsorm.FlatFileSource) error {
    // 1. Validate table name and file path
    // 2. Check if already populated (sync.Map, same as Array)
    // 3. Acquire per-table mutex
    // 4. Determine parser:
    //    a. If FlatFileOptions specifies a format, use that parser
    //    b. Otherwise, auto-detect from file extension
    // 5. Determine mode:
    //    a. If FlatFileDir.IsDirectory(): directory mode
    //    b. Otherwise: single-file mode
    // 6. Parse data:
    //    Directory mode:
    //      - Read all files matching the configured extension in the directory
    //      - Each file is one record (call parser.ParseFile)
    //      - Filename (without extension) becomes the primary key value
    //      - If FlatFileTimestamps: read file mtime/ctime for created_at/updated_at
    //      - If FlatFileSoftDeletes: check for deleted_at field, skip if soft-deleted
    //    Single-file mode:
    //      - Open file and call parser.ParseFile
    //      - If FlatFileTimestamps: read file mtime for all rows
    //      - If FlatFileSoftDeletes: check for deleted_at field in each row
    // 7. Resolve column names:
    //    a. If FlatFileColumns: use custom names
    //    b. Otherwise: merge all keys from all rows (like array driver)
    // 8. Handle format-specific post-processing:
    //    a. CSV: header validation if FlatFileConfig.StrictHeaders is true
    //    b. JSON: nested object flattening if FlatFileConfig.Flatten is true
    // 9. Infer or use explicit schema (FlatFileSchema)
    // 10. Determine primary key if source implements FlatFilePrimaryKey
    // 11. Create SQLite table (with PRIMARY KEY constraint if declared)
    // 12. Insert rows in batches
    // 13. Mark as populated
    // ...
}

// inferSchema infers column types from Go-native values in the parsed rows.
// For JSON, types are native (int, float64, bool, string, nil).
// For CSV, the parser converts strings to Go types during parsing.
// Type widening applies the same rules as the array driver.
func (f *FlatFile) inferSchema(rows []map[string]any) (map[string]string, error) {
    // ...
}

// readDirectory reads all files from a directory using the specified parser
// and returns them as rows. The filename (without extension) is set as the
// primary key value. If timestamps is true, file mtime/ctime are included.
func (f *FlatFile) readDirectory(dirPath string, parser contractsorm.FileParser, config contractsorm.FlatFileConfig, pkCol string, timestamps, softDeletes bool) ([]map[string]any, error) {
    // ...
}

// detectParser determines which parser to use based on file extension or config.
func (f *FlatFile) detectParser(filePath string, config contractsorm.FlatFileConfig) (contractsorm.FileParser, error) {
    // ...
}

// Cleanup removes cached entries — same pattern as Array.Cleanup.
func (f *FlatFile) Cleanup(db *sql.DB) {
    // ...
}
```

### Built-in Parsers

#### CSV Parser

```go
// database/driver/flatfile_csv.go

package driver

import (
    "encoding/csv"
    "io"
    "strconv"
    "strings"
    "time"

    contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// csvParser implements FileParser for CSV files.
type csvParser struct{}

func (p *csvParser) Format() string         { return "csv" }
func (p *csvParser) Extensions() []string   { return []string{".csv", ".tsv"} }

func (p *csvParser) ParseFile(reader io.Reader, config contractsorm.FlatFileConfig) ([]map[string]any, error) {
    // 1. Create csv.Reader with config (Comma, Comment, LazyQuotes)
    // 2. If HasHeader: read first row as column names
    //    If !HasHeader and FlatFileColumns: use custom names
    //    If !HasHeader and no FlatFileColumns: generate col_0, col_1, ...
    // 3. If StrictHeaders: validate against ExpectedHeaders
    // 4. Read all records
    // 5. Convert string values to Go-native types:
    //    - Skip empty strings and NullIf values during type detection
    //    - Try int → float64 → bool → time → string
    // 6. Apply TrimSpace if configured
    // 7. Return []map[string]any
    // ...
}
```

**CSV-specific features**:
- Custom delimiters (comma, tab, semicolon)
- Headerless mode with custom or generated column names
- Header validation (strict mode for schema drift detection)
- `SkipRows` for files with preamble lines
- `NullIf` for treating specific strings as NULL
- `LazyQuotes` for messy CSV quoting
- String-to-type inference (all CSV values are strings; parser converts)

#### JSON Parser

```go
// database/driver/flatfile_json.go

package driver

import (
    "bufio"
    "encoding/json"
    "io"
    "strings"

    contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// jsonParser implements FileParser for JSON and JSONL files.
type jsonParser struct{}

func (p *jsonParser) Format() string         { return "json" }
func (p *jsonParser) Extensions() []string   { return []string{".json", ".jsonl"} }

func (p *jsonParser) ParseFile(reader io.Reader, config contractsorm.FlatFileConfig) ([]map[string]any, error) {
    // 1. Determine mode:
    //    a. If IsJSONL: read line by line with bufio.Scanner, parse each as JSON object
    //    b. Otherwise: parse entire file as JSON
    // 2. If RootPath: navigate to the specified path (e.g., "data.users")
    // 3. Expect an array of objects at the target path
    // 4. If Flatten: flatten nested objects with dot notation up to FlattenDepth
    //    Otherwise: store nested objects/arrays as JSON strings
    // 5. Apply NullIfMissing (missing keys → nil) or default (missing keys → zero value)
    // 6. Apply TrimStrings if configured
    // 7. Return []map[string]any with Go-native types
    // ...
}

// navigatePath traverses a nested JSON object using a dot-separated path.
func navigatePath(data any, path string) (any, error) { ... }

// flattenObject flattens nested objects using dot notation up to maxDepth.
func flattenObject(prefix string, obj map[string]any, depth, maxDepth int) map[string]any { ... }
```

**JSON-specific features**:
- Native type inference (JSON has int, float, bool, string, null — no string parsing needed)
- JSONL/NDJSON support (one object per line, streaming-friendly)
- Root path extraction (navigate to `data.users` in nested JSON)
- Nested object flattening with depth control (dot notation: `user.name`)
- Missing key handling (`NullIfMissing` for SQL NULL vs zero values)
- Nested objects/arrays stored as JSON strings by default (queryable via SQLite JSON functions)

### Type Inference

Type inference happens in two stages:

**Stage 1 — Parser converts to Go-native types**:
- CSV: String values are parsed (try int → float → bool → time → string)
- JSON: Native types are preserved directly (no parsing needed)

**Stage 2 — Driver infers SQLite column types from Go-native values**:

```
Go type               → SQLite type
──────────────────────────────────────
nil                   → (skip, column defaults to TEXT if all nil)
int, int64            → INTEGER
float64               → REAL
bool                  → INTEGER
string                → TEXT
string (time format)  → DATETIME (detected via parsing)
time.Time             → DATETIME
map, []any            → TEXT (stored as JSON string)
```

Type widening applies the same rules as the array driver:
- `INTEGER` + `REAL` → `REAL`
- Any incompatible mix → `TEXT`

### Directory Mode (One File Per Record)

In directory mode (enabled via `FlatFileDir`), each file in the directory is one record. This is the approach used by both **Orbit** and **Laravel Paper**:

- **Orbit**: Stores each record as a separate file (`.md`, `.json`, `.yaml`) in a content directory. The filename is derived from the primary key.
- **Laravel Paper**: The filename (without extension) becomes the slug, which is the primary key. E.g., `content/posts/hello-world.md` → slug = `"hello-world"`.

The neat flat-file driver applies the same pattern:

```
content/posts/
  ├── hello-world.json      → slug = "hello-world"
  ├── my-second-post.json   → slug = "my-second-post"
  └── draft-post.json       → slug = "draft-post"
```

The filename (without extension) is automatically inserted as the value for the primary key column declared via `FlatFilePrimaryKey`. If no `FlatFilePrimaryKey` is declared, a default `slug` column is used.

The parser is selected per-file based on extension, so a directory could theoretically contain mixed formats (though this is uncommon and not recommended).

Directory mode is particularly useful for:
- **Content management**: Blog posts, pages, documentation — each content piece is a separate file
- **Configuration**: Each config entity in its own file for easier editing and version control
- **Git-friendly workflows**: Individual files diff cleanly in version control

### File-Metadata Timestamps

When `FlatFileTimestamps` is enabled, the driver adds `created_at` and `updated_at` columns derived from the filesystem:

- **`updated_at`**: Set to the file's modification time (`mtime`)
- **`created_at`**: Set to the file's creation time (`ctime`), or falls back to `mtime` if the platform doesn't support `ctime`

This is similar to Laravel Paper's `#[Timestamps]` attribute. Note that Git checkouts reset mtimes to the deploy time, so for Git-deployed content, a date field in the file content itself is more reliable.

In directory mode, each record gets its own timestamps from its individual file. In single-file mode, all rows share the same timestamp (the file's mtime).

### Soft Deletes

When `FlatFileSoftDeletes` is enabled, the driver checks for a `deleted_at` field in each parsed record. If present and non-null/non-empty, the record is considered soft-deleted:

- The record is still loaded into the SQLite table (with `deleted_at` populated)
- A default `WHERE deleted_at IS NULL` filter is applied to all queries
- Users can explicitly query soft-deleted records with `Where("deleted_at IS NOT NULL")`

This is similar to Orbit's `SoftDeletes` trait. The neat ORM's existing soft delete infrastructure can be leveraged for this feature.

### Custom Parsers

Users can register custom parsers for additional formats (YAML, Markdown with frontmatter, TOML, XML, etc.):

```go
// Register a custom YAML parser at init time
func init() {
    if ff, ok := database.GetDriver("flatfile").(*driver.FlatFile); ok {
        ff.parsers.Register(&yamlParser{})
    }
}

// yamlParser implements contractsorm.FileParser
type yamlParser struct{}

func (p *yamlParser) Format() string       { return "yaml" }
func (p *yamlParser) Extensions() []string { return []string{".yaml", ".yml"} }

func (p *yamlParser) ParseFile(reader io.Reader, config contractsorm.FlatFileConfig) ([]map[string]any, error) {
    // Parse YAML and return []map[string]any with Go-native types
    // ...
}
```

The driver auto-detects the parser from the file extension. If the extension is ambiguous, the format can be specified explicitly via `FlatFileConfig.FileExtension` or a future `FlatFileFormat` interface.

### Flat-File Parsing Flow

```
┌──────────────┐     ┌──────────────┐     ┌─────────────────┐     ┌──────────────┐
│ FlatFileSource│───▶│ Detect Parser│────▶│  Determine Mode │────▶│  Parse Data  │
│  .FilePath() │     │ (by extension│     │  (dir or file)  │     │  (FileParser)│
└──────────────┘     │  or config)  │     └─────────────────┘     └──────────────┘
                     └──────────────┘                                      │
                                                                           ▼
┌──────────────┐     ┌──────────────┐     ┌─────────────────┐     ┌──────────────┐
│  Mark as     │◀────│  Insert Rows │◀────│  Create SQLite  │◀────│  Resolve     │
│  Populated   │     │  (batched)   │     │  Table          │     │  Schema/Cols │
└──────────────┘     └──────────────┘     └─────────────────┘     └──────────────┘
                                                  ▲
                                                  │
                                          ┌──────────────┐
                                          │  Post-process│
                                          │  (flatten,   │
                                          │   validate,  │
                                          │   timestamps)│
                                          └──────────────┘
```

### Integration Points

#### 1. Driver Registration

```go
// database/orm/orm.go — createDriver()
case "flatfile":
    return driver.NewFlatFile()

// database/query/query_clone.go — newDriverForDialect()
case "flatfile":
    return driver.NewFlatFile()
```

#### 2. DSN Builder

```go
// database/db/config_builder.go — BuildDSN()
case "flatfile":
    return b.buildSQLiteDSN() // flatfile uses SQLite in-memory, same as array
```

#### 3. Config Validation

```go
// database/db/config_builder.go — ConnectionConfig.Validate()
case "flatfile":
    // database path is optional; empty defaults to :memory:
    return nil
```

#### 4. Placeholder Funcs

```go
// database/driver/placeholder.go — PlaceholderFuncs
"flatfile": sqlitePlaceholder,
```

#### 5. Query Builder — Model() Hook

```go
// database/query/query_model.go — Model()
if q.driver != nil && q.driver.Dialect() == "flatfile" {
    if source, ok := value.(contractsorm.FlatFileSource); ok {
        tableName := source.TableName()
        // ... same populate-once pattern as array driver
        if ffDriver, ok := q.driver.(contractsorm.FlatFilePopulator); ok {
            if err := ffDriver.Populate(q.ctx, q.db, source); err != nil {
                q.buildError = err
            }
        }
    }
}
```

#### 6. Connection Pool Configuration

```go
// database/orm/orm.go — buildQuery()
if connConfig.Driver == "sqlite" || connConfig.Driver == "array" || connConfig.Driver == "flatfile" {
    sqlDB.SetMaxOpenConns(1)
    sqlDB.SetMaxIdleConns(1)
}
```

#### 7. SQLite PRAGMA Optimizations

```go
// database/orm/orm.go — buildQuery()
if connConfig.Driver == "sqlite" || connConfig.Driver == "array" || connConfig.Driver == "flatfile" {
    _, _ = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
    _, _ = sqlDB.ExecContext(ctx, "PRAGMA synchronous=NORMAL;")
    _, _ = sqlDB.ExecContext(ctx, "PRAGMA foreign_keys=ON;")
    _, _ = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout=5000;")
}
```

#### 8. Database.Close() — Cleanup

The `Cleanupper` interface handles the flatfile driver automatically:

```go
// database/orm/orm.go — Close()
if drv, ok := r.drivers[r.connection]; ok {
    if c, ok := drv.(Cleanupper); ok {
        c.Cleanup(db)
    }
}
```

#### 9. detectDatabaseName

```go
// database/db.go — detectDatabaseName()
case "sqlite", "turso", "array", "flatfile":
    return "main"
```

#### 10. detectDriverName — NOT Needed

Same reasoning as the array driver: `detectDriverName()` reflects on `sqlDB.Driver().Type()`, which will always return `"sqlite"` since the flatfile driver embeds SQLite. When using `NewFromSQLDB` with a flatfile-backed `*sql.DB`, the caller must use `WithDriver("flatfile")` explicitly.

### Example Usage

#### CSV — Single File

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/dracory/neat"
    _ "modernc.org/sqlite"
)

// UserSource implements FlatFileSource to point to a CSV file.
type UserSource struct{}

func (s *UserSource) TableName() string { return "users" }
func (s *UserSource) FilePath() string  { return "data/users.csv" }

// Explicit schema for CSV type inference override
func (s *UserSource) Schema() map[string]string {
    return map[string]string{
        "id":      "int",
        "name":    "string",
        "email":   "string",
        "active":  "bool",
        "created": "time",
    }
}

// Declare primary key
func (s *UserSource) PrimaryKey() string { return "id" }

type User struct {
    ID      int
    Name    string
    Email   string
    Active  bool
    Created time.Time
}

func main() {
    config := neat.DBConfig{
        Default: "ff_db",
        Connections: map[string]neat.ConnectionConfig{
            "ff_db": {Driver: "flatfile"},
        },
    }

    database, err := neat.New(config)
    if err != nil {
        log.Fatal(err)
    }
    defer database.Close()

    var users []User
    err = database.Query().
        Model(&UserSource{}).
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

#### CSV — Headerless with Custom Column Names

```go
// SensorSource reads a headerless CSV with custom column names
type SensorSource struct{}

func (s *SensorSource) TableName() string { return "sensors" }
func (s *SensorSource) FilePath() string  { return "data/sensors.csv" }

func (s *SensorSource) Columns() []string {
    return []string{"timestamp", "temperature", "humidity", "pressure"}
}

func (s *SensorSource) Options() contractsorm.FlatFileConfig {
    return contractsorm.FlatFileConfig{
        HasHeader: false,
        Comma:     ';',
        NullIf:    "N/A",
    }
}

type Sensor struct {
    Timestamp   time.Time
    Temperature float64
    Humidity    float64
    Pressure    float64
}
```

#### CSV — Header Validation (Schema Drift Detection)

```go
// ProductSource validates that the CSV file has the expected headers
type ProductSource struct{}

func (s *ProductSource) TableName() string { return "products" }
func (s *ProductSource) FilePath() string  { return "data/products.csv" }

func (s *ProductSource) Options() contractsorm.FlatFileConfig {
    return contractsorm.FlatFileConfig{
        ExpectedHeaders: []string{"id", "name", "price", "category"},
        StrictHeaders:   true,
    }
}
```

#### JSON — Single File

```go
// OrderSource reads JSON with nested objects and flattens them
type OrderSource struct{}

func (s *OrderSource) TableName() string { return "orders" }
func (s *OrderSource) FilePath() string  { return "data/orders.json" }

func (s *OrderSource) Options() contractsorm.FlatFileConfig {
    return contractsorm.FlatFileConfig{
        Flatten:      true,
        FlattenDepth: 2,
    }
}

func (s *OrderSource) Schema() map[string]string {
    return map[string]string{
        "id":             "int",
        "customer_name":  "string",
        "customer_email": "string",
        "total":          "float",
        "address_city":   "string",
        "address_zip":    "string",
    }
}

type Order struct {
    ID            int
    CustomerName  string
    CustomerEmail string
    Total         float64
    AddressCity   string
    AddressZip    string
}
```

#### JSON — JSONL / NDJSON

```go
// EventSource reads a JSONL file (one event per line)
type EventSource struct{}

func (s *EventSource) TableName() string { return "events" }
func (s *EventSource) FilePath() string  { return "data/events.jsonl" }

func (s *EventSource) Options() contractsorm.FlatFileConfig {
    return contractsorm.FlatFileConfig{
        IsJSONL:       true,
        NullIfMissing: true,
    }
}

type Event struct {
    ID        int
    Type      string
    Timestamp time.Time
    Payload   string // stored as JSON string
}
```

#### JSON — Root Path Extraction

```go
// ConfigSource reads a nested JSON file with metadata wrapper
type ConfigSource struct{}

func (s *ConfigSource) TableName() string { return "config" }
func (s *ConfigSource) FilePath() string  { return "data/config.json" }

func (s *ConfigSource) Options() contractsorm.FlatFileConfig {
    return contractsorm.FlatFileConfig{
        RootPath: "settings.items",
    }
}
```

#### Directory Mode (One File Per Record)

```go
// PostSource reads a directory of JSON files, one post per file
type PostSource struct{}

func (s *PostSource) TableName() string { return "posts" }
func (s *PostSource) FilePath() string  { return "content/posts" }

// Enable directory mode
func (s *PostSource) IsDirectory() bool { return true }

// Use "slug" as primary key (filename becomes the slug value)
func (s *PostSource) PrimaryKey() string { return "slug" }

// Derive timestamps from file metadata
func (s *PostSource) Timestamps() bool { return true }

// Enable soft deletes (respect deleted_at field in JSON)
func (s *PostSource) SoftDeletes() bool { return true }

type Post struct {
    Slug        string
    Title       string
    Content     string
    Published   bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

Directory structure:
```
content/posts/
  ├── hello-world.json
  ├── my-second-post.json
  └── draft-post.json
```

Each file contains one JSON object:
```json
{
    "title": "Hello World",
    "content": "My first post.",
    "published": true
}
```

The driver reads all `.json` files, adds `slug` from the filename, and adds `created_at`/`updated_at` from file metadata.

### Sample Data Files

#### users.csv (single-file mode)

```csv
id,name,email,active,created
1,Alice,alice@example.com,true,2024-01-15T10:30:00Z
2,Bob,bob@example.com,false,2024-02-20T14:45:00Z
3,Charlie,charlie@example.com,true,2024-03-10T09:00:00Z
```

#### events.jsonl (JSONL format)

```jsonl
{"id": 1, "type": "login", "timestamp": "2024-01-15T10:30:00Z", "user_id": 42}
{"id": 2, "type": "logout", "timestamp": "2024-01-15T11:00:00Z", "user_id": 42}
{"id": 3, "type": "purchase", "timestamp": "2024-01-15T12:15:00Z", "user_id": 17, "amount": 49.99}
```

## Implementation Plan

### Phase 1: Contracts and Core Driver
1. Create `contracts/database/orm/flatfile_source.go` with all interfaces: `FlatFileSource`, `FlatFileDir`, `FlatFileSchema`, `FlatFileColumns`, `FlatFilePrimaryKey`, `FlatFileTimestamps`, `FlatFileSoftDeletes`, `FlatFileOptions`, `FlatFileConfig`, `FlatFilePopulator`, `FileParser`, `FileParserRegistry`
2. Extract `isSimpleIdentifier` as a shared package-level function in `database/driver/` (used by both `Array` and `FlatFile`)
3. Implement `database/driver/flatfile.go` — driver struct, `NewFlatFile()`, `Dialect()`, `Populate()`, `Cleanup()`, parser registry, `detectParser()`, `readDirectory()`
4. Implement `FileParserRegistry` (register, get by format, get by extension)

### Phase 2: Built-in Parsers
1. Implement `database/driver/flatfile_csv.go` — `csvParser`:
   - CSV parsing with `encoding/csv` (Comma, Comment, LazyQuotes)
   - Header detection / headerless mode with `FlatFileColumns` / generated `col_N`
   - Header validation (`StrictHeaders`, `ExpectedHeaders`)
   - String-to-type inference (int, float, bool, time, string)
   - `SkipRows`, `NullIf`, `TrimSpace`
2. Implement `database/driver/flatfile_json.go` — `jsonParser`:
   - JSON array parsing with `encoding/json`
   - JSONL/NDJSON mode with `bufio.Scanner`
   - Root path extraction (`navigatePath`)
   - Nested object flattening (`flattenObject` with depth control)
   - `NullIfMissing`, `TrimStrings`
   - Native type preservation (no string parsing needed)

### Phase 3: Shared Features
1. Implement directory mode (`readDirectory`) — read all matching files, filename as primary key
2. Implement file-metadata timestamps (`FlatFileTimestamps`) — `os.Stat` for mtime/ctime
3. Implement soft deletes (`FlatFileSoftDeletes`) — check `deleted_at`, add default `WHERE deleted_at IS NULL`
4. Implement primary key support (`FlatFilePrimaryKey`) — add `PRIMARY KEY` constraint in `createTable`
5. Implement schema inference (`inferSchema`) from Go-native values returned by parsers
6. Implement batched inserts using `batchSize := 500 / len(sortedCols)` (same formula as array driver)

### Phase 4: Integration
1. Register `flatfile` driver in `createDriver()` and `newDriverForDialect()`
2. Add `flatfile` case to `BuildDSN()` and `ConnectionConfig.Validate()`
3. Add `flatfile` to `PlaceholderFuncs` map
4. Add `flatfile` to `detectDatabaseName()` (add to existing `"sqlite", "turso", "array"` case)
5. Add `flatfile` to connection pool configuration in `orm.go` (`SetMaxOpenConns(1)`, `SetMaxIdleConns(1)`)
6. Add `flatfile` to SQLite PRAGMA optimizations in `orm.go`
7. Add `Model()` hook in `query_model.go` for `FlatFileSource` detection
8. Refactor `Orm.Close()` to use `Cleanupper` interface instead of per-driver type assertions

### Phase 5: Tests
1. Driver unit tests in `database/driver/flatfile_test.go`:
   - Parser registry (register, get by format, get by extension)
   - Parser auto-detection from file extension
   - Directory mode (read all files, filename as primary key)
   - Timestamps (file mtime/ctime)
   - Soft deletes
   - Primary key constraint
   - Invalid identifiers (SQL injection prevention)
   - Concurrent population
   - Cleanup method
   - MaxFlatFileRows limit enforcement
   - File not found / directory not found errors
2. CSV parser tests in `database/driver/flatfile_csv_test.go`:
   - Basic CSV parsing
   - Type inference (int, float, bool, time, text)
   - Schema inference with mixed types (widening)
   - Schema inference skips empty strings and NullIf values
   - Explicit schema via `FlatFileSchema`
   - Empty CSV (table created, zero rows)
   - Custom delimiter (tab, semicolon)
   - HasHeader = false with generated column names
   - HasHeader = false with custom column names via `FlatFileColumns`
   - HasHeader = false with invalid identifier in `FlatFileColumns` (error)
   - HasHeader = true with `FlatFileColumns` (ignored)
   - Header validation: strict mode passes
   - Header validation: strict mode error on mismatch
   - Header validation: non-strict mode ignores mismatch
   - SkipRows
   - NullIf (empty string → NULL)
   - LazyQuotes
3. JSON parser tests in `database/driver/flatfile_json_test.go`:
   - Basic JSON array parsing
   - Type inference from native JSON types (int, float, bool, string, null)
   - Time detection from RFC3339 string values
   - Schema inference with mixed types (widening)
   - Explicit schema via `FlatFileSchema`
   - Empty JSON array (table created, zero rows)
   - JSONL parsing (one object per line)
   - JSONL with trailing newline / empty lines (skipped)
   - Root path extraction (nested structure)
   - Root path not found (error)
   - Root path pointing to non-array (error)
   - Nested object flattening with depth limit
   - Nested object flattening with depth 0 (unlimited)
   - Nested arrays stored as JSON strings (not flattened)
   - Missing keys with `NullIfMissing = false` (zero values)
   - Missing keys with `NullIfMissing = true` (SQL NULL)
   - Custom column names via `FlatFileColumns`
   - Invalid JSON (parse error)
4. Directory mode tests:
   - Basic populate from directory of JSON files
   - Filename as primary key value
   - Custom file extension via `FileExtension`
   - Empty directory (table created, zero rows)
   - Non-matching file in directory (skipped)
   - Nested subdirectories (skipped)
   - Timestamps: each record gets individual timestamps
   - Soft deletes: records with non-null deleted_at excluded
   - Soft deletes: explicit query for soft-deleted records
5. Integration test in `database/flatfile_integration_test.go`
6. Example in `examples/flatfile-driver/` with sample CSV, JSON, JSONL, and directory-mode files

### Phase 6: Documentation
1. Create `examples/flatfile-driver/README.md`
2. Add flatfile driver to main docs (driver-registration page, API reference)
3. Document type inference rules, format-specific options, and custom parser registration
4. Mark `csv-driver.md` and `json-driver.md` as superseded by this proposal

## Design Decisions

### Why a Single Driver Instead of Separate CSV/JSON Drivers?

The CSV and JSON driver proposals shared ~80% of their design:
- Same SQLite embedding architecture
- Same `Populate`/`Cleanup` pattern with `sync.Map`
- Same directory mode, timestamps, soft deletes, primary key features
- Same 10 integration points (driver registration, DSN, config, placeholders, etc.)
- Same shared code (`isSimpleIdentifier`, `isPopulated`, `createTable`, `insertRows`)

A unified driver eliminates this duplication while keeping format-specific logic cleanly separated via the `FileParser` interface. Adding a new format (YAML, Markdown, TOML) only requires implementing one interface — no new driver, no new integration points, no new config validation.

### Why Pluggable Parsers Instead of Hardcoded Format Switch?

A `FileParser` interface with a registry allows:
- **Extensibility**: Users can register custom parsers for any format without modifying the driver
- **Testability**: Parsers can be unit-tested in isolation
- **Separation of concerns**: The driver handles shared logic (SQLite, schema, batching, directory mode); parsers handle format-specific parsing
- **Auto-detection**: Parser selection from file extension is automatic but overridable

### Why a Single `FlatFileConfig` Instead of Per-Format Configs?

A single config struct with format-specific fields is simpler than a hierarchy of config types. Fields that don't apply to the selected format are simply ignored. This avoids:
- Type assertion chains to access format-specific config
- Multiple config interfaces
- Complexity for users who only need one format

If the config grows too large in the future, it can be refactored into a map of format-specific options.

### Why Embed SQLite (Same as Array Driver)?

Same reasoning as the array driver — all query builder features work out of the box (WHERE, JOIN, ORDER BY, aggregates, etc.), no need to implement a custom query engine. The driver only handles "how data gets into the table", not "how queries are executed".

### Why Not Just Use the Array Driver?

Users could parse files themselves and return `[]map[string]any` from an `ArraySource`. However, this requires significant boilerplate:
- Manual file opening and reading
- Manual format-specific parsing (CSV delimiters, JSON nesting, JSONL streaming)
- No standard way to handle directory mode, timestamps, or soft deletes
- No type inference from file content

The flat-file driver eliminates this boilerplate and provides a standard, tested implementation.

### Security: File Path Validation

Same as the array driver — the `FilePath()` method returns a path that the driver will open. The driver does NOT restrict file paths by default (the caller controls which files to open). Applications can wrap `FlatFileSource` with path validation logic if needed.

### Shared Code with Array Driver

The following logic is shared across the array and flat-file drivers:
- `isSimpleIdentifier` — identifier validation (shared package-level function)
- `isPopulated` / `markPopulated` / `getTableMutex` — sync.Map management
- `Cleanup` — stale entry removal
- `createTable` — SQLite table creation from schema
- `insertRows` — batched SQLite inserts
- Type widening logic in schema inference

With two drivers sharing the same pattern, extracting a shared `populator` base struct is justified. This can be done as part of this proposal or deferred to a follow-up refactoring.

## Benefits

1. **Zero Boilerplate**: Query CSV, JSON, JSONL files directly without manual parsing
2. **Full Query Builder**: All ORM features work (WHERE, JOIN, ORDER BY, aggregates, etc.)
3. **Multi-Format**: One driver handles CSV, JSON, JSONL, and any custom format via pluggable parsers
4. **Extensible**: Register custom parsers for YAML, Markdown, TOML, XML, etc. without modifying the driver
5. **Type Safety**: Automatic type inference — JSON has native types; CSV uses string-to-type inference; both overridable via `FlatFileSchema`
6. **Directory Mode**: One file per record with filename-as-primary-key — Git-friendly, content-management-friendly (like Orbit and Laravel Paper)
7. **File-Metadata Timestamps**: Derive `created_at`/`updated_at` from file mtime/ctime (like Laravel Paper's `#[Timestamps]`)
8. **Soft Deletes**: Respect `deleted_at` field for soft-deleted records (like Orbit's `SoftDeletes` trait)
9. **CSV-Specific**: Custom delimiters, headerless mode, header validation, `LazyQuotes`, `SkipRows`
10. **JSON-Specific**: JSONL/NDJSON, root path extraction, nested object flattening, missing key handling
11. **Primary Key Support**: Declare a primary key for faster lookups and ORM association compatibility
12. **Consistency**: Same architecture and patterns as the array driver
13. **Test Fixtures**: Easy to load test data from flat files (common in Go testing)
14. **Single Integration**: One driver registration, one set of integration points — adding formats doesn't require touching ORM internals

## Risks and Mitigations

### Risk 1: Memory Usage for Large Files
- **Issue**: Loading an entire file into in-memory SQLite could consume significant memory
- **Mitigation**: `MaxFlatFileRows` limit (100,000 rows); JSONL mode with `bufio.Scanner` for streaming; future streaming mode enhancement

### Risk 2: Type Inference Ambiguity (CSV)
- **Issue**: CSV values like "123" could be int or string (e.g., zip codes with leading zeros)
- **Mitigation**: `FlatFileSchema` interface allows explicit type declaration; inference is opt-in default

### Risk 3: File Encoding (CSV)
- **Issue**: CSV files may use non-UTF-8 encodings (Latin-1, Windows-1252)
- **Mitigation**: Phase 1 supports UTF-8 only. Encoding support can be added as a future `FlatFileConfig.Encoding` field when needed.

### Risk 4: Nested Object Complexity (JSON)
- **Issue**: Deeply nested JSON can produce many columns when flattened, or complex JSON strings when stored as-is
- **Mitigation**: `FlattenDepth` limits nesting depth; default strategy (store as JSON string) is safe and queryable via SQLite JSON functions

### Risk 5: Config Bloat
- **Issue**: `FlatFileConfig` contains fields for all formats, which could grow large
- **Mitigation**: Acceptable for now — format-specific fields are clearly documented. If it grows too large, refactor into a map of format-specific options.

### Risk 6: Custom Parser Quality
- **Issue**: User-registered custom parsers may produce inconsistent data types or invalid schemas
- **Mitigation**: The driver validates all column names via `isSimpleIdentifier` and infers schema from Go-native types returned by the parser. Parsers that return invalid data will produce clear errors.

## Future Enhancements

1. **Streaming Mode**: For very large files, use SQLite's virtual table mechanism to stream rows on demand instead of loading all into memory
2. **Write Support**: Allow ORM operations (Create, Update, Delete) to write back to files. Approaches:
   - **Direct file writing**: Bypass SQLite for writes, operating on files directly (like Orbit and Laravel Paper)
   - **SQLite-to-file sync**: After writes, export the SQLite table back to the original format
   - **Directory-mode write**: In directory mode, create/update/delete individual files per record
3. **Append-Only Mode**: Restrict write operations to inserts only (no updates or deletes), useful for log files and audit trails (like FlatModel's `AppendOnly` trait)
4. **Backup Before Write**: Create timestamped backup copies of files before applying writes (like FlatModel's `Backupable` trait)
5. **YAML Parser**: Built-in parser for YAML files (common for configuration, like Orbit's YAML driver)
6. **Markdown Parser**: Built-in parser for Markdown files with YAML frontmatter + Markdown body (like Orbit's Markdown driver and Laravel Paper's markdown support)
7. **TOML Parser**: Built-in parser for TOML files (common in Go configuration)
8. **JSON Query Functions**: Leverage SQLite's JSON1 extension for querying nested JSON string columns (`json_extract`, `json_array_length`, etc.)
9. **Relationships Between File-Backed Models**: Support `belongsTo`/`hasMany` between file-backed models (like Laravel Paper's `belongsToPaper`/`hasManyPaper`)
10. **Remote Sources**: Support HTTP/HTTPS URLs as file paths
11. **Compressed Files**: Support `.csv.gz`, `.json.gz`, `.jsonl.gz` files with automatic decompression
12. **File Encoding**: Support non-UTF-8 encodings (Latin-1, Windows-1252) via a `FlatFileConfig.Encoding` field
13. **Schema Validation**: Validate records against a schema definition (JSON Schema, CSV schema) before populating
14. **Shared Populator Base**: Extract common logic between array and flat-file drivers into a shared struct

## References

- Array driver implementation: `database/driver/array.go`
- Array source contracts: `contracts/database/orm/array_source.go`
- Go `encoding/csv` package: https://pkg.go.dev/encoding/csv
- Go `encoding/json` package: https://pkg.go.dev/encoding/json
- JSON Lines specification: https://jsonlines.org/
- SQLite in-memory databases: https://www.sqlite.org/inmemorydb.html
- SQLite JSON1 extension: https://www.sqlite.org/json1.html
- Sushi (Laravel array driver): https://github.com/calebporzio/sushi
- FlatModel (Laravel CSV driver): https://packagist.org/packages/flatmodel/laravel-csv-flatmodel
- Orbit (Laravel flat-file Eloquent): https://packagist.org/packages/ryangjchandler/orbit
- Laravel Paper (flat-file Eloquent): https://packagist.org/packages/jacobjoergensen/laravel-paper
