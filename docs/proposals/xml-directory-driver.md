# XMLDB Driver

**Date**: August 9, 2026
**Status**: Proposal
**Priority**: Medium

## Problem Statement

The `csvdb` and `jsondb` drivers let you point at a directory of CSV or JSON files and query them as database tables. There is no equivalent for XML files — users with XML data (API responses, config exports, legacy data feeds, SOAP envelopes, sitemaps) must still write boilerplate parsing code or use the per-file `NewXmlFileSource` helper for each file individually.

The existing `support/xmlsource` package provides `NewXmlFileSource(filePath)` for single-file use with the array driver, but:
- You must call it once per file and pass each result to `Model()` separately
- There is no "point at a directory and query everything" experience
- Joins across XML files require manually creating multiple sources and registering them

### Desired Experience

```go
config := neat.DBConfig{
    Default: "xml_db",
    Connections: map[string]neat.ConnectionConfig{
        "xml_db": {
            Driver:   "xmldb",
            Database: "data/",   // directory path
        },
    },
}

database, _ := neat.New(config)
defer database.Close()

// data/users.xml → "users" table
var users []User
err := database.Query().Model(&User{}).Where("active = ?", true).Get(&users)

// data/products.xml → "products" table
var products []Product
err = database.Query().Model(&Product{}).Where("price > ?", 50).Get(&products)

// JOIN across XML files — works because all tables are in one SQLite DB
var orders []OrderWithUser
err = database.Query().
    Table("orders").
    LeftJoin("users", "orders.user_id = users.id").
    Get(&orders)
```

The directory is the database. Each `.xml` file is a table. The filename (without the extension) is the table name. XML attributes and leaf sub-elements define the columns. No per-file configuration, no `ArraySource` structs, no manual parsing.

## Proposed Solution

Implement a new `XMLDB` driver that mirrors the `CSVDB` and `JSONDB` drivers' architecture:

1. **Embeds `*SQLite`** for all `Driver` interface methods (Open, Close, Ping, BeginTx, Placeholder, Dialect)
2. **Overrides `Open`** to scan the directory for XML files, parse each one, and populate an in-memory SQLite database with one table per file
3. **Returns `Dialect() == "sqlite"`** so the query builder generates SQLite-compatible SQL and uses SQLite placeholders — no query builder changes needed

The driver is stateless. All state lives in the in-memory SQLite database. When the connection is closed, everything is gone.

### Mental Model

```
Directory (database)          SQLite (in-memory)
┌──────────────────┐         ┌──────────────────┐
│ data/            │         │                  │
│   users.xml   ───┼────────▶│ "users" table    │
│   products.xml ──┼────────▶│ "products" table │
│   orders.xml  ───┼────────▶│ "orders" table   │
└──────────────────┘         └──────────────────┘
```

- `data/users.xml` → table `"users"` (filename without extension)
- `.xml` files: parsed using `encoding/xml` streaming tokenizer
- The root element is the container; each direct child element is a row
- Column values come from:
  - Attributes on the child element: `<user id="1">` → column `"id"`
  - Leaf sub-elements (text-only): `<name>Alice</name>` → column `"name"`
  - Nested sub-elements: stored as JSON strings (queryable via SQLite JSON functions)
- String values are type-inferred (int, float, bool, time, string) — same as CSV
- All tables populated at `Open` time (eager, not lazy)

### Architecture

```
database/
  driver/
    xmldb.go              // XMLDB driver (embeds *SQLite, overrides Open)
    xmldb_xml.go          // XML parsing + schema inference helper
    xmldb_test.go         // Driver unit tests

integration_tests/
  xmldb/
    helper.go             // SetupXMLDBTest, XML fixtures, model types
    xmldb_query_test.go   // Query, WHERE, ORDER BY, LIMIT, First, Find, Count
    xmldb_join_test.go    // LEFT JOIN, INNER JOIN, 3-table JOIN, JOIN+WHERE
    xmldb_aggregate_test.go  // COUNT, SUM, AVG, MIN/MAX, GROUP BY
    xmldb_type_inference_test.go  // Type inference, nested elements, attributes

examples/
  xmldb-driver/
    main.go               // Example usage
    main_test.go
    README.md
    data/
      users.xml           // Sample XML
      products.xml        // Sample XML
      orders.xml          // Sample XML
```

No new contracts. No new interfaces. No changes to `contracts/database/orm/`. No changes to `database/query/`. The driver plugs into the existing driver registration system — same integration points as `csvdb` and `jsondb`.

## Driver Implementation

