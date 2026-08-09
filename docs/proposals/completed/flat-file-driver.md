# Flat-File Source Adapters

**Date**: August 9, 2026 (revised from June 24, 2026)
**Status**: Superseded — implemented as per-file source packages (see notes below)
**Priority**: Medium
**Supersedes**: `csv-driver.md`, `json-driver.md`
**Revisions**:
- **v2 (2026-08-09)**: Replaced the standalone `flatfile` driver design with array-source adapters. The array source helper (`NewArraySourceFrom`, completed 2026-08-08) and the existing array driver now provide all SQLite orchestration. This revision deletes ~250 lines of duplicated driver logic, 10 integration points, and a new driver registration. The only array-driver change required is adding an optional `ArrayPrimaryKey` interface.
- **v3 (2026-08-09)**: Superseded. The proposed `support/flatfile/` package with `FileParser` interface, parser registry, directory mode, timestamps, soft deletes, and PK injection was replaced by three thin packages using only the Go standard library: `support/csvsource/`, `support/jsonsource/`, `support/xmlsource/`. Each provides `New*Source(content, tableName)` (string input) and `New*FileSource(filePath)` (file input) constructors that return `*arraysource.Model`. No array driver changes or new contracts were needed. Directory-as-database mode is covered by a separate proposal: `csv-directory-driver.md`.

## Problem Statement

The array driver allows querying in-memory data via SQLite, but it requires the user to supply `Rows()` as `[]map[string]any`. The `NewArraySourceFrom` helper (completed 2026-08-08) eliminated boilerplate for **Go structs and map literals**, but there is still no built-in way to query flat-file data formats — CSV, JSON, JSONL/NDJSON, YAML, or Markdown with frontmatter — without first parsing and converting them into Go data structures.

Users who have flat-file data (exports, config, test fixtures, datasets, content files) must write boilerplate parsing code before they can use the ORM's query builder.

### Current Limitations

- No native flat-file support — users must parse files themselves and feed rows into the array driver
- No automatic type inference from file content (CSV needs string parsing; JSON has native types but no ORM integration)
- No support for format-specific features (CSV delimiters/headers, JSON nested objects/JSONL, YAML frontmatter)
- No directory-based mode (one file per record, like Orbit and Laravel Paper)
- No file-metadata timestamps (e.g., file mtime as `updated_at`)
- No soft delete support for file-backed records
- No primary key declaration for file-backed tables (the array driver itself has no PK support)
- No custom column names for headerless files
- No header/schema validation to detect drift
- No streaming for large files

### Example of the Desired Experience

```go
config := neat.DBConfig{
    Default: "ff_db",
    Connections: map[string]neat.ConnectionConfig{
        "ff_db": {
            Driver: "array", // flat-file sources run on the array driver
        },
    },
}

database, _ := neat.New(config)
defer database.Close()

// Query a CSV file directly
var users []User
err := database.Query().
    Model(flatfile.NewFromCSV("data/users.csv", "users")).
    Where("country = ?", "US").
    OrderBy("name", "asc").
    Get(&users)

// Query a JSON file directly
var products []Product
err = database.Query().
    Model(flatfile.NewFromJSON("data/products.json", "products")).
    Where("price > ?", 50).
    Get(&products)

// Query a directory of JSON files (one file per record)
var posts []Post
err = database.Query().
    Model(flatfile.NewFromDirectory("content/posts", "posts",
        flatfile.WithPrimaryKey("slug"),
        flatfile.WithTimestamps(),
        flatfile.WithSoftDeletes(),
    )).
    Where("published = ?", true).
    OrderBy("updated_at", "desc").
    Get(&posts)
```

## Proposed Solution

Implement flat-file support as **`ArraySource` adapters** in a new `support/flatfile` package, plus a small extension to the array driver for primary key support.

The existing array driver already handles every shared concern: SQLite embedding, schema inference, type widening, batched inserts, concurrency via `sync.Map`, cleanup, and all 10 ORM integration points (driver registration, DSN, config validation, placeholders, the `Model()` hook, connection pool config, PRAGMAs, `Close()` cleanup, `detectDatabaseName`). Flat-file sources reuse all of it by implementing `ArraySource`.

What this proposal adds:
1. **`ArrayPrimaryKey`** — one optional interface on the array driver's contracts, plus a small `createTable` change to honor it. Benefits both array and flat-file sources.
2. **`flatfile.Source`** — an `ArraySource` implementation that reads files via pluggable parsers. Handles directory mode, file-metadata timestamps, soft deletes, and primary key value injection inside `Rows()`.
3. **`FileParser` interface + registry** — format-specific parsing (CSV, JSON, JSONL) with auto-detection from file extension. Extensible to YAML, Markdown, TOML without touching the array driver.
4. **Constructor helpers** — `NewFromCSV`, `NewFromJSON`, `NewFromDirectory` for ergonomic one-liner usage.

What this proposal does **not** add (inherited from the array driver):
- A new driver registration
- A new dialect
- DSN builder cases
- Config validation cases
- Placeholder func entries
- Connection pool configuration
- SQLite PRAGMA optimizations
- `detectDatabaseName` cases
- `Model()` hook changes
- `Close()` cleanup changes
- Schema inference, type widening, batched inserts, concurrency, cleanup

### Architecture

