# Proposal: Azure Table Storage Support Investigation

**Date**: August 8, 2026 (original) · September 5, 2026 (revised)
**Status**: Revised — core premise disproven by `aztablessql`
**Priority**: Low

> **Revision note (2026-09-05):** The original version of this proposal argued that Azure Table Storage **cannot** be used through `database/sql`, and that supporting it would require a "second interface alongside `Driver`" for non-SQL backends — a fundamental architectural change. That premise is now disproven. The [`aztablessql`](https://github.com/dracory/aztablessql) module is a working `database/sql` driver for Azure Table Storage: it registers `"aztables"` via `sql.Register`, implements `database/sql/driver.Driver`, and translates a SQL subset into Table Storage REST/OData calls. This means Neat can support Azure Table Storage as **just another driver returning a `*sql.DB`** — exactly like its existing MySQL, PostgreSQL, and SQLite drivers — with zero architectural change. The options and recommendation below have been rewritten accordingly. The analysis of *other* API-backed NoSQL stores (Cosmos DB native API, DynamoDB, MongoDB, Cassandra, Redis) remains valid, because none of those have `database/sql` drivers.

## Problem

The comparison tables do not currently list Azure Table Storage. The question is whether Neat **should** support it, and if so, what that looks like.

Azure Table Storage is a NoSQL key-value store with limited query capabilities. While it does not use SQL natively, the [`aztablessql`](https://github.com/dracory/aztablessql) driver now provides a `database/sql`-compatible SQL surface over it — a subset of SQL (point reads, `AND` filters, single-entity mutations, `LIMIT`, `COUNT(*)`, DDL, batch/entity-group transactions) mapped onto the operations Table Storage actually supports. This changes the integration picture fundamentally: Neat's existing `Driver` interface (which returns `*sql.DB`) works as-is.

## Background

### What Azure Table Storage Is

Azure Table Storage is a NoSQL data store that:
- Stores structured but **schemaless** data in tables
- Uses a **partition key + row key** composite primary key (no auto-increment, no sequential IDs)
- Is accessed via **HTTP REST API** with OData query options
- Scales horizontally by partition
- Provides **eventual consistency** by default
- Has **no joins**, no foreign keys, no transactions (beyond entity group transactions within a single partition)
- Supports **limited querying** via OData `$filter`, `$top`, and `$select`

### Query Capabilities

Azure Table Storage supports a subset of query operations via OData filter syntax. From the [official documentation](https://learn.microsoft.com/en-us/rest/api/storageservices/writing-linq-queries-against-the-table-service):

**Supported:**
- `Where` — filtering on string, numeric, boolean, and datetime properties
- `Take` — limit results (max 1000 per request)
- `Select` — project a subset of properties
- Comparison operators: `==`, `>`, `<`, `<=`, `>=`, `.Equals()`, `.CompareTo()`
- Logical operators: `&&` (AND), `||` (OR)
- Prefix matching via `CompareTo()` (e.g., `LastName.CompareTo("A") >= 0 && LastName.CompareTo("B") < 0`)

**Not supported:**
- Joins
- GROUP BY / aggregates (COUNT, SUM, AVG — must be computed client-side)
- ORDER BY (results are sorted only by PartitionKey + RowKey)
- OFFSET (pagination via continuation tokens only)
- Subqueries
- UNION

### Azure Table Storage vs SQL Databases

| Feature | SQL Databases | Azure Table Storage |
|---------|---------------|---------------------|
| Query language | SQL | OData filter syntax (limited) — but `aztablessql` exposes a SQL subset over it |
| Protocol | TCP (MySQL/PostgreSQL wire) | HTTP REST |
| `database/sql` driver? | ✅ (native) | ✅ (`aztablessql`) |
| Joins | ✅ | ❌ |
| Foreign keys | ✅ | ❌ |
| Transactions | ✅ (ACID) | ⚠️ (entity group transactions within partition only — exposed via `aztablessql` batch API) |
| Aggregates | ✅ (COUNT, SUM, AVG) | ⚠️ `COUNT(*)` only (client-side, O(n)); no SUM/AVG |
| GROUP BY | ✅ | ❌ |
| ORDER BY | ✅ | ❌ (only by PartitionKey + RowKey) |
| UPDATE | ✅ | ✅ (merge/replace entity, point only) |
| DELETE | ✅ | ✅ (delete entity, point only) |
| Auto-increment | ✅ | ❌ (must generate keys client-side) |
| Schema | Fixed | Schemaless (entities can have any properties) |
| Indexes | ✅ (B-tree, GIN, etc.) | ⚠️ (only PartitionKey + RowKey are indexed) |
| Data types | Rich (JSON, arrays, etc.) | Limited (string, int32, int64, double, bool, datetime, GUID, binary) |

### Go SDK and the `aztablessql` driver

The Azure SDK for Go provides `github.com/Azure/azure-sdk-for-go/sdk/data/aztables`, an HTTP REST client with its own types (`ServiceClient`, `Client`) that does **not** implement the `database/sql` driver interface.

The [`aztablessql`](https://github.com/dracory/aztablessql) module bridges that gap. It is a `database/sql` driver (`sql.Register("aztables", ...)`) that:

- Holds an `aztables.ServiceClient` (an HTTP connection to Azure) behind a standard `driver.Conn`.
- Parses a SQL subset (`INSERT`, `INSERT OR REPLACE` / `UPSERT INTO`, `INSERT OR MERGE`, `SELECT`, `UPDATE`, `DELETE`, `CREATE TABLE`, `DROP TABLE`, `SHOW TABLES`, `SELECT COUNT(*)`) via a regex/state-machine parser.
- Translates `WHERE` clauses to OData `$filter` strings, `LIMIT` to `$top`, and `SELECT col1, col2` to `$select`.
- Issues HTTP REST calls for INSERT/UPDATE/DELETE instead of SQL statements.
- Handles pagination via continuation tokens (not OFFSET).
- Exposes batch/entity-group transactions (atomic, single-partition, up to 100 ops) through `database/sql`'s `Conn.Raw` escape hatch.
- Surfaces ETag-based optimistic concurrency via an optional `AND ETag = ?` condition on `UPDATE`/`DELETE`.
- Returns clear errors for unsupported operations (JOINs, GROUP BY, OR, LIKE, IS NULL, IN, subqueries, `BEGIN`/`COMMIT`/`ROLLBACK`).

This is the critical fact that revises the original proposal: **Azure Table Storage can be used through `database/sql`**, because `aztablessql` implements the driver interface and translates a SQL subset to OData/REST. There is no need for a "second interface alongside `Driver`."

## Feasibility Assessment

### Why This Is No Longer Fundamentally Different

The original proposal argued that every Neat driver ultimately hands back a `*sql.DB` and executes SQL through `database/sql`, and that Azure Table Storage could not do this. The first half is still true; the second half is no longer true. `aztablessql` hands back a `*sql.DB` (via `sql.Open("aztables", connStr)`) and executes a SQL subset through `database/sql` — exactly the pattern every other Neat driver uses. The SQL is translated to OData/REST internally, but that translation is invisible to the caller and to Neat.

Neat's existing `Driver` interface works unchanged:

```go
type Driver interface {
    Open(dsn string) (*sql.DB, error)
    Close(db *sql.DB) error
    Ping(ctx context.Context, db *sql.DB) error
    BeginTx(ctx context.Context, db *sql.DB, opts *sql.TxOptions) (*sql.Tx, error)
    Placeholder(n int) string
    Dialect() string
}
```

`aztablessql.Driver.Open` returns a `*sql.DB`. `BeginTx` would fail (Table Storage has no cross-partition transactions), but that is a runtime constraint, not an interface incompatibility — the same way SQLite has limitations relative to Postgres. Neat can register `aztablessql` as a driver, return a `*sql.DB`, and generate SQL against it. The only requirement is that the generated SQL stays within `aztablessql`'s supported subset.

### What the CSV/JSON/XML Driver Pattern Has to Do With This

The original proposal used the `csvdb`/`jsondb`/`xmldb`/`godb`/`array` drivers — which reduce their source to an in-memory SQLite database at `Open()` — to argue that Azure Table Storage could not follow the same pattern because it is live, unbounded, remote, and write-back required. That argument is correct but now irrelevant: `aztablessql` does **not** follow the load-into-SQLite pattern. It is a genuine `database/sql` driver that proxies each call to the live Azure Table Storage REST API, the same way the MySQL driver proxies each call to the MySQL wire protocol. There is no snapshot, no in-memory copy, and writes go straight to Azure.

### What Supporting Azure Table Storage Now Requires

Because `aztablessql` already exists, supporting it in Neat is no longer a 30-40 hour adapter project. It is:

1. **Add `github.com/dracory/aztablessql` as a dependency** (or document it as an optional driver users can import themselves).
2. **Register `"aztables"` as a driver type** in Neat's driver registry, mapping it to `aztablessql` and returning a `*sql.DB` from `sql.Open("aztables", connStr)`.
3. **Set `Dialect()` appropriately** (e.g. `"aztables"`) so Neat's SQL generation can stay within the supported subset.
4. **Document the constraints** — which Neat features work and which don't against this driver (see the matrix below). Neat's query builder must avoid generating JOINs, GROUP BY, OR, ORDER BY (beyond key columns), OFFSET, subqueries, and `BEGIN`/`COMMIT`/`ROLLBACK` when targeting `aztables`.
5. **Handle the composite key model** — Neat's primary key model assumes a single auto-incrementing ID; Azure Table Storage requires `PartitionKey` + `RowKey` chosen by the application. This is a modeling concern, not an architectural one, but it affects how Neat maps models to entities.

Steps 1-3 are trivial. Step 4 is documentation plus possibly query-builder guards that reject unsupported clauses when the dialect is `aztables`. Step 5 is the only genuinely new modeling work, and it is the same work any Table Storage user has to do regardless of ORM.

### Feature Compatibility Matrix

| Neat Feature | Azure Table Storage (via `aztablessql`) Feasibility |
|--------------|-------------------------------|
| Query Builder (SELECT) | ✅ Works (single-table, `AND` filters, projection, `LIMIT`; point read via `PartitionKey`+`RowKey`) |
| INSERT | ✅ Works (`INSERT`, `INSERT OR REPLACE`, `INSERT OR MERGE`, `UPSERT INTO`) |
| UPDATE | ✅ Works (merge/replace, point only — `WHERE PartitionKey = ? AND RowKey = ?`) |
| DELETE | ✅ Works (point only — `WHERE PartitionKey = ? AND RowKey = ?`) |
| WHERE clauses | ⚠️ Limited (`=`, `!=`, `<>`, `>`, `>=`, `<`, `<=`; `AND` only, no `OR`, no `LIKE`, no `IS NULL`, no `IN`) |
| ORDER BY | ❌ Not supported by Table Storage (only PartitionKey + RowKey ordering) |
| LIMIT / OFFSET | ⚠️ `LIMIT` works (`$top`); `OFFSET` not supported (continuation-token pagination only) |
| GROUP BY / HAVING | ❌ Not supported |
| Aggregates | ⚠️ `COUNT(*)` works (client-side, O(n)); no `SUM`/`AVG`/`MIN`/`MAX` |
| JOINs | ❌ Not supported |
| Associations (BelongsTo, HasMany) | ❌ Not supported at DB level (no joins) |
| Transactions | ⚠️ No `BEGIN`/`COMMIT`/`ROLLBACK`; batch/entity-group transactions available via `Conn.Raw` (single partition, ≤100 ops) |
| Savepoints | ❌ Not supported |
| Soft Deletes | ✅ Works (filter on a deleted flag property) |
| Migrations | ❌ Not applicable (schemaless); `CREATE TABLE`/`DROP TABLE` work but take no column definitions |
| Schema Builder | ❌ Not applicable (schemaless) |
| Views | ❌ Not supported |
| Array-Backed Sources | ❌ Not applicable |
| Factories / Seeders | ✅ Works (insert entities) |
| Observers | ✅ Works (application-level) |
| Polymorphic Associations | ❌ Not supported (no joins) |
| ToSql Interface | ✅ Works (the driver consumes SQL) |
| Raw SQL | ⚠️ Works, but must stay within `aztablessql`'s supported subset |
| Pessimistic Locking | ❌ Not supported |
| Query Scopes | ✅ Works (application-level) |
| Select (projection) | ✅ Works (`$select`) |
| Chunk processing | ✅ Works (via continuation-token pagination) |
| Optimistic concurrency | ✅ Works (ETag via `AND ETag = ?` on UPDATE/DELETE) |

**Result**: Roughly the same ~40% of Neat's features work as the original proposal estimated — but the critical difference is that this is now achievable **with zero architectural change**, because `aztablessql` provides the `database/sql` driver that the original proposal claimed could not exist.

## Options

### Option A: Full NoSQL Backend Abstraction (Not Recommended — and Now Unnecessary for Azure Table Storage)

Redesign Neat to support both SQL (`database/sql`) and NoSQL (HTTP API) backends behind a common interface.

**Effort**: 100+ hours (essentially building a second ORM)
**Risk**: Very high — would complicate the codebase, slow down SQL feature development, and create a maintenance burden
**Verdict**: Not recommended. For Azure Table Storage specifically, this is now unnecessary because `aztablessql` already provides the `database/sql` driver. This option remains relevant only if Neat wants to support API-backed stores that have **no** `database/sql` driver (see [Other NoSQL Stores](#what-about-other-nosql--api-backed-stores) below).

### Option B: Use `aztablessql` as a `database/sql` Driver (Recommended — Low Effort)

Register `github.com/dracory/aztablessql` as a Neat driver type. Neat's existing `Driver` interface, query builder, and `*sql.DB`-based execution path all work unchanged. The driver translates the supported SQL subset to Azure Table Storage REST/OData calls internally.

**How it would work**:
1. Add `github.com/dracory/aztablessql` as a dependency (or document it as an optional user-imported driver).
2. Register `"aztables"` as a driver type in Neat, returning `sql.Open("aztables", connStr)`.
3. Set `Dialect()` to `"aztables"` so Neat knows to constrain SQL generation.
4. Optionally add query-builder guards that reject unsupported clauses (JOINs, GROUP BY, OR, ORDER BY beyond keys, OFFSET, subqueries, `BEGIN`/`COMMIT`/`ROLLBACK`) with clear errors when the dialect is `aztables`.
5. Document the composite `PartitionKey` + `RowKey` key model and which Neat features are available.

**Supported features** (via `aztablessql`):
- SELECT with WHERE filtering (`=`, `!=`, `<>`, `>`, `>=`, `<`, `<=`; `AND` only) on string, numeric, boolean, datetime properties
- Point reads (`WHERE PartitionKey = ? AND RowKey = ?` → `GetEntity`)
- Projection (`SELECT col1, col2`)
- Limit (`LIMIT n` → `$top`)
- INSERT / INSERT OR REPLACE / INSERT OR MERGE / UPSERT INTO
- UPDATE (merge/replace, point only)
- DELETE (point only)
- COUNT(*) (client-side, O(n))
- CREATE TABLE / DROP TABLE / SHOW TABLES
- Chunk processing (via continuation tokens)
- Soft deletes (filter on a flag property)
- Optimistic concurrency (ETag via `AND ETag = ?`)
- Batch/entity-group transactions (via `Conn.Raw`, single partition, ≤100 ops)

**Explicitly unsupported** (rejected by `aztablessql` with clear errors, or by Table Storage itself):
- JOINs
- GROUP BY / HAVING
- SUM / AVG / MIN / MAX aggregates (compute client-side)
- ORDER BY (beyond PartitionKey + RowKey)
- OFFSET (use continuation tokens)
- OR, LIKE, IS NULL, IN
- Subqueries
- UNION
- Cross-partition transactions (`BEGIN`/`COMMIT`/`ROLLBACK`)
- Savepoints
- Foreign keys
- Associations (at DB level)
- Auto-increment IDs
- Schema builder / column definitions in DDL
- Migrations (schemaless)
- Views
- Pessimistic locking

**Effort**: 4-8 hours (driver registration + dialect wiring + documentation; optionally query-builder guards)
**Risk**: Low — no architectural change; `aztablessql` is already built and tested (unit + integration tests against Azurite, CI on GitHub Actions)
**Verdict**: Recommended. This is now a small, low-risk integration rather than the 30-40 hour adapter project the original proposal described.

### Option C: Array Source Workaround (Zero Effort)

Neat already has an "array-backed source" feature (`array` driver) that allows querying in-memory data. Azure Table Storage entities could be loaded into memory and queried via the array source.

**How it would work**:
1. User fetches entities from Azure Table Storage using the Azure SDK (or `aztablessql`) directly
2. User converts entities to Go structs
3. User wraps the slice in `neat.NewArraySourceFrom(items)`
4. User queries via Neat's query builder (in-memory, no SQL)

**Effort**: 0 hours (already works via array source)
**Risk**: None
**Verdict**: This already works but provides no real Azure Table Storage integration — it's just in-memory querying after manual data loading. Option B is strictly better now that it is low-effort.

### Option D: No Support (Also Valid)

Do not support Azure Table Storage. Document that Neat is a SQL ORM and that while a `database/sql` driver exists for Azure Table Storage (`aztablessql`), Neat chooses not to register it as a first-class driver.

**Effort**: Zero (or 1-2 hours of documentation noting `aztablessql` exists for users who want to wire it up themselves)
**Risk**: None
**Verdict**: Valid if there is no user demand. Users can still import `aztablessql` directly and pass the resulting `*sql.DB` to Neat manually, since it is a standard `database/sql` driver.

## Recommendation: Option B (Use `aztablessql`)

The original recommendation was to defer until user demand justified a 30-40 hour OData adapter. That calculus has changed: `aztablessql` already exists, is tested, and is a standard `database/sql` driver. Supporting it in Neat is now a 4-8 hour task of driver registration, dialect wiring, and documentation — not an adapter project.

### Why Option B Now

- **No architectural change.** `aztablessql` returns a `*sql.DB` and implements `database/sql/driver.Driver`. Neat's `Driver` interface, query builder, and execution path work unchanged.
- **Low effort.** Driver registration + dialect + documentation. Optionally, query-builder guards that reject unsupported clauses when the dialect is `aztables`.
- **Real query capabilities.** The SQL subset `aztablessql` supports (point reads, `AND` filters with all comparison operators, projection, `LIMIT`, `COUNT(*)`, CRUD, DDL, batch transactions, ETag concurrency) covers genuine use cases.
- **Already tested.** `aztablessql` ships unit tests plus integration tests against Azurite (the official Azure Storage emulator), with CI on GitHub Actions.

### What Neat Would Still Need to Handle

The only non-trivial work is **modeling**, not architecture:

1. **Composite key model.** Neat's primary key model assumes a single auto-incrementing ID. Azure Table Storage requires `PartitionKey` + `RowKey` chosen by the application. Neat would need to let a model declare a composite key (or treat `PartitionKey`/`RowKey` as two required columns) when the dialect is `aztables`.
2. **Query-builder constraints.** When the dialect is `aztables`, Neat's query builder should refuse to generate (or the driver should reject) JOINs, GROUP BY, OR, ORDER BY beyond key columns, OFFSET, subqueries, UNION, and `BEGIN`/`COMMIT`/`ROLLBACK`. `aztablessql` already rejects these at the driver level with clear errors, but Neat-side guards would give earlier, more actionable feedback.
3. **No schema builder / migrations.** Table Storage is schemaless; `CREATE TABLE` takes no column definitions. Neat's schema builder and migration system simply don't apply to this dialect and should be documented as unsupported.

### Why Not Option A (Full Abstraction)

Option A (redesigning Neat to support both SQL and NoSQL backends behind a common non-`database/sql` interface) is not recommended regardless of demand. For Azure Table Storage it is now unnecessary — `aztablessql` already provides the `database/sql` driver. Option A remains relevant only for API-backed stores that have no `database/sql` driver at all (see below).

### If There Is No User Demand → Option D

If no users are asking for Azure Table Storage support, even the 4-8 hours of integration work may not be justified. In that case, document that `aztablessql` exists as a standalone `database/sql` driver that users can import and pass to Neat manually via `sql.Open("aztables", connStr)`. Since it is a standard driver, no Neat-side changes are required for a user to use it today — they just won't get dialect-aware query-builder guards or first-class documentation.

## What About Other NoSQL / API-Backed Stores?

The original proposal's architectural argument — that API-backed stores without a `database/sql` driver require a second interface alongside `Driver` — **remains valid for every store in the table below except Azure Table Storage**. `aztablessql` is the exception: it proves that a sufficiently simple API-backed store *can* be wrapped in a `database/sql` driver by translating a SQL subset to the store's native query language. The same approach could in principle be attempted for other stores, but none of the following have a production-grade `database/sql` driver today.

| Store | Protocol | Query Language | `database/sql` Driver? | Same Constraint? |
|-------|----------|----------------|------------------------|-------------------|
| Azure Table Storage | HTTP REST | OData `$filter` | ✅ Yes (`aztablessql`) | **No** — constraint lifted |
| Azure Cosmos DB (native API) | HTTP REST | SQL API (Cosmos-specific) | No | Yes |
| AWS DynamoDB | HTTP API | Key-condition + filter expressions | No | Yes |
| MongoDB | TCP (wire protocol) | MQL (MongoDB Query Language) | No | Yes |
| Apache Cassandra | TCP (native) | CQL (SQL-like but not `database/sql`) | No (community drivers exist but are partial) | Yes |
| Redis | TCP (RESP) | Redis commands | No | Yes |

Each of these (except Azure Table Storage) would require either a `database/sql` driver to be built for it (the `aztablessql` approach, generalized), or the original Option A — a second interface alongside `Driver` for API-backed backends, with a code path in the ORM that executes against an API connection rather than a `*sql.DB`. The query translation would differ per store (SQL API for Cosmos DB, filter expressions for DynamoDB, MQL for MongoDB, CQL for Cassandra), but the architectural gap is identical for all of them.

### Could the `aztablessql` Approach Generalize?

`aztablessql` works because Azure Table Storage's query model is simple enough to express as a small SQL subset: single-table filters, projection, limiting, and point CRUD. Stores with richer but still non-SQL query languages (Cosmos DB SQL API, MongoDB's MQL, Cassandra's CQL) are harder to map onto a SQL subset, but not impossible — CQL in particular is already SQL-like. The limiting factor is engineering effort and whether the resulting SQL subset is useful enough to justify the driver. For Azure Table Storage, that bar has been cleared. For the others, it has not.

#### Azure Cosmos DB Specifically

Azure Cosmos DB has multiple API surfaces:

- **Cosmos DB SQL API** — despite the name "SQL", this is Cosmos DB's own query language over JSON documents, accessed via HTTP REST. It is not SQL in the `database/sql` sense and has no `database/sql` driver. Same constraint as the other stores in the table above.
- **Cosmos DB MongoDB API** — wire-compatible with MongoDB, but MongoDB has no `database/sql` driver. Same constraint.
- **Cosmos DB Cassandra API** — wire-compatible with Cassandra. CQL resembles SQL, but there is no production-grade `database/sql` driver for Cassandra/Cosmos Cassandra API. Same constraint.
- **Cosmos DB for PostgreSQL** (formerly Hyperscale/Citus) — this **is** PostgreSQL-compatible and works with Neat's existing PostgreSQL driver. No new driver needed. But this is just PostgreSQL, not Cosmos DB's native API.
- **Cosmos DB for Table** — wire-compatible with Azure Table Storage. This means `aztablessql` can also target Cosmos DB for Table by pointing the connection string at the Cosmos DB Table endpoint, the same way it targets Azure Table Storage. This is a notable side benefit of the `aztablessql` approach: it covers both Azure Table Storage and the Cosmos DB for Table API with one driver.

In short: only the PostgreSQL API of Cosmos DB works with Neat today (via the existing PostgreSQL driver), and the Cosmos DB for Table API works via `aztablessql`. Every other Cosmos DB API still hits the same wall as the other stores in the table.

#### Implications for the Proposal

If Neat ever decides to support the remaining API-backed NoSQL stores, the work is either "build a `database/sql` driver per store" (the `aztablessql` approach, generalized) or "build the non-SQL backend abstraction once" (Option A). The `aztablessql` precedent suggests the driver approach is feasible for stores with simple enough query models, and preferable to Option A where it works, because it requires no change to Neat's architecture. Option A remains the fallback for stores whose query model cannot be reasonably expressed as a SQL subset.

## Documentation Update

If we go with Option B, we should:

1. **Add `aztablessql` to the driver list** in `docs/driver-registration.html` as a supported (or documented) driver type, with a link to the `aztablessql` repository.
2. **Add Azure Table Storage to the comparison tables** with a clear note that it is a NoSQL store accessed via a `database/sql` driver (`aztablessql`), and that only a subset of Neat's features apply (per the matrix above).
3. **Document the composite key model** (`PartitionKey` + `RowKey`) and which Neat features are unavailable against this dialect.
4. **Document the array source workaround** in `docs/array-source.html` as an alternative for users who want in-memory querying of NoSQL data without the driver.

If we go with Option D, we should:

1. **Add a note to `docs/driver-registration.html`** explaining that while Neat is a SQL ORM, a `database/sql` driver for Azure Table Storage (`aztablessql`) exists and can be used with Neat by passing `sql.Open("aztables", connStr)` to Neat manually.
2. **Do not add Azure Table Storage to comparison tables** as a first-class driver, but mention `aztablessql` as a user-wireable option.
3. **Document the array source workaround** in `docs/array-source.html`.

## Estimated Effort

- **Option B (Use `aztablessql`)**: 4-8 hours (driver registration + dialect wiring + documentation; optionally query-builder guards for unsupported clauses)
- **Option C (Array Source Workaround)**: 0 hours (already works)
- **Option A (Full Abstraction)**: 100+ hours (not recommended; unnecessary for Azure Table Storage now that `aztablessql` exists)
- **Option D (No Support)**: 1-2 hours (documentation only, noting `aztablessql` exists)

## Success Criteria (Option B)

1. `aztablessql` is registered as a Neat driver type (`"aztables"`) and `sql.Open("aztables", connStr)` returns a working `*sql.DB` through Neat.
2. `docs/driver-registration.html` lists Azure Table Storage (via `aztablessql`) with a link to the driver and a note about the supported feature subset.
3. The comparison tables include Azure Table Storage with a clear "NoSQL via `database/sql` driver" annotation.
4. The composite `PartitionKey` + `RowKey` key model is documented.
5. Unsupported features (JOINs, GROUP BY, OR, ORDER BY, OFFSET, transactions, associations, migrations, schema builder) are documented as unavailable for this dialect.
6. Optionally: query-builder guards reject unsupported clauses with clear errors when the dialect is `aztables`.

## Decision Required

**Should Neat register `aztablessql` as a first-class driver (Option B), or document it as a user-wireable option without first-class support (Option D)?**

Arguments for Option B:
- `aztablessql` is a standard `database/sql` driver — no architectural change to Neat.
- Low effort (4-8 hours) for real value: CRUD, filtering, projection, limiting, batch transactions, optimistic concurrency.
- Would differentiate Neat from other Go ORMs (none register an Azure Table Storage driver).
- Useful for users building on Azure who want ORM-like abstraction over Table Storage.
- Covers both Azure Table Storage and Cosmos DB for Table with one driver.

Arguments for Option D:
- Even 4-8 hours of effort is not justified without user demand.
- Users can already use `aztablessql` with Neat manually via `sql.Open("aztables", connStr)` — first-class support is a convenience, not a necessity.
- ~60% of Neat's features don't apply (JOINs, aggregates, associations, migrations, etc.), so the integration is inherently partial.
- No SQL ORM competitor supports it (not a competitive gap today).

**Recommendation**: Implement Option B if there is any user demand — the effort is now low enough that the bar is much lower than the original proposal's 30-40 hour estimate. If there is genuinely no demand, Option D is fine, but document `aztablessql`'s existence so users know they can wire it up themselves.

## References

- `aztablessql` driver: https://github.com/dracory/aztablessql
- Azure Table Storage Go SDK: https://github.com/Azure/azure-sdk-for-go/tree/main/sdk/data/aztables
- Azure Table Storage Documentation: https://learn.microsoft.com/en-us/azure/storage/tables/
- Azure Cosmos DB for Table: https://learn.microsoft.com/en-us/azure/cosmos-db/table/
- Azurite (local emulator): https://learn.microsoft.com/en-us/azure/storage/common/storage-use-azurite
- Neat Array Source: `docs/array-source.html`
- Neat Driver Registration: `docs/driver-registration.html`
