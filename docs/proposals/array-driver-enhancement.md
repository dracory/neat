# Array Driver Enhancement

**Date**: June 24, 2026
**Status**: Proposal
**Priority**: Medium

## Problem Statement

The array driver provides a clean way to query in-memory Go data through the ORM using SQLite as a backing store. However, a comparison with [Sushi](https://github.com/calebporzio/sushi) — Laravel's equivalent "array driver" for Eloquent — reveals three feature gaps that limit the array driver's usefulness for production workloads and external data sources.

### Current Limitations

- **No persistent caching** — the array driver always populates an in-memory SQLite database from scratch on every process start. For large datasets or external data sources (APIs, flat files), this means re-fetching and re-inserting data on every run.
- **No stale-cache detection** — there is no mechanism to determine whether cached data is still fresh or needs to be rebuilt. Sushi uses file modification timestamps to decide this automatically.
- **No post-migration hook** — the array driver creates the table and inserts rows, but provides no way for the user to customize the table after creation (e.g., adding indexes, constraints, or default values).

### Comparison with Sushi

| Feature | Sushi | neat Array Driver | Gap |
|---|---|---|---|
| Core concept (array → SQLite) | ✅ | ✅ | — |
| Schema auto-detection | ✅ (first row only) | ✅ (all rows + type widening) | — |
| Custom schema | ✅ | ✅ | — |
| Empty datasets | ✅ | ✅ | — |
| Relationships | ✅ | ✅ | — |
| Concurrency safety | ❌ | ✅ | — |
| SQL injection protection | ❌ | ✅ | — |
| Row limit | ❌ | ✅ | — |
| Type widening | ❌ | ✅ | — |
| Cleanup on close | ❌ | ✅ | — |
| **Persistent caching** | ✅ | ❌ | **Missing** |
| **Stale-cache detection** | ✅ | ❌ | **Missing** |
| **Post-migration hook** | ✅ | ❌ | **Missing** |
| **Cache reference path** | ✅ | ❌ | **Missing** |

## Proposed Enhancements

### 1. Persistent Caching

Allow the array driver to persist the populated SQLite database to a file on disk, so that subsequent process starts can skip re-population entirely.

#### Design

```go
// contracts/database/orm/array_source.go (additions)

// ArrayCache is an optional interface for sources that want persistent caching.
// If implemented, the driver will cache the populated SQLite database to the
// file returned by CachePath() and reuse it on subsequent runs if the cache
// is still fresh (see ArrayCacheReference below).
type ArrayCache interface {
    CachePath() string // path to the .sqlite cache file
}

// ArrayCacheReference is an optional interface for controlling cache freshness.
// If implemented, the driver compares the modification time of the file returned
// by CacheReferencePath() against the cache file. If the reference file is newer,
// the cache is considered stale and the table is rebuilt.
//
// If not implemented, the driver falls back to comparing the cache file's mtime
// against the source's Rows() call time (i.e., always rebuild if no reference
// path is provided).
type ArrayCacheReference interface {
    CacheReferencePath() string // file whose mtime determines cache freshness
}

// ArrayCacheEnabled controls whether caching is active for a source.
// If not implemented, caching is disabled (current behavior).
type ArrayCacheEnabled interface {
    ShouldCache() bool
}
```

#### Usage Example

```go
type StatusSource struct{}

func (s *StatusSource) TableName() string { return "statuses" }
func (s *StatusSource) Rows() ([]map[string]any, error) {
    return []map[string]any{
        {"id": 1, "name": "Pending", "color": "yellow"},
        {"id": 2, "name": "Active", "color": "green"},
        {"id": 3, "name": "Inactive", "color": "red"},
    }, nil
}

// Enable persistent caching
func (s *StatusSource) CachePath() string {
    return "/tmp/neat_cache_statuses.sqlite"
}

// Only rebuild cache when the source file changes
func (s *StatusSource) CacheReferencePath() string {
    return "data/statuses.json" // external data file
}

func (s *StatusSource) ShouldCache() bool {
    return true
}
```

#### Population Flow with Caching

```
┌──────────────────────────────────────────────────────────┐
│                    Populate(ctx, db, source)              │
└────────────────────────┬─────────────────────────────────┘
                         │
                         ▼
              ┌─────────────────────┐
              │ Implements          │
              │ ArrayCacheEnabled?  │
              └────────┬────────────┘
                       │
              ┌──── Yes ────┐──── No ────┐
              │             │             │
              ▼             │             ▼
    ┌─────────────────┐     │    ┌─────────────────┐
    │ Cache file      │     │    │ Current behavior │
    │ exists?         │     │    │ (in-memory only) │
    └────────┬────────┘     │    └─────────────────┘
             │              │
     ┌── Yes ──┐── No ──┐   │
     │          │       │   │
     ▼          ▼       │   │
┌─────────┐ ┌────────┐  │   │
│ Cache   │ │ Build  │  │   │
│ fresh?  │ │ table  │  │   │
└────┬────┘ └───┬────┘  │   │
     │          │       │   │
  Yes│          │       │   │
     ▼          │       │   │
┌──────────┐   │       │   │
│ Attach   │   │       │   │
│ cache    │   │       │   │
│ file to  │   │       │   │
│ db conn  │   │       │   │
└──────────┘   │       │   │
               │       │   │
               ▼       ▼   │
        ┌─────────────────┐│
        │ Populate rows   ││
        │ + persist to    ││
        │ cache file      ││
        └─────────────────┘│
```

#### Implementation Details

The `Populate` method would be extended:

1. Check if source implements `ArrayCacheEnabled` and `ShouldCache()` returns `true`.
2. If yes, check if the cache file at `CachePath()` exists.
3. If the cache file exists, check freshness:
   - If source implements `ArrayCacheReference`, compare mtime of `CacheReferencePath()` against cache file mtime.
   - If no reference path, compare cache file mtime against process start time (always stale → rebuild).
4. If cache is fresh, attach the cache file to the current `*sql.DB` connection via `ATTACH DATABASE` and skip population.
5. If cache is stale or missing, populate as usual, then copy the in-memory database to the cache file using `VACUUM INTO` or the SQLite backup API.

#### Cache Invalidation

The cache is invalidated when:
- The cache file does not exist (first run or manually deleted).
- The reference file's mtime is newer than the cache file's mtime.
- `ShouldCache()` returns `false` (caching disabled at runtime).

### 2. Post-Migration Hook

Allow sources to customize the SQLite table after creation — adding indexes, constraints, or default values.

#### Design

```go
// contracts/database/orm/array_source.go (addition)

// ArrayPostMigrate is an optional interface that allows the source to
// customize the table after it has been created and populated. This is
// useful for adding indexes, unique constraints, or default values.
//
// The hook is called once, after table creation and row insertion, but
// before the table is marked as populated. Any error returned will
// abort the Populate call.
type ArrayPostMigrate interface {
    PostMigrate(ctx context.Context, db *sql.DB, tableName string) error
}
```

#### Usage Example

```go
type ProductSource struct{}

func (s *ProductSource) TableName() string { return "products" }
func (s *ProductSource) Rows() ([]map[string]any, error) {
    return []map[string]any{
        {"id": 1, "name": "Lawn Mower", "price": 226.99, "category": "tools"},
        {"id": 2, "name": "Leaf Blower", "price": 134.99, "category": "tools"},
        {"id": 3, "name": "Rake", "price": 9.99, "category": "tools"},
    }, nil
}

// Add indexes after table creation
func (s *ProductSource) PostMigrate(ctx context.Context, db *sql.DB, tableName string) error {
    _, err := db.ExecContext(ctx, fmt.Sprintf(
        `CREATE INDEX IF NOT EXISTS "idx_%s_category" ON "%s" ("category")`,
        tableName, tableName,
    ))
    if err != nil {
        return err
    }

    _, err = db.ExecContext(ctx, fmt.Sprintf(
        `CREATE INDEX IF NOT EXISTS "idx_%s_price" ON "%s" ("price")`,
        tableName, tableName,
    ))
    return err
}
```

#### Integration Point

In `Array.Populate()`, after `insertRows` and before `markPopulated`:

```go
// database/driver/array.go — Populate()

// ... existing code: create table, insert rows ...

// Post-migration hook
if hook, ok := source.(contractsorm.ArrayPostMigrate); ok {
    if err := hook.PostMigrate(ctx, db, tableName); err != nil {
        return fmt.Errorf("post-migration hook failed for %s: %w", tableName, err)
    }
}

a.markPopulated(db, tableName)
return nil
```

### 3. Cache Reference Path for External Sources

This is covered as part of enhancement #1 (Persistent Caching) via the `ArrayCacheReference` interface. It allows sources that pull data from external files to tie cache freshness to the external file's modification time rather than the Go source file.

> **Note**: For sources that read from CSV, JSON, JSONL, or other flat-file formats, consider using the [flat-file driver](flat-file-driver.md) instead — it handles file parsing natively and could benefit from the same caching features described here (see [Future Enhancements](#future-enhancements)).

#### Usage with External Data

```go
type RoleSource struct{}

func (s *RoleSource) TableName() string { return "roles" }
func (s *RoleSource) Rows() ([]map[string]any, error) {
    // Parse roles from an external JSON file at runtime
    // (Note: the flat-file driver handles this natively — this example
    // shows how caching works with a manual ArraySource implementation)
    data, _ := os.ReadFile("data/roles.json")
    var roles []map[string]any
    json.Unmarshal(data, &roles)
    return roles, nil
}

func (s *RoleSource) CachePath() string {
    return "/tmp/neat_cache_roles.sqlite"
}

// Cache is only rebuilt when roles.json changes
func (s *RoleSource) CacheReferencePath() string {
    return "data/roles.json"
}

func (s *RoleSource) ShouldCache() bool {
    return true
}
```

## Implementation Plan

### Phase 1: Post-Migration Hook (Low Risk)

This is the simplest enhancement and has no impact on existing behavior since it's purely additive.

1. Add `ArrayPostMigrate` interface to `contracts/database/orm/array_source.go`
2. Add hook call in `Array.Populate()` after `insertRows`, before `markPopulated`
3. Add unit tests:
   - Source with `PostMigrate` that creates an index → verify index exists
   - Source with `PostMigrate` that returns an error → verify Populate fails
   - Source without `PostMigrate` → verify no change in behavior (regression)
4. Update example in `examples/array-driver/main.go` to demonstrate `PostMigrate`

### Phase 2: Persistent Caching (Medium Risk)

1. Add `ArrayCache`, `ArrayCacheReference`, `ArrayCacheEnabled` interfaces to `contracts/database/orm/array_source.go`
2. Implement cache logic in `Array.Populate()`:
   - Check cache file existence and freshness
   - Attach cache file via `ATTACH DATABASE` if fresh
   - Persist in-memory DB to cache file after population using `VACUUM INTO`
3. Add unit tests:
   - Source with caching enabled → verify cache file is created
   - Second run with fresh cache → verify population is skipped
   - Stale cache (reference file newer) → verify table is rebuilt
   - Cache disabled → verify current behavior (no cache file)
   - Cache file deleted → verify table is rebuilt
   - Cache file corrupt → verify graceful fallback to re-population
4. Add integration test via `database.Query().Model()`
5. Update `Cleanup()` to optionally remove cache files

### Phase 3: Documentation

1. Update `examples/array-driver/README.md` with caching and post-migration examples
2. Add caching section to driver documentation
3. Document cache invalidation rules and edge cases

## Design Decisions

### Why Opt-In Caching?

Caching is disabled by default (current behavior) to avoid surprising users with files on disk. Sources must explicitly implement `ArrayCacheEnabled` and return `true` from `ShouldCache()` to enable it. This keeps the simple case simple (in-memory, ephemeral) while allowing power users to opt into persistence.

### Why Not Cache at the Driver Level?

Sushi caches at the driver/model level using a fixed directory. neat's approach puts cache control in the `ArraySource` implementation because:
- The source knows where the data comes from and can best determine the cache key
- The source knows which external file to reference for staleness checks
- Different sources in the same application can have different caching strategies
- It avoids global state (cache directory configuration) in the driver

### Why VACUUM INTO for Cache Persistence?

SQLite's `VACUUM INTO` command creates a clean copy of the database in a single operation. It's simpler and more reliable than the backup API for this use case. Alternative: use `sqlite3_backup` via the Go SQLite driver's backup mechanism if `VACUUM INTO` is not supported by the embedded driver.

### Why Attach Instead of Open?

When a cache file is fresh, the driver uses `ATTACH DATABASE 'path' AS cache` and then runs queries against `cache.table_name` instead of `main.table_name`. This avoids closing and reopening the connection, which would break any active transactions or prepared statements. Alternatively, the driver could open the cache file directly as the primary database — this is simpler but requires the `*sql.DB` to be opened with the cache file path from the start.

**Recommended approach**: Open the cache file directly as the primary database when a fresh cache exists. This is simpler and avoids `ATTACH` complexity. The `Populate` method would:
1. Close the current in-memory connection
2. Open the cache file as the new `*sql.DB`
3. Replace the connection in the ORM

However, this requires changes to how `*sql.DB` is managed, which is currently owned by the ORM. A simpler alternative: always use in-memory, but after population, copy to cache file. On next start, check cache freshness first — if fresh, read the cache file into memory (open it as `:memory:` with `mode=memory` or copy rows). This avoids connection swapping but may be slower for large datasets.

**Decision**: Start with the simplest approach — always populate in-memory, persist to cache file after population, and on next run, if cache is fresh, load the cache file into the in-memory database. Optimize later if profiling shows this is a bottleneck.

### Post-Migration Hook: Why ctx and db Parameters?

The hook receives `context.Context` and `*sql.DB` so it can:
- Execute SQL within the same transaction context
- Respect cancellation/deadlines
- Access the raw connection for advanced operations

The `tableName` parameter is passed because the source may not know its own table name at the SQL level (it could be dynamically generated in the future).

## Risks and Mitigations

### Risk 1: Cache File Corruption
- **Issue**: A cache file could be corrupted by a crash during write, manual editing, or disk errors.
- **Mitigation**: The driver wraps cache loading in a try/catch — if the cache file fails to open or query, it falls back to full re-population. A warning is logged.

### Risk 2: Cache File Permissions
- **Issue**: The cache path may not be writable or readable.
- **Mitigation**: The driver checks file permissions before attempting to use the cache. If the path is not writable, caching is silently disabled with a warning log.

### Risk 3: Concurrent Cache Access
- **Issue**: Multiple processes could try to read/write the same cache file simultaneously.
- **Mitigation**: Use file locking (SQLite's built-in WAL mode or `flock`) for the cache file. For the initial implementation, document that cache files are per-process and should use unique paths in multi-process scenarios.

### Risk 4: Post-Migration Hook Misuse
- **Issue**: A user's `PostMigrate` hook could drop the table or corrupt data.
- **Mitigation**: The hook runs after population, so data is already inserted. If the hook fails, `Populate` returns an error and the table is not marked as populated (so it will be retried on the next call). Document that hooks should be idempotent (use `IF NOT EXISTS`).

## Future Enhancements

1. **Cache compression**: Support `.sqlite.gz` cache files for large datasets
2. **Cache versioning**: Include a schema version in the cache file to auto-invalidate on driver updates
3. **Shared cache directory**: A driver-level option for a default cache directory, so sources don't need to specify full paths
4. **Cache statistics**: Expose cache hit/miss metrics for observability
5. **Pre-migration hook**: A hook that runs before table creation, allowing custom schema modifications (e.g., adding virtual columns or generated columns)
6. **Apply caching and post-migration to flat-file driver**: The persistent caching and post-migration hook patterns proposed here are equally applicable to the [flat-file driver](flat-file-driver.md). Both drivers share the same SQLite-backed architecture, so caching (`CachePath`, `CacheReferencePath`, `ShouldCache`) and `PostMigrate` could be extracted into shared interfaces in `contracts/database/orm/` and implemented by both drivers.

## References

- Sushi (Laravel array driver): https://github.com/calebporzio/sushi
- Array driver implementation: `database/driver/array.go`
- Array source contracts: `contracts/database/orm/array_source.go`
- Flat-file driver proposal: `docs/proposals/flat-file-driver.md`
- SQLite VACUUM INTO: https://www.sqlite.org/lang_vacuum.html
- SQLite ATTACH DATABASE: https://www.sqlite.org/lang_attach.html