```go
// database/driver/xmldb.go

package driver

import (
    "database/sql"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    _ "modernc.org/sqlite"
)

// XMLDB implements the Driver interface for XML-directory-backed storage.
// It embeds *SQLite for all standard Driver methods and overrides Open to
// scan a directory of XML files and populate an in-memory SQLite database.
//
// The directory path is passed as the DSN (from ConnectionConfig.Database).
// Each .xml file in the directory becomes a table named after the filename
// (without the .xml extension). The root element is the container; each
// direct child element is a row. Attributes and leaf sub-elements define
// the columns.
//
// The driver is stateless — all state lives in the in-memory SQLite database.
type XMLDB struct {
    *SQLite
}

func NewXMLDB() *XMLDB {
    return &XMLDB{SQLite: NewSQLite()}
}

func (x *XMLDB) Dialect() string {
    return "sqlite"
}

func (x *XMLDB) Open(dirPath string) (*sql.DB, error) {
    db, err := sql.Open("sqlite", ":memory:")
    if err != nil {
        return nil, fmt.Errorf("xmldb: failed to open in-memory SQLite: %w", err)
    }

    if dirPath == "" || dirPath == ":memory:" {
        return db, nil
    }

    info, err := os.Stat(dirPath)
    if err != nil {
        _ = db.Close()
        return nil, fmt.Errorf("xmldb: cannot access directory %s: %w", dirPath, err)
    }
    if !info.IsDir() {
        _ = db.Close()
        return nil, fmt.Errorf("xmldb: %s is not a directory", dirPath)
    }

    entries, err := os.ReadDir(dirPath)
    if err != nil {
        _ = db.Close()
        return nil, fmt.Errorf("xmldb: cannot read directory %s: %w", dirPath, err)
    }

    seenTables := make(map[string]string) // lower(tableName) → original filename

    for _, entry := range entries {
        if entry.IsDir() {
            continue
        }
        ext := strings.ToLower(filepath.Ext(entry.Name()))
        if ext != ".xml" {
            continue
        }

        tableName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
        filePath := filepath.Join(dirPath, entry.Name())

        lowerName := strings.ToLower(tableName)
        if prevFile, exists := seenTables[lowerName]; exists {
            _ = db.Close()
            return nil, fmt.Errorf("xmldb: table name collision: %s and %s produce the same table name (case-insensitive)", prevFile, entry.Name())
        }
        seenTables[lowerName] = entry.Name()

        if err := populateXMLFile(db, tableName, filePath); err != nil {
            _ = db.Close()
            return nil, fmt.Errorf("xmldb: failed to populate table %s from %s: %w", tableName, filePath, err)
        }
    }

    return db, nil
}
```

### XML Parsing and Schema Inference

```go
// database/driver/xmldb_xml.go

// MaxXMLRows limits the number of data rows (child elements) that can be
// loaded from a single XML file to prevent unbounded memory/CPU consumption.
const MaxXMLRows = 100000

// populateXMLFile reads an XML file, infers schema, creates a table,
// and inserts all rows in a transaction. Validates:
// - Table name is a simple identifier (SQL injection prevention)
// - Column names are simple identifiers
// - No duplicate column names (case-insensitive)
// - Row count does not exceed MaxXMLRows
func populateXMLFile(db *sql.DB, tableName string, filePath string) error {
    // 1. ParseXMLFile → rows (via support/xmlsource)
    // 2. If 0 rows or 0 columns, skip table (no schema = no table)
    // 3. Validate table name with isSimpleIdentifier
    // 4. inferSchema → columns, types (reuses CSV-style string type inference)
    // 5. Validate column names (simple identifiers, no case-insensitive dupes)
    // 6. CREATE TABLE with inferred schema
    // 7. BEGIN TRANSACTION
    // 8. INSERT rows in batches (SQLite parameter limit ~999)
    // 9. COMMIT
}
```

### Key Differences from CSVDB and JSONDB

XML is a markup language, not a data format with native types. This creates unique challenges compared to CSV (string-based) and JSON (native types):