```
contracts/
  database/
    orm/
      array_source.go          // ADD: ArrayPrimaryKey interface (one method)

database/
  driver/
    array.go                   // MODIFY: createTable honors ArrayPrimaryKey
    array_test.go              // ADD: primary key constraint tests

support/
  flatfile/
    source.go                  // Source struct (implements ArraySource + ArraySchema + ArrayPrimaryKey)
    source_test.go
    options.go                 // Option functions (WithTimestamps, WithSoftDeletes, etc.)
    options_test.go
    csv_parser.go              // CSV parser (implements FileParser)
    csv_parser_test.go
    json_parser.go             // JSON / JSONL parser (implements FileParser)
    json_parser_test.go
    parser_registry.go         // FileParserRegistry
    parser_registry_test.go
    constructors.go            // NewFromCSV, NewFromJSON, NewFromDirectory
    constructors_test.go
    directory.go               // directory-mode reading helper
    directory_test.go
    timestamps.go              // file mtime/ctime helper
    timestamps_test.go
    flatten.go                 // JSON nested-object flattening helper
    flatten_test.go

examples/
  flatfile-sources/
    main.go                    // Example usage (CSV, JSON, directory mode)
    main_test.go
    README.md
    data/
      users.csv                // Sample CSV (single-file mode)
      products.json            // Sample JSON (single-file mode)
      events.jsonl             // Sample JSONL (single-file mode)
      posts/                   // Sample directory mode (one file per record)
        hello-world.json
        my-second-post.json
```

No files are added under `database/driver/` for flat-file logic. The `database/driver/array.go` file receives one small modification (primary key support). All flat-file code lives under `support/flatfile/`, mirroring the existing `support/arraysource/` pattern.

## Array Driver Extension: Primary Key

The array driver's `createTable` currently emits columns without any `PRIMARY KEY` constraint. This proposal adds one optional interface to the array contracts and a small change to `createTable` to honor it.

### Contract Addition

```go
// contracts/database/orm/array_source.go (addition)

// ArrayPrimaryKey is an optional interface for sources that want a PRIMARY KEY
// constraint on a column. The driver appends "PRIMARY KEY" to the declared
// column's definition when creating the SQLite table.
//
// This benefits both array sources (e.g., NewArraySourceFrom with a struct
// whose ID field should be the PK) and flat-file sources (e.g., directory
// mode where the filename becomes the PK value).
type ArrayPrimaryKey interface {
    PrimaryKey() string // column name to mark as PRIMARY KEY
}
```

### Driver Modification

```go
// database/driver/array.go — createTable (modified)

func (a *Array) createTable(ctx context.Context, db *sql.DB, tableName string, schema map[string]string, sortedCols []string, pkCol string) error {
    var columns []string

    for _, col := range sortedCols {
        sqlType := schema[col]
        // ... existing type conversion switch unchanged ...
        colDef := fmt.Sprintf("\"%s\" %s", col, sqlType)
        if col == pkCol && pkCol != "" {
            colDef += " PRIMARY KEY"
        }
        columns = append(columns, colDef)
    }

    sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS \"%s\" (%s)", tableName, strings.Join(columns, ", "))
    _, err := db.ExecContext(ctx, sql)
    return err
}
```

The `Populate` method passes the primary key column name (if the source implements `ArrayPrimaryKey`) to `createTable`:

```go
// database/driver/array.go — Populate (addition near line 76)

var pkCol string
if pks, ok := source.(contractsorm.ArrayPrimaryKey); ok {
    pkCol = pks.PrimaryKey()
    if pkCol != "" && !a.isSimpleIdentifier(pkCol) {
        return fmt.Errorf("invalid primary key column name: %s", pkCol)
    }
    if pkCol != "" {
        if _, ok := schema[pkCol]; !ok {
            return fmt.Errorf("primary key column %q not found in schema", pkCol)
        }
    }
}

// ... later:
if err := a.createTable(ctx, db, tableName, schema, sortedCols, pkCol); err != nil {
    return fmt.Errorf("failed to create table %s: %w", tableName, err)
}
```

This is a backward-compatible change: existing array sources don't implement `ArrayPrimaryKey`, so `pkCol` is empty and `createTable` behaves exactly as before.

## Flat-File Source Adapter

### Source Struct

```go
// support/flatfile/source.go

package flatfile

import (
    "os"
    "path/filepath"
    "strings"

    contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// Source implements contractsorm.ArraySource, ArraySchema, and ArrayPrimaryKey.
// It reads flat file(s) via a FileParser and returns []map[string]any from Rows().
//
// Use the constructor helpers (NewFromCSV, NewFromJSON, NewFromDirectory) instead
// of constructing Source directly.
type Source struct {
    table      string
    filePath   string
    isDir      bool
    parser     FileParser
    config     Config
    schema     map[string]string // optional explicit schema (nil = infer)
    pkCol      string            // primary key column ("" = none)
    timestamps bool              // add created_at/updated_at from file metadata
    softDeletes bool             // respect deleted_at field
    columns    []string          // custom column names (headerless CSV, etc.)
}

// TableName returns the SQLite table name.
func (s *Source) TableName() string { return s.table }

// Rows reads and parses the file(s), applies directory mode, timestamps,
// soft deletes, and primary key injection, then returns the rows.
// The array driver handles schema inference, table creation, and insertion.
func (s *Source) Rows() ([]map[string]any, error) {
    if s.isDir {
        return s.readDirectory()
    }
    return s.readSingleFile()
}

// Schema returns an explicit schema if set, satisfying ArraySchema.
// Returns nil if not set, letting the array driver infer from rows.
func (s *Source) Schema() map[string]string { return s.schema }

// PrimaryKey returns the primary key column name, satisfying ArrayPrimaryKey.
// Returns "" if no primary key is declared.
func (s *Source) PrimaryKey() string { return s.pkCol }
```

