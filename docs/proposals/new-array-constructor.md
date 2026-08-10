# NewArrayDB Convenience Constructor

**Date**: August 10, 2026
**Status**: Proposal
**Priority**: Low

## Problem Statement

The array driver is an in-memory, SQLite-backed store that needs no DSN, host, port, or credentials. Yet users must write 8 lines of boilerplate config to use it:

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

None are tailored for the array driver.

## Proposed Solution

Add a single convenience function to `config.go`:

```go
// NewArrayDB creates an in-memory array-driver Database with zero configuration.
// It is the simplest way to query slices of structs, maps, CSV, JSON, or XML
// data using the full query builder (Where, OrderBy, First, Get, etc.).
//
//	database, err := neat.NewArrayDB()
//	if err != nil { ... }
//	defer database.Close()
//
//	database.Query().
//	    Model(neat.NewArraySourceFrom(statuses)).
//	    Where("name = ?", "Active").
//	    First(&result)
func NewArrayDB(opts ...database.Option) (*Database, error) {
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
    return neat.NewArrayDB()
}
```

## Design Decisions

- **Name**: `NewArrayDB` conveys that you get a `*Database` backed by the array driver. It follows the `New`/`NewFromDSN`/`NewFromSQLDB` naming pattern.
- **Options**: Accepts `...database.Option` so users can still pass `database.WithDebug(true)` etc.
- **No new types**: Reuses existing `DBConfig` internally — no new structs, no new interfaces.
- **Scope**: Only the array driver. CSV/JSON/XML sources already work through `NewArraySourceFrom` / `NewCsvSource` / etc. at the model level — they don't need a separate `Database`.

## Files Changed

| File | Change |
|---|---|
| `config.go` | Add `NewArrayDB()` function (~10 lines) |
| `examples/array-driver/main.go` | Simplify `newDatabase()` to call `neat.NewArrayDB()` |

## Risks

- **None**: The function is a pure convenience wrapper around `neat.New` with a hardcoded config. No behavioral change, no new dependencies, no API breakage.
