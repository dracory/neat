# NewMemoryDB Convenience Constructor

**Date**: August 10, 2026
**Status**: Proposal
**Priority**: Low

## Problem Statement

The array driver is an in-memory, SQLite-backed database that needs no DSN, host, port, or credentials. It powers all in-memory data sources — array slices, CSV, JSON, XML. Yet users must write 8 lines of boilerplate config to use it:

```go
config := neat.DBConfig{
    Default: "array_db",
    Connections: map[string]neat.ConnectionConfig{
        "array_db": {
            Driver: "array",
        },
    },
}
database, err := neat.New(config)
```

This is the minimum setup shown in `examples/array-driver/main.go:71-86`. The three existing public constructors are:

- `neat.New(DBConfig)` — full config (what the example uses)
- `neat.NewFromDSN(dsn)` — needs a DSN string (irrelevant for arrays)
- `neat.NewFromSQLDB(sqlDB)` — needs an open `*sql.DB` (irrelevant for arrays)

None are tailored for in-memory data sources.

## Proposed Solution

Add a single convenience function to `config.go`:

```go
// NewMemoryDB creates an in-memory database with zero configuration.
// It is the simplest way to query slices of structs, maps, CSV, JSON, or XML
// data using the full query builder (Where, OrderBy, First, Get, JOINs, etc.).
//
// Multiple sources can be loaded into the same database — each becomes a
// table, enabling JOINs across them.
//
//	database, err := neat.NewMemoryDB()
//	if err != nil { ... }
//	defer database.Close()
//
//	// Load multiple sources — each becomes a table in the same SQLite DB
//	database.Query().
//	    Model(neat.NewArraySourceFrom(statuses)).
//	    Where("name = ?", "Active").
//	    First(&result)
//
//	database.Query().
//	    Model(neat.NewCsvSource(csv, "users")).
//	    Get(&users)
//
//	// JOIN across sources — both tables exist in the same in-memory DB
//	database.Query().
//	    Table("statuses").
//	    LeftJoin("users ON statuses.user_id = users.id").
//	    Get(&joined)
func NewMemoryDB(opts ...database.Option) (*Database, error) {
    return New(DBConfig{
        Default: "array_db",
        Connections: map[string]ConnectionConfig{
            "array_db": {Driver: "array"},
        },
    }, opts...)
}
```

## Before / After

### Before (8 lines)

```go
func newDatabase() (*neat.Database, error) {
    config := neat.DBConfig{
        Default: "array_db",
        Connections: map[string]neat.ConnectionConfig{
            "array_db": {Driver: "array"},
        },
    }
    database, err := neat.New(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create database: %w", err)
    }
    return database, nil
}
```

### After (1 line)

```go
func newDatabase() (*neat.Database, error) {
    return neat.NewMemoryDB()
}
```

## Design Decisions

- **Name**: `NewMemoryDB` conveys that you get an in-memory `*Database` — not specific to arrays, since it powers all in-memory sources (array slices, CSV, JSON, XML). It follows the `New`/`NewFromDSN`/`NewFromSQLDB` naming pattern.
- **Multiple tables**: Each `Model()` call with a different source populates a new table in the same SQLite database (see `query_model.go:18-41`). This enables JOINs across sources — e.g., an array source joined with a CSV source.
- **Options**: Accepts `...database.Option` so users can still pass `database.WithDebug(true)` etc.
- **No new types**: Reuses existing `DBConfig` internally — no new structs, no new interfaces.
- **Scope**: One constructor for all in-memory sources. The driver name stays `"array"` for backward compatibility — `NewMemoryDB` is just a convenience wrapper.

## Files Changed

| File | Change |
|---|---|
| `config.go` | Add `NewMemoryDB()` function (~10 lines) |
| `examples/array-driver/main.go` | Simplify `newDatabase()` to call `neat.NewMemoryDB()` |

## Relationship to GODB Proposal

The [GODB proposal](go-directory-driver.md) is complementary, not competing. Both create in-memory SQLite databases with multiple tables and JOIN support, but serve different use cases:

| | `NewMemoryDB` (this proposal) | GODB |
|---|---|---|
| **Config** | Zero — `neat.NewMemoryDB()` | Explicit — `DBConfig` with `Tables` field |
| **When tables load** | Lazily — each `Model()` call populates a table | Eagerly — all tables at `Open()` time |
| **Data source** | Runtime data — slices, CSV strings, JSON strings | Compiled-in Go variables — `blogs.Blogs` |
| **Table declaration** | Implicit — whatever you pass to `Model()` | Explicit — `godb.Tables{"blogs": Blogs}` |

`NewMemoryDB` is the simple shortcut for ad-hoc runtime data. GODB is the structured approach for compiled-in reference data. If GODB is implemented, it would benefit from a similar convenience constructor (e.g., `NewGODB(tables)`).

## Risks

- **None**: The function is a pure convenience wrapper around `neat.New` with a hardcoded config. No behavioral change, no new dependencies, no API breakage.