### Reading a Single File

```go
// support/flatfile/source.go (continued)

func (s *Source) readSingleFile() ([]map[string]any, error) {
    f, err := os.Open(s.filePath)
    if err != nil {
        return nil, fmt.Errorf("flatfile: cannot open %s: %w", s.filePath, err)
    }
    defer f.Close()

    rows, err := s.parser.ParseFile(f, s.config)
    if err != nil {
        return nil, fmt.Errorf("flatfile: parse error in %s: %w", s.filePath, err)
    }

    if s.timestamps {
        stat, err := os.Stat(s.filePath)
        if err != nil {
            return nil, fmt.Errorf("flatfile: cannot stat %s: %w", s.filePath, err)
        }
        applyTimestamps(rows, stat)
    }

    if s.softDeletes {
        rows = filterSoftDeleted(rows)
    }

    return rows, nil
}
```

### Reading a Directory (One File Per Record)

```go
// support/flatfile/directory.go

// readDirectory reads all files matching the configured extension in the
// directory. Each file is one record. The filename (without extension) is
// inserted as the primary key column value.
//
// This mirrors Orbit's file-per-record approach and Laravel Paper's
// slug-as-filename pattern:
//   content/posts/hello-world.json → slug = "hello-world"
func (s *Source) readDirectory() ([]map[string]any, error) {
    entries, err := os.ReadDir(s.filePath)
    if err != nil {
        return nil, fmt.Errorf("flatfile: cannot read directory %s: %w", s.filePath, err)
    }

    ext := s.config.FileExtension
    if ext == "" {
        // auto-detect from first matching file, or default to parser's extensions
        ext = s.parser.Extensions()[0]
    }

    var rows []map[string]any
    for _, entry := range entries {
        if entry.IsDir() {
            continue // skip subdirectories
        }
        if !strings.HasSuffix(entry.Name(), ext) {
            continue
        }

        fullPath := filepath.Join(s.filePath, entry.Name())
        f, err := os.Open(fullPath)
        if err != nil {
            return nil, fmt.Errorf("flatfile: cannot open %s: %w", fullPath, err)
        }

        records, err := s.parser.ParseFile(f, s.config)
        f.Close()
        if err != nil {
            return nil, fmt.Errorf("flatfile: parse error in %s: %w", fullPath, err)
        }

        // Each file is one record. If the parser returns multiple rows,
        // use only the first (or error — see Design Decisions).
        if len(records) == 0 {
            continue
        }
        row := records[0]

        // Inject filename as primary key value
        if s.pkCol != "" {
            base := strings.TrimSuffix(entry.Name(), ext)
            row[s.pkCol] = base
        }

        if s.timestamps {
            stat, err := os.Stat(fullPath)
            if err != nil {
                return nil, fmt.Errorf("flatfile: cannot stat %s: %w", fullPath, err)
            }
            applyTimestamps([]map[string]any{row}, stat)
        }

        if s.softDeletes && isSoftDeleted(row) {
            continue
        }

        rows = append(rows, row)
    }

    return rows, nil
}
```

### Timestamps and Soft Deletes Helpers

```go
// support/flatfile/timestamps.go

// applyTimestamps adds created_at and updated_at to each row from file metadata.
// updated_at = mtime, created_at = ctime (falls back to mtime on platforms
// where ctime is unavailable).
func applyTimestamps(rows []map[string]any, stat os.FileInfo) {
    mtime := stat.ModTime()
    // ctime is platform-dependent; on most Unix systems stat.Sys() exposes it.
    // For portability, fall back to mtime if ctime cannot be determined.
    ctime := fileCreationTime(stat) // returns mtime if ctime unavailable
    for _, row := range rows {
        row["created_at"] = ctime
        row["updated_at"] = mtime
    }
}

// filterSoftDeleted removes rows where deleted_at is non-null and non-empty.
// Soft-deleted rows are excluded from the SQLite table entirely.
//
// Note: this is a simpler model than the ORM's soft-delete infrastructure
// (which loads rows and filters via WHERE). For flat files, excluding at
// parse time is cleaner — the file is the source of truth. Users who want
// to query soft-deleted records can construct a second Source without
// WithSoftDeletes().
func filterSoftDeleted(rows []map[string]any) []map[string]any {
    out := rows[:0]
    for _, row := range rows {
        if !isSoftDeleted(row) {
            out = append(out, row)
        }
    }
    return out
}

func isSoftDeleted(row map[string]any) bool {
    v, ok := row["deleted_at"]
    if !ok || v == nil {
        return false
    }
    if s, ok := v.(string); ok && s == "" {
        return false
    }
    return true
}
```

## FileParser Interface and Registry