| Concern | CSVDB | JSONDB | XMLDB |
|---------|-------|--------|-------|
| Type inference | String parsing (int → float → bool → time → string) | Native JSON types (int64, float64, bool, string, nil) | String parsing (same as CSV — XML has no native types) |
| Date detection | String format guessing (RFC3339, `2006-01-02`, etc.) | RFC3339 only | String format guessing (same as CSV — multiple date formats) |
| Nested data | Not supported (flat rows only) | Nested objects/arrays → JSON strings | Nested elements → JSON strings (queryable via SQLite JSON functions) |
| Column sources | Header row (first line) | Object keys (union across all rows) | Attributes + leaf sub-element names (union across all rows) |
| Missing columns | Ragged rows (fewer fields → NULL) | Missing keys → NULL (natural in JSON) | Missing attributes/elements → NULL |
| Schema source | First row (header) | Union of all object keys across all rows | Union of all attribute names + leaf element names across all rows |
| BOM stripping | Required (Excel exports) | Not needed | Not needed (XML declaration handles encoding) |
| Format variants | One format (CSV) | Three extensions: `.json`, `.jsonl`, `.ndjson` | One extension: `.xml` |
| Row delimiter | Newline | Array elements or newline (JSONL) | Child elements within root element |
| Streaming | Line-by-line | `json.Decoder` token stream or `bufio.Scanner` | `xml.Decoder` token stream (natural streaming) |

### XML Structure Convention

The driver expects a simple, common XML pattern: a root element containing repeated child elements. This is the same convention used by `support/xmlsource`:

```xml
<users>
  <user id="1">
    <name>Alice</name>
    <email>alice@example.com</email>
    <active>true</active>
  </user>
  <user id="2">
    <name>Bob</name>
    <email>bob@example.com</email>
    <active>false</active>
  </user>
</users>
```

- Root element `<users>` → container (not a row)
- Each `<user>` child → one row
- Attribute `id="1"` → column `"id"` with value `1` (int64)
- Leaf element `<name>Alice</name>` → column `"name"` with value `"Alice"` (string)
- Leaf element `<active>true</active>` → column `"active"` with value `1` (int64, bool→int)
- Nested element `<address><city>NYC</city></address>` → column `"address"` with JSON string `{"city":"NYC"}`

### Type Inference

XML has no native types — all values are strings. Type inference uses the same approach as CSVDB (string-to-type inference via `strconv.ParseInt`, `strconv.ParseFloat`, bool detection, date format guessing):

| XML value | SQLite type |
|-----------|-------------|
| `"42"` (attribute or text) | INTEGER |
| `"19.99"` | REAL |
| `"true"`, `"false"` | INTEGER (0/1) |
| `"hello"` | TEXT |
| `"2024-01-15T10:30:00Z"` | DATETIME (RFC3339) |
| `"2024-01-15"` | DATETIME (date-only format) |
| `"2024-01-15 10:30:00"` | DATETIME (space-separated datetime) |
| `""` (empty string) | NULL |
| Nested element | TEXT (JSON string) |

Mixed types in a column widen: INTEGER → REAL → TEXT. This is identical to CSVDB's widening behavior and reuses the same `widenType` function.

### Column Name Union

Unlike CSV (fixed header) or JSON (object keys), XML columns come from two sources:
1. **Attributes** on the row element: `<user id="1" active="true">` → columns `"id"`, `"active"`
2. **Leaf sub-element names**: `<name>Alice</name>` → column `"name"`

The schema is the union of all attribute names and leaf sub-element names across all rows. If row 1 has `<email>` but row 2 does not, `"email"` is still a column (with NULL for row 2). This mirrors JSONDB's union-of-keys approach.

If an attribute and a sub-element share the same name (e.g., `<user id="1"><id>1</id></user>`), the sub-element value takes precedence (last write wins during row parsing). This is the existing `support/xmlsource` behavior — no change needed.

## Integration Points

The driver requires one-line additions to the same 7 files touched by `jsondb`. No new contracts, no query builder changes, no new interfaces.

### 1. Driver Constant

```go
// contracts/database/config.go
DriverXMLDB    Driver = "xmldb"
```

### 2. Driver Registration

```go
// database/orm/orm.go — createDriver()
case "xmldb":
    return driver.NewXMLDB()
```

### 3. DSN Builder

```go
// database/db/config_builder.go — BuildDSN()
case "xmldb":
    return b.buildXMLDBDSN()
```

`buildXMLDBDSN` returns `ConnectionConfig.Database` if non-empty, else `:memory:`. For XMLDB, `Database` holds the directory path.

### 4. Config Validation

```go
// database/db/config_builder.go — ConnectionConfig.Validate()
case "sqlite", "array", "csvdb", "jsondb", "xmldb":
    // database path is optional; empty defaults to :memory:
    // For XMLDB, the Database field holds the directory path.
    return nil
```

### 5. Connection Pool Configuration

```go
// database/orm/orm.go — configureConnectionPool()
pinSingleConn := connConfig.Driver == "sqlite" || connConfig.Driver == "array" || connConfig.Driver == "csvdb" || connConfig.Driver == "jsondb" || connConfig.Driver == "xmldb"
```