```go
// support/flatfile/source.go (interface declarations)

// FileParser is implemented by each format parser (CSV, JSON, YAML, etc.).
// A parser reads a file and returns rows as []map[string]any with Go-native
// types (int, float64, bool, string, time.Time, nil, nested maps/slices).
//
// In single-file mode, ParseFile is called once for the entire file.
// In directory mode, ParseFile is called once per file (each file = one record).
type FileParser interface {
    // Format returns the parser's format name (e.g., "csv", "json", "yaml").
    Format() string

    // Extensions returns the file extensions this parser handles (e.g., [".csv", ".tsv"]).
    Extensions() []string

    // ParseFile reads from reader and returns rows.
    // config provides format-specific options.
    ParseFile(reader io.Reader, config Config) ([]map[string]any, error)
}

// FileParserRegistry manages available format parsers.
type FileParserRegistry struct {
    parsers      map[string]FileParser   // by format name
    byExtension  map[string]FileParser   // by extension
}

func NewParserRegistry() *FileParserRegistry { ... }
func (r *FileParserRegistry) Register(parser FileParser) { ... }
func (r *FileParserRegistry) Get(format string) (FileParser, bool) { ... }
func (r *FileParserRegistry) GetByExtension(ext string) (FileParser, bool) { ... }
func (r *FileParserRegistry) Formats() []string { ... }
```

### Config

```go
// support/flatfile/options.go

// Config holds shared and format-specific parsing configuration.
// Format-specific fields are used only by the relevant parser.
type Config struct {
    // Shared
    FileExtension string // File extension for directory mode (default: auto-detect)

    // CSV-specific
    Comma         rune
    Comment       rune
    HasHeader     bool   // default: true
    SkipRows      int
    NullIf        string
    TrimSpace     bool
    LazyQuotes    bool
    ExpectedHeaders []string
    StrictHeaders   bool

    // JSON-specific
    RootPath      string // dot-separated path to row array (e.g., "data.users")
    IsJSONL       bool
    Flatten       bool
    FlattenDepth  int    // 0 = unlimited; default: 3
    NullIfMissing bool
    TrimStrings   bool
}

// Option is a functional option for configuring a Source.
type Option func(*Source)

func WithSchema(schema map[string]string) Option { ... }
func WithPrimaryKey(col string) Option { ... }
func WithTimestamps() Option { ... }
func WithSoftDeletes() Option { ... }
func WithColumns(cols []string) Option { ... }
func WithConfig(cfg Config) Option { ... }
func WithParser(p FileParser) Option { ... }
```

### Constructor Helpers

```go
// support/flatfile/constructors.go

// NewFromCSV creates a Source that reads a CSV file.
// Table name is used as the SQLite table name.
func NewFromCSV(filePath, table string, opts ...Option) *Source {
    s := &Source{
        table:    table,
        filePath: filePath,
        parser:   &csvParser{},
        config:   Config{HasHeader: true, Comma: ','},
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}

// NewFromJSON creates a Source that reads a JSON or JSONL file.
// Format (JSON vs JSONL) is auto-detected from the file extension, or set
// via WithConfig(Config{IsJSONL: true}).
func NewFromJSON(filePath, table string, opts ...Option) *Source {
    s := &Source{
        table:    table,
        filePath: filePath,
        parser:   &jsonParser{},
        config:   Config{FlattenDepth: 3},
    }
    if strings.HasSuffix(filePath, ".jsonl") || strings.HasSuffix(filePath, ".ndjson") {
        s.config.IsJSONL = true
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}

// NewFromDirectory creates a Source in directory mode.
// Each file in the directory is one record. The parser is auto-detected
// from the file extension, or set via WithParser.
func NewFromDirectory(dirPath, table string, opts ...Option) *Source {
    s := &Source{
        table:    table,
        filePath: dirPath,
        isDir:    true,
        parser:   nil, // auto-detect per file
    }
    for _, opt := range opts {
        opt(s)
    }
    if s.parser == nil {
        s.parser = defaultRegistry.GetByExtension(s.config.FileExtension)
    }
    return s
}
```

## Built-in Parsers

### CSV Parser

```go
// support/flatfile/csv_parser.go

package flatfile

import (
    "encoding/csv"
    "io"
    "strconv"
    "strings"
    "time"
)

// csvParser implements FileParser for CSV files.
type csvParser struct{}

func (p *csvParser) Format() string       { return "csv" }
func (p *csvParser) Extensions() []string { return []string{".csv", ".tsv"} }

func (p *csvParser) ParseFile(reader io.Reader, config Config) ([]map[string]any, error) {
    // 1. Create csv.Reader with config (Comma, Comment, LazyQuotes)
    // 2. If HasHeader: read first row as column names
    //    If !HasHeader and config provides custom columns: use those
    //    If !HasHeader and no custom columns: generate col_0, col_1, ...
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

### JSON Parser

```go
// support/flatfile/json_parser.go

package flatfile

import (
    "bufio"
    "encoding/json"
    "io"
    "strings"
)

// jsonParser implements FileParser for JSON and JSONL files.
type jsonParser struct{}

func (p *jsonParser) Format() string       { return "json" }
func (p *jsonParser) Extensions() []string { return []string{".json", ".jsonl", ".ndjson"} }