XMLDB uses in-memory SQLite, so it shares the same single-connection constraint.

### 6. Database Name Detection

```go
// database/db.go — detectDatabaseName()
case "sqlite", "turso", "array", "csvdb", "jsondb", "xmldb":
    return "main"
```

### 7. Schema Builder

```go
// database/schema/schema.go — New()
case contractsdatabase.DriverXMLDB:
    // XMLDB uses SQLite grammar since Dialect() returns "sqlite".
    sqliteGrammar := grammars.NewSqlite(log, prefix)
    driverSchema = NewSqliteSchema(sqliteGrammar, orm, prefix)
    grammar = sqliteGrammar
    processor = processors.NewSqlite()
```

## Robustness Features

All robustness features from `csvdb` and `jsondb` carry over to `xmldb`:

- **Table name collision detection**: Files that produce colliding table names (e.g., `Users.xml` and `users.xml` on case-sensitive filesystems) are detected and rejected.
- **MaxXMLRows limit**: 100,000-row limit per XML file to prevent unbounded memory/CPU consumption. Rows are read incrementally via `xml.Decoder` token stream so the limit is enforced without loading the entire file into memory.
- **Duplicate column detection**: If two columns (attributes or elements) differ only in case (e.g., `Name` and `name`), this is detected as a duplicate since SQLite column names are case-insensitive.
- **Invalid identifier detection**: Table names derived from filenames and column names from attributes/elements are validated with `isSimpleIdentifier` to prevent SQL injection.
- **Transaction wrapping**: Batch INSERTs are wrapped in a `BEGIN`/`COMMIT` transaction for atomicity and performance.
- **Empty file handling**: An XML file with no child elements (e.g., `<users></users>`) is skipped — no table created. Consistent with JSONDB's empty-array behavior (no data = no table).
- **0-column skip**: If rows exist but no columns could be inferred (e.g., all child elements are empty with no attributes), the table is skipped. Consistent with JSONDB's `[{}]` handling.
- **XML namespace handling**: Attribute and element names use their local names (namespace prefixes stripped). This matches the existing `support/xmlsource` behavior which uses `attr.Name.Local` and `start.Name.Local`.
- **Non-XML files skipped**: Files without `.xml` extension are ignored. Files with `.xml` extension but invalid XML content return an error.

## Testing

### Unit Tests (`database/driver/`)
- XML parsing: basic structure, attributes, leaf elements, nested elements
- Schema inference: type inference (int, float, bool, time, string), type widening, union of columns
- Nested elements → JSON strings
- Date format detection (RFC3339, date-only, space-separated datetime, US date format)
- Table name collision, case-insensitive extension matching
- MaxXMLRows limit enforcement (streaming)
- Invalid identifiers, duplicate columns (case-insensitive)
- Empty root element (no child elements) → skip table
- 0-column rows → skip table
- Non-XML files skipped, subdirectories skipped
- Transaction wrapping, batch insertion

### Integration Tests (`integration_tests/xmldb/`)
- **Query**: basic query, WHERE equals, WHERE bool, WHERE on nested JSON string, ORDER BY asc/desc, LIMIT, LIMIT+OFFSET, First, Find, Count
- **JOIN**: LEFT JOIN, INNER JOIN, 3-table JOIN, JOIN with WHERE
- **Aggregates**: COUNT, COUNT with WHERE, SUM, AVG, MIN/MAX, GROUP BY, SUM with JOIN+GROUP BY
- **Type inference**: integer (from attribute and text), float, bool, string, datetime (RFC3339 and other formats), nested element as JSON string, NULL handling, attribute + element columns

### Example (`examples/xmldb-driver/`)
- Working example with sample XML data (users.xml, products.xml, orders.xml)
- Demonstrates WHERE, JOIN (across XML files), and aggregate queries
- Shows nested element querying via SQLite JSON functions
- Shows attribute-based columns

## Open Questions

### 1. XML Namespace Handling

XML files may use namespaces (e.g., `<ns:user xmlns:ns="...">`). The existing `support/xmlsource` uses `attr.Name.Local` and `start.Name.Local`, which strips the namespace prefix. This means `<ns:name>` and `<name>` would both produce column `"name"`. This is the existing behavior and is acceptable — namespace-aware column naming would add complexity with little benefit for the directory-as-database use case.

**Recommendation**: Keep existing behavior (use local names, ignore namespaces). Document this limitation.

### 2. Mixed Content Elements