func (p *jsonParser) ParseFile(reader io.Reader, config Config) ([]map[string]any, error) {
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

## Type Inference

Type inference happens in two stages, split across the parser and the array driver:

**Stage 1 — Parser converts to Go-native types** (in `support/flatfile/`):
- CSV: String values are parsed (try int → float → bool → time → string)
- JSON: Native types are preserved directly (no parsing needed)

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

**Stage 2 — Array driver infers SQLite column types from Go-native values** (in `database/driver/array.go`, already implemented):
- `inferSchema` scans all rows and applies type widening
- `INTEGER` + `REAL` → `REAL`
- Any incompatible mix → `TEXT`
- Overridable via `ArraySchema` (which `flatfile.Source` implements)

Stage 2 is inherited from the array driver with no changes. This is a key benefit of the adapter approach: the flat-file package never touches SQLite type mapping.

## Directory Mode (One File Per Record)

In directory mode (enabled via `NewFromDirectory`), each file in the directory is one record. This is the approach used by both **Orbit** and **Laravel Paper**:

- **Orbit**: Stores each record as a separate file (`.md`, `.json`, `.yaml`) in a content directory. The filename is derived from the primary key.
- **Laravel Paper**: The filename (without extension) becomes the slug, which is the primary key. E.g., `content/posts/hello-world.md` → slug = `"hello-world"`.

The neat flat-file source applies the same pattern:

```
content/posts/
  ├── hello-world.json      → slug = "hello-world"
  ├── my-second-post.json   → slug = "my-second-post"
  └── draft-post.json       → slug = "draft-post"
```

The filename (without extension) is automatically inserted as the value for the primary key column declared via `WithPrimaryKey`. If no primary key is declared, no slug column is added.

The parser is selected per-file based on extension, so a directory could theoretically contain mixed formats (though this is uncommon and not recommended).

Directory mode is particularly useful for:
- **Content management**: Blog posts, pages, documentation — each content piece is a separate file
- **Configuration**: Each config entity in its own file for easier editing and version control
- **Git-friendly workflows**: Individual files diff cleanly in version control

## File-Metadata Timestamps

When `WithTimestamps()` is enabled, the source adds `created_at` and `updated_at` columns derived from the filesystem:

- **`updated_at`**: Set to the file's modification time (`mtime`)
- **`created_at`**: Set to the file's creation time (`ctime`), or falls back to `mtime` if the platform doesn't support `ctime`

This is similar to Laravel Paper's `#[Timestamps]` attribute. Note that Git checkouts reset mtimes to the deploy time, so for Git-deployed content, a date field in the file content itself is more reliable.

In directory mode, each record gets its own timestamps from its individual file. In single-file mode, all rows share the same timestamp (the file's mtime).

## Soft Deletes

When `WithSoftDeletes()` is enabled, the source checks for a `deleted_at` field in each parsed record. If present and non-null/non-empty, the record is **excluded from the rows returned by `Rows()`** — it never enters the SQLite table.

This is a simpler model than the ORM's soft-delete infrastructure (which loads rows and filters via `WHERE deleted_at IS NULL`). For flat files, excluding at parse time is cleaner because the file is the source of truth. Users who want to query soft-deleted records can construct a second `Source` without `WithSoftDeletes()`.

This is similar to Orbit's `SoftDeletes` trait.

## Custom Parsers

Users can register custom parsers for additional formats (YAML, Markdown with frontmatter, TOML, XML, etc.):

```go
// Register a custom YAML parser
flatfile.RegisterParser(&yamlParser{})

// yamlParser implements flatfile.FileParser
type yamlParser struct{}

func (p *yamlParser) Format() string       { return "yaml" }
func (p *yamlParser) Extensions() []string { return []string{".yaml", ".yml"} }

func (p *yamlParser) ParseFile(reader io.Reader, config flatfile.Config) ([]map[string]any, error) {
    // Parse YAML and return []map[string]any with Go-native types
    // ...
}
```

The source auto-detects the parser from the file extension. If the extension is ambiguous, the parser can be specified explicitly via `WithParser()`.

## Parsing Flow

```
┌──────────────┐     ┌──────────────┐     ┌─────────────────┐     ┌──────────────┐
│ Source.Rows()│───▶│ Detect Parser│────▶│  Determine Mode │────▶│  Parse Data  │
│              │     │ (by extension│     │  (dir or file)  │     │  (FileParser)│
└──────────────┘     │  or config)  │     └─────────────────┘     └──────────────┘
                     └──────────────┘                                      │
                                                                           ▼
┌──────────────┐     ┌──────────────┐     ┌─────────────────┐     ┌──────────────┐
│  Return      │◀────│  Apply       │◀────│  Apply          │◀────│  Apply       │
│  []map       │     │  Soft Deletes│     │  Timestamps     │     │  PK Injection│
│  (to array   │     │  (filter)    │     │  (file mtime)   │     │  (dir mode)  │
│   driver)    │     └──────────────┘     └─────────────────┘     └──────────────┘
└──────────────┘
                     ┌──────────────────────────────────────────────────────────┐
                     │  Array driver handles: schema inference, CREATE TABLE,   │
                     │  batched INSERT, concurrency, cleanup, PK constraint     │
                     └──────────────────────────────────────────────────────────┘
```

The flat-file source is responsible only for producing `[]map[string]any`. Everything from schema inference onward is the array driver's job.

## Example Usage

### CSV — Single File

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/dracory/neat"
    "github.com/dracory/neat/support/flatfile"
    _ "modernc.org/sqlite"
)

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
            "ff_db": {Driver: "array"},
        },
    }

    database, err := neat.New(config)
    if err != nil {
        log.Fatal(err)
    }
    defer database.Close()

    var users []User
    err = database.Query().
        Model(flatfile.NewFromCSV("data/users.csv", "users",
            flatfile.WithPrimaryKey("id"),
            flatfile.WithSchema(map[string]string{
                "id":      "int",
                "name":    "string",
                "email":   "string",
                "active":  "bool",
                "created": "time",
            }),
        )).
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

### CSV — Headerless with Custom Column Names

```go
src := flatfile.NewFromCSV("data/sensors.csv", "sensors",
    flatfile.WithColumns([]string{"timestamp", "temperature", "humidity", "pressure"}),
    flatfile.WithConfig(flatfile.Config{
        HasHeader: false,
        Comma:     ';',
        NullIf:    "N/A",
    }),
)
```

### CSV — Header Validation (Schema Drift Detection)

```go
src := flatfile.NewFromCSV("data/products.csv", "products",
    flatfile.WithConfig(flatfile.Config{
        ExpectedHeaders: []string{"id", "name", "price", "category"},
        StrictHeaders:   true,
    }),
)
```

### JSON — Single File with Flattening

```go
src := flatfile.NewFromJSON("data/orders.json", "orders",
    flatfile.WithConfig(flatfile.Config{
        Flatten:      true,
        FlattenDepth: 2,
    }),
    flatfile.WithSchema(map[string]string{
        "id":             "int",
        "customer_name":  "string",
        "customer_email": "string",
        "total":          "float",
        "address_city":   "string",
        "address_zip":    "string",
    }),
)
```

### JSON — JSONL / NDJSON

```go
src := flatfile.NewFromJSON("data/events.jsonl", "events",
    flatfile.WithConfig(flatfile.Config{
        NullIfMissing: true,
    }),
)
```

### JSON — Root Path Extraction

```go
src := flatfile.NewFromJSON("data/config.json", "config",
    flatfile.WithConfig(flatfile.Config{
        RootPath: "settings.items",
    }),
)
```

### Directory Mode (One File Per Record)