XML allows mixed content (text and child elements interleaved): `<p>Hello <b>world</b></p>`. The existing `support/xmlsource` ignores text directly inside row elements (treats it as whitespace between child elements). Mixed content in leaf elements is captured as the full text content.

**Recommendation**: Keep existing behavior. Mixed content is rare in data-oriented XML (which is the target use case).

### 3. Reuse of `support/xmlsource` Package

The existing `support/xmlsource` package has `parseXMLString`, `parseXMLFile`, `parseXMLReader`, and `inferAndConvert` functions. The `xmldb` driver could either:
- **A**: Reuse `support/xmlsource` functions directly (import the package)
- **B**: Copy the logic into `database/driver/xmldb_xml.go` (self-contained)

**Recommendation**: Option A (reuse), following the same approach as JSONDB. The `support/xmlsource` parsing functions already return `[]map[string]any` with type-inferred values, which is exactly what `xmldb` needs. However, `support/xmlsource` panics on parse errors (it's designed for the `Model()` API where panics are acceptable). The `xmldb` driver needs errors, not panics.

**Final recommendation**: Export the parsing functions from `support/xmlsource` as functions that return errors (`ParseXMLFile`, `ParseXMLString`, `ParseXMLReader`), and have the existing `NewXmlSource` / `NewXmlFileSource` wrappers call them and panic on error. This mirrors the JSONDB approach of exporting `ParseJSONFile`, `NormalizeRows`, etc. The `inferAndConvert` function should also be exported as `InferAndConvert` for reuse in schema inference.

### 4. Row Element Name vs. Table Name

In XML, the root element name and child element names are arbitrary. The convention is:
- Root element: container (e.g., `<users>`)
- Child elements: rows (e.g., `<user>`)
- Table name: derived from filename, not from element names

This means `data/people.xml` with `<people><person>...</person></people>` creates a table named `"people"` (from filename), not `"person"` (from child element name). This is consistent with CSVDB and JSONDB where the table name comes from the filename.

**Recommendation**: Keep filename-based table naming. The child element name is irrelevant for table naming. This is the existing `support/xmlsource` behavior.

### 5. CDATA Sections

XML supports CDATA sections for unescaped text: `<description><![CDATA[Hello & world]]></description>`. The `encoding/xml` decoder handles CDATA transparently — it appears as `CharData` tokens, which the existing parser already captures for leaf elements.

**Recommendation**: No special handling needed. CDATA is handled by `encoding/xml` and flows through the existing parser as text content.

### 6. XML Declaration and Encoding

XML files may start with `<?xml version="1.0" encoding="UTF-8"?>`. The `encoding/xml` decoder handles this transparently. No BOM stripping is needed (unlike CSV), because XML's encoding declaration handles character encoding.

**Recommendation**: No special handling needed. `encoding/xml` handles the XML declaration and encoding detection.

## Relationship to Existing Work

| Component | Relationship |
|-----------|-------------|
| `csvdb` driver | Type inference template — same string-to-type inference logic (`inferAndConvert`, `widenType`) |
| `jsondb` driver | Architectural template — same embedding pattern, same integration points, same 0-column skip behavior |
| `support/xmlsource` | Parsing logic to reuse (or extract from) — `parseXMLReader`, `parseElement`, `parseChildElement`, `inferAndConvert` |
| `support/arraysource` | Not used — `xmldb` populates SQLite directly via `database/sql`, not via the array driver's `Model()` hook |
| `flat-file-driver.md` proposal | The v3 revision notes that "directory-as-database mode is covered by a separate proposal" — this is that proposal for XML (alongside `csv-directory-driver.md` for CSV and `json-directory-driver.md` for JSON) |

## Future Extensions

- **Mixed-format directory**: A single `filedb` driver that handles all extensions (`.csv`, `.json`, `.jsonl`, `.ndjson`, `.xml`) in one directory — would unify `csvdb`, `jsondb`, and `xmldb` into one driver
- **Streaming queries**: For directories with many large files, populate tables lazily (on first query) instead of eagerly at `Open` time
- **Persistent caching**: Save the populated SQLite database to a file on disk and reuse it on subsequent runs if the source files haven't changed (mirrors the `array-driver-enhancement` proposal's `ArrayCache` interface)
- **XPath filtering**: Allow filtering which child elements become rows via an XPath-like configuration (e.g., only `<user>` elements with `active="true"`)
- **Attribute-only mode**: For XML files where all data is in attributes (e.g., `<user id="1" name="Alice" active="true"/>`), the driver already handles this — no sub-elements needed