```go
src := flatfile.NewFromDirectory("content/posts", "posts",
    flatfile.WithPrimaryKey("slug"),
    flatfile.WithTimestamps(),
    flatfile.WithSoftDeletes(),
)
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

The source reads all `.json` files, adds `slug` from the filename, and adds `created_at`/`updated_at` from file metadata.

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

### Phase 1: Array Driver Primary Key Support
1. Add `ArrayPrimaryKey` interface to `contracts/database/orm/array_source.go`
2. Modify `Array.Populate` to detect `ArrayPrimaryKey` and validate the column name
3. Modify `Array.createTable` to accept a `pkCol` parameter and append `PRIMARY KEY`
4. Add tests in `database/driver/array_test.go`:
   - Primary key constraint is emitted when `ArrayPrimaryKey` is implemented
   - No constraint when not implemented (backward compatibility)
   - Invalid PK column name → error
   - PK column not in schema → error

### Phase 2: Flat-File Source and Parser Framework
1. Create `support/flatfile/source.go` — `Source` struct implementing `ArraySource`, `ArraySchema`, `ArrayPrimaryKey`
2. Create `support/flatfile/options.go` — `Config`, `Option`, `With*` functions
3. Create `support/flatfile/parser_registry.go` — `FileParserRegistry`, `RegisterParser`
4. Create `support/flatfile/constructors.go` — `NewFromCSV`, `NewFromJSON`, `NewFromDirectory`
5. Create `support/flatfile/timestamps.go` — `applyTimestamps`, `fileCreationTime`
6. Create `support/flatfile/directory.go` — `readDirectory` with filename-as-PK injection
7. Soft deletes filtering in `source.go` (`filterSoftDeleted`, `isSoftDeleted`)

### Phase 3: Built-in Parsers
1. Implement `support/flatfile/csv_parser.go`:
   - CSV parsing with `encoding/csv` (Comma, Comment, LazyQuotes)
   - Header detection / headerless mode with custom or generated column names
   - Header validation (StrictHeaders, ExpectedHeaders)
   - SkipRows
   - String-to-type inference (int → float → bool → time → string)
   - NullIf and TrimSpace
2. Implement `support/flatfile/json_parser.go`:
   - JSON array parsing with native type preservation
   - JSONL/NDJSON parsing with `bufio.Scanner`
   - Root path extraction (`navigatePath`)
   - Nested object flattening (`flattenObject` with depth control)
   - NullIfMissing, TrimStrings
3. Implement `support/flatfile/flatten.go` — flattening helper

### Phase 4: Tests
1. Source tests in `support/flatfile/source_test.go`:
   - Single-file mode (CSV, JSON)
   - Directory mode (read all files, filename as PK)
   - Timestamps (file mtime/ctime)
   - Soft deletes (filtering)
   - Primary key value injection (directory mode)
   - Custom column names
   - Explicit schema via `WithSchema`
   - File not found / directory not found errors
2. CSV parser tests in `support/flatfile/csv_parser_test.go`:
   - Basic CSV parsing
   - Type inference (int, float, bool, time, text)
   - Custom delimiter (tab, semicolon)
   - HasHeader = false with generated column names
   - HasHeader = false with custom column names
   - Header validation: strict mode passes / fails
   - SkipRows
   - NullIf (empty string → NULL)
   - LazyQuotes
3. JSON parser tests in `support/flatfile/json_parser_test.go`:
   - Basic JSON array parsing
   - Type inference from native JSON types
   - Time detection from RFC3339 string values
   - JSONL parsing (one object per line)
   - JSONL with trailing newline / empty lines (skipped)
   - Root path extraction (nested structure)
   - Root path not found (error)
   - Root path pointing to non-array (error)
   - Nested object flattening with depth limit
   - Missing keys with NullIfMissing = false / true
   - Invalid JSON (parse error)
4. Directory mode tests in `support/flatfile/directory_test.go`:
   - Basic populate from directory of JSON files
   - Filename as primary key value
   - Custom file extension via `FileExtension`
   - Empty directory (zero rows)
   - Non-matching file in directory (skipped)
   - Nested subdirectories (skipped)
   - Timestamps: each record gets individual timestamps
   - Soft deletes: records with non-null deleted_at excluded
5. Array driver primary key tests (from Phase 1)
6. Integration test: full `database.Query().Model(flatfile.NewFromCSV(...)).Get(&out)` flow
7. Example in `examples/flatfile-sources/` with sample CSV, JSON, JSONL, and directory-mode files

### Phase 5: Documentation
1. Create `examples/flatfile-sources/README.md`
2. Add flat-file sources to main docs (support packages page, API reference)
3. Document type inference rules, format-specific options, and custom parser registration
4. Mark `csv-driver.md` and `json-driver.md` as superseded by this proposal

## Design Decisions

### Why Array Source Adapters Instead of a New `flatfile` Driver?

The original version of this proposal (June 24, 2026) designed a standalone `flatfile` driver that embedded SQLite and duplicated the array driver's orchestration. At the time, the array source helper (`NewArraySourceFrom`) did not exist — using the array driver required hand-writing a custom struct with `Rows()`. The flat-file driver was motivated by eliminating that boilerplate.

The array source helper was completed on August 8, 2026, changing the calculus. The array driver already provides:
- SQLite embedding and all query builder features (WHERE, JOIN, ORDER BY, aggregates)
- `Populate`/`Cleanup` pattern with `sync.Map` for concurrency safety
- Schema inference with type widening
- Batched inserts (`batchSize := 500 / len(sortedCols)`)
- `isSimpleIdentifier` validation (SQL injection prevention)
- Row limit enforcement (`MaxArrayRows = 100000`)
- All 10 ORM integration points (driver registration, DSN, config, placeholders, `Model()` hook, connection pool, PRAGMAs, `Close()` cleanup, `detectDatabaseName`)

A standalone `flatfile` driver would duplicate all of this — the original proposal's own "Shared Code with Array Driver" section identified ~80% duplication and deferred extraction to a follow-up.

The adapter approach eliminates the duplication entirely:
- **No new driver** — flat-file sources use `Driver: "array"` in `DBConfig`
- **No new integration points** — no driver registration, DSN case, config validation, placeholder entry, PRAGMA case, `detectDatabaseName` case, or `Model()` hook change
- **No duplicated orchestration** — schema inference, table creation, batched inserts, concurrency, and cleanup are inherited
- **Future array driver improvements inherited for free** — persistent caching, stale-cache detection, and post-migration hooks (proposed in `array-driver-enhancement.md`) will work with flat-file sources automatically once implemented

The only array driver change required is `ArrayPrimaryKey` — one optional interface and a small `createTable` modification. This is a beneficial change for the array driver regardless of flat-file support (e.g., `NewArraySourceFrom(statuses)` with a struct whose `ID` field should be the PK).

### Why Pluggable Parsers Instead of Hardcoded Format Switch?

A `FileParser` interface with a registry allows:
- **Extensibility**: Users can register custom parsers for any format without modifying the source
- **Testability**: Parsers can be unit-tested in isolation
- **Separation of concerns**: The source handles shared logic (directory mode, timestamps, soft deletes, PK injection); parsers handle format-specific parsing
- **Auto-detection**: Parser selection from file extension is automatic but overridable via `WithParser`

### Why a Single `Config` Instead of Per-Format Configs?

A single config struct with format-specific fields is simpler than a hierarchy of config types. Fields that don't apply to the selected format are simply ignored. This avoids:
- Type assertion chains to access format-specific config
- Multiple config interfaces
- Complexity for users who only need one format

If the config grows too large in the future, it can be refactored into a map of format-specific options.

### Why Soft Deletes at Parse Time Instead of WHERE Filtering?

The ORM's soft-delete infrastructure loads rows into SQLite and applies `WHERE deleted_at IS NULL`. For flat files, this would mean loading soft-deleted records into memory and then filtering them out — wasteful when the file is the source of truth.

Excluding at parse time is cleaner: soft-deleted records never enter the SQLite table. Users who want to query soft-deleted records can construct a second `Source` without `WithSoftDeletes()`. This is a deliberate simplification; if WHERE-based soft deletes become necessary, it can be added as a future enhancement.

### Why Directory Mode Returns Only the First Record Per File?

In directory mode, each file is one record. If a parser returns multiple rows for a single file (e.g., a CSV file in a directory), only the first row is used. This matches the Orbit/Laravel Paper model where one file = one record. A future enhancement could add a `WithMultiRecordDirectory` option if multi-record-per-file directories are needed.

### Security: File Path Validation

The `FilePath` field returns a path that the source will open. The source does NOT restrict file paths by default (the caller controls which files to open). Applications can wrap `Source` with path validation logic if needed. This matches the array driver's stance on `Rows()` content.

## Benefits

1. **Zero Boilerplate**: Query CSV, JSON, JSONL files directly without manual parsing
2. **Full Query Builder**: All ORM features work (WHERE, JOIN, ORDER BY, aggregates, etc.) — inherited from the array driver
3. **Multi-Format**: One source family handles CSV, JSON, JSONL, and any custom format via pluggable parsers
4. **Extensible**: Register custom parsers for YAML, Markdown, TOML, XML, etc. without modifying the source or driver
5. **Type Safety**: Automatic type inference — JSON has native types; CSV uses string-to-type inference; both overridable via `WithSchema`
6. **Directory Mode**: One file per record with filename-as-primary-key — Git-friendly, content-management-friendly (like Orbit and Laravel Paper)
7. **File-Metadata Timestamps**: Derive `created_at`/`updated_at` from file mtime/ctime (like Laravel Paper's `#[Timestamps]`)
8. **Soft Deletes**: Respect `deleted_at` field for soft-deleted records (like Orbit's `SoftDeletes` trait)
9. **CSV-Specific**: Custom delimiters, headerless mode, header validation, `LazyQuotes`, `SkipRows`
10. **JSON-Specific**: JSONL/NDJSON, root path extraction, nested object flattening, missing key handling
11. **Primary Key Support**: Declare a primary key for faster lookups and ORM association compatibility — via `ArrayPrimaryKey`, benefiting both array and flat-file sources
12. **No Driver Duplication**: Reuses the array driver's SQLite orchestration, integration points, and future enhancements (caching, post-migrate hooks)
13. **Test Fixtures**: Easy to load test data from flat files (common in Go testing)
14. **Consistent Pattern**: Mirrors the existing `support/arraysource/` package structure

## Risks and Mitigations

### Risk 1: Memory Usage for Large Files
- **Issue**: Loading an entire file into in-memory SQLite could consume significant memory
- **Mitigation**: The array driver's `MaxArrayRows` limit (100,000 rows) applies; JSONL mode with `bufio.Scanner` for streaming; future streaming mode enhancement

### Risk 2: Type Inference Ambiguity (CSV)
- **Issue**: CSV values like "123" could be int or string (e.g., zip codes with leading zeros)
- **Mitigation**: `WithSchema` allows explicit type declaration; inference is opt-in default

### Risk 3: File Encoding (CSV)
- **Issue**: CSV files may use non-UTF-8 encodings (Latin-1, Windows-1252)
- **Mitigation**: Phase 1 supports UTF-8 only. Encoding support can be added as a future `Config.Encoding` field when needed.

### Risk 4: Nested Object Complexity (JSON)
- **Issue**: Deeply nested JSON can produce many columns when flattened, or complex JSON strings when stored as-is
- **Mitigation**: `FlattenDepth` limits nesting depth; default strategy (store as JSON string) is safe and queryable via SQLite JSON functions

### Risk 5: Config Bloat
- **Issue**: `Config` contains fields for all formats, which could grow large
- **Mitigation**: Acceptable for now — format-specific fields are clearly documented. If it grows too large, refactor into a map of format-specific options.

### Risk 6: Custom Parser Quality
- **Issue**: User-registered custom parsers may produce inconsistent data types or invalid schemas
- **Mitigation**: The array driver validates all column names via `isSimpleIdentifier` and infers schema from Go-native types returned by the parser. Parsers that return invalid data will produce clear errors.

### Risk 7: Array Driver Coupling
- **Issue**: Flat-file sources depend on the array driver's `Populate` behavior. If the array driver changes `Populate` in a breaking way, flat-file sources break too.
- **Mitigation**: This is a feature, not a bug — the `ArraySource` contract is stable and shared by `NewArraySourceFrom`, `NewArraySource`, and `NewWithSchema`. Changes to `Populate` are contract-bound. The coupling is the same as the existing array source helper's.

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
12. **File Encoding**: Support non-UTF-8 encodings (Latin-1, Windows-1252) via a `Config.Encoding` field
13. **Schema Validation**: Validate records against a schema definition (JSON Schema, CSV schema) before populating
14. **WHERE-Based Soft Deletes**: Load soft-deleted records into SQLite and filter via `WHERE deleted_at IS NULL` instead of excluding at parse time, enabling runtime toggling
15. **Persistent Caching** (inherited): Once `array-driver-enhancement.md` is implemented, flat-file sources inherit persistent caching and stale-cache detection for free via `ArrayCache`/`ArrayCacheReference` interfaces
16. **Post-Migration Hook** (inherited): Once `array-driver-enhancement.md` is implemented, flat-file sources inherit `ArrayPostMigrate` for adding indexes and constraints after table creation

## References

- Array driver implementation: `database/driver/array.go`
- Array source contracts: `contracts/database/orm/array_source.go`
- Array source helper: `support/arraysource/from_slice.go`
- Struct reflection helper: `support/structref/structref.go`
- Array source helper proposal: `docs/proposals/completed/array-source-helper.md`
- Array driver enhancement proposal: `docs/proposals/completed/array-driver-enhancement.md`
- Go `encoding/csv` package: https://pkg.go.dev/encoding/csv
- Go `encoding/json` package: https://pkg.go.dev/encoding/json
- JSON Lines specification: https://jsonlines.org/
- SQLite in-memory databases: https://www.sqlite.org/inmemorydb.html
- SQLite JSON1 extension: https://www.sqlite.org/json1.html
- Sushi (Laravel array driver): https://github.com/calebporzio/sushi
- FlatModel (Laravel CSV driver): https://packagist.org/packages/flatmodel/laravel-csv-flatmodel
- Orbit (Laravel flat-file Eloquent): https://packagist.org/packages/ryangjchandler/orbit
- Laravel Paper (flat-file Eloquent): https://packagist.org/packages/jacobjoergensen/laravel-paper
