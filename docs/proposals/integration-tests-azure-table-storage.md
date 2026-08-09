# Proposal: Azure Table Storage Support Investigation

**Date**: August 8, 2026
**Status**: Proposal
**Priority**: Low

## Problem

The comparison tables do not currently list Azure Table Storage. The question is whether Neat **should** support it, and if so, what that looks like.

Azure Table Storage is a NoSQL key-value store with limited query capabilities. While it does not use SQL, it supports a LINQ-style query syntax (mapped to OData `$filter`) that covers basic filtering, projection, and limiting — concepts that map to parts of Neat's query builder. However, it is accessed via HTTP REST API, not `database/sql`, which creates a fundamental integration challenge.

## Background

### What Azure Table Storage Is

Azure Table Storage is a NoSQL data store that:
- Stores structured but **schemaless** data in tables
- Uses a **partition key + row key** composite primary key (no auto-increment, no sequential IDs)
- Is accessed via **HTTP REST API** with OData query options
- Scales horizontally by partition
- Provides **eventual consistency** by default
- Has **no joins**, no foreign keys, no transactions (beyond entity group transactions within a single partition)
- Supports **limited querying** via OData `$filter`, `$top`, and `$select` (or LINQ syntax in .NET)

### Query Capabilities

Azure Table Storage supports a subset of query operations via OData filter syntax (or LINQ in .NET). From the [official documentation](https://learn.microsoft.com/en-us/rest/api/storageservices/writing-linq-queries-against-the-table-service):

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
| Query language | SQL | OData filter syntax (limited) |
| Protocol | TCP (MySQL/PostgreSQL wire) | HTTP REST |
| Joins | ✅ | ❌ |
| Foreign keys | ✅ | ❌ |
| Transactions | ✅ (ACID) | ⚠️ (entity group transactions within partition only) |
| Aggregates | ✅ (COUNT, SUM, AVG) | ❌ (must be computed client-side) |
| GROUP BY | ✅ | ❌ |
| ORDER BY | ✅ | ⚠️ (only by PartitionKey + RowKey) |
| UPDATE | ✅ | ✅ (replace/merge entity) |
| DELETE | ✅ | ✅ (delete entity) |
| Auto-increment | ✅ | ❌ (must generate keys client-side) |
| Schema | Fixed | Schemaless (entities can have any properties) |
| Indexes | ✅ (B-tree, GIN, etc.) | ⚠️ (only PartitionKey + RowKey are indexed) |
| Data types | Rich (JSON, arrays, etc.) | Limited (string, int32, int64, double, bool, datetime, GUID, binary) |

### Go SDK

The Azure SDK for Go provides `github.com/Azure/azure-sdk-for-go/sdk/data/aztables`:
- Uses HTTP REST API, not `database/sql`
- Has its own client types (`ServiceClient`, `Client`)
- Does **not** implement the `database/sql` driver interface
- Uses `aztables.EDMEntity` for entities, not Go structs with tags

This is the critical difference: **Azure Table Storage cannot be used through `database/sql`**. It has its own SDK with its own types, its own connection model, and its own query syntax.

## Feasibility Assessment

### Why This Is Fundamentally Different

Neat ORM is built on `database/sql`. Every driver Neat supports (MySQL, PostgreSQL, SQLite, SQL Server, Oracle, Turso) implements the `database/sql` driver interface. The query builder generates SQL. The schema builder generates DDL. The migration system runs SQL.

Azure Table Storage:
- Has no `database/sql` driver (and cannot have one — it's not SQL)
- Has no SQL to generate
- Has no DDL (tables are created via REST API, schemaless)
- Has no SQL migrations (schema is defined per-entity, not per-table)

### What Would Be Required

To support Azure Table Storage, Neat would need to:

1. **Abstract away `database/sql`**: The entire query builder, schema builder, and migration system would need a non-`database/sql` backend. This is a fundamental architectural change.

2. **Create a new abstraction layer**: A backend interface that can be implemented by both `database/sql` drivers (SQL databases) and HTTP-based APIs (Azure Table Storage, Cosmos DB, DynamoDB, etc.).

3. **Translate query builder operations to OData filters**: Neat's `Where("age > ?", 18)` would need to become an OData filter like `age gt 18`.

4. **Handle schemaless entities**: Neat's model-to-table mapping assumes a fixed schema. Azure Table Storage entities can have any properties.

5. **Handle partition key + row key**: Neat's primary key model assumes a single auto-incrementing ID. Azure Table Storage requires a composite key chosen by the application.

6. **Remove SQL-specific features**: Joins, GROUP BY, HAVING, aggregates, subqueries, UNION — none of these exist in Azure Table Storage.

### Feature Compatibility Matrix

| Neat Feature | Azure Table Storage Feasibility |
|--------------|-------------------------------|
| Query Builder (SELECT) | ⚠️ Limited (single-table, OData filter, no JOINs) |
| INSERT | ✅ Possible (insert entity) |
| UPDATE | ✅ Possible (replace/merge entity) |
| DELETE | ✅ Possible (delete entity) |
| WHERE clauses | ⚠️ Limited (OData filters, max 15 comparisons, no complex conditions) |
| ORDER BY | ❌ Only by PartitionKey + RowKey |
| LIMIT / OFFSET | ⚠️ Limited ($top max 1000, no offset — pagination via continuation tokens) |
| GROUP BY / HAVING | ❌ Not supported |
| Aggregates (COUNT, SUM, AVG) | ❌ Not supported (client-side only) |
| JOINs | ❌ Not supported |
| Associations (BelongsTo, HasMany) | ❌ Not supported (no joins) |
| Transactions | ⚠️ Entity group transactions only (same partition) |
| Savepoints | ❌ Not supported |
| Soft Deletes | ⚠️ Possible (filter on deleted flag) |
| Migrations | ❌ Not applicable (schemaless) |
| Schema Builder | ❌ Not applicable (schemaless) |
| Views | ❌ Not supported |
| Array-Backed Sources | ❌ Not applicable |
| Factories / Seeders | ⚠️ Possible (insert entities) |
| Observers | ✅ Possible (application-level) |
| Polymorphic Associations | ❌ Not supported |
| ToSql Interface | ❌ Not applicable (no SQL) |
| Raw SQL | ❌ Not applicable (no SQL) |
| Pessimistic Locking | ❌ Not supported |
| Query Scopes | ✅ Possible (application-level) |
| Select (projection) | ✅ Possible ($select) |
| Chunk processing | ⚠️ Possible (via continuation tokens) |

**Result**: ~40% of Neat's features could work with an adapter; ~60% would not work or would require fundamental rearchitecting.

## Options

### Option A: Full NoSQL Backend Abstraction (Not Recommended)

Redesign Neat to support both SQL (`database/sql`) and NoSQL (HTTP API) backends behind a common interface.

**Effort**: 100+ hours (this is essentially building a second ORM)
**Risk**: Very high — would complicate the codebase, slow down SQL feature development, and create a maintenance burden
**Verdict**: Not recommended. This would be a different project entirely.

### Option B: OData Query Adapter (Possible — Medium Effort)

Build a dedicated Azure Table Storage adapter that translates Neat's query builder operations to OData filter syntax. This is more feasible than it first appears because the query concepts map well:

| Neat Query Builder | OData / Azure Tables |
|--------------------|---------------------|
| `Where("age > ?", 30)` | `$filter=age gt 30` |
| `Where("name = ?", "Smith")` | `$filter=name eq 'Smith'` |
| `Where("active = ?", true)` | `$filter=active eq true` |
| `Limit(10)` / `Take(10)` | `$top=10` |
| `Select("name", "email")` | `$select=name,email` |
| `Where("age > ? AND age < ?", 18, 65)` | `$filter=age gt 18 and age lt 65` |

**How it would work**:
1. Add `github.com/Azure/azure-sdk-for-go/sdk/data/aztables` as a dependency
2. Register `"azure-table"` as a new driver type in Neat (not via `database/sql`)
3. Create an adapter that implements Neat's query interface but translates to Azure Table Storage REST calls
4. Map `Where()` calls to OData `$filter` strings
5. Map `Limit()` / `Take()` to `$top`
6. Map `Select()` to `$select`
7. Handle pagination via continuation tokens (not OFFSET)
8. Return clear errors for unsupported operations (JOINs, GROUP BY, aggregates, ORDER BY beyond PartitionKey+RowKey)

**Supported features**:
- SELECT with WHERE filtering (string, numeric, boolean, datetime)
- Projection (Select specific properties)
- Limit (Take, max 1000 per page)
- Pagination (via continuation tokens)
- INSERT (insert entity)
- UPDATE (replace/merge entity)
- DELETE (delete entity)
- Chunk processing (via continuation tokens)
- Soft deletes (filter on deleted flag)

**Explicitly unsupported** (with clear error messages):
- JOINs
- GROUP BY / HAVING
- Aggregates (COUNT, SUM, AVG — compute client-side)
- ORDER BY (beyond PartitionKey + RowKey)
- OFFSET (use continuation tokens)
- Subqueries
- UNION
- Transactions (beyond entity group transactions)
- Savepoints
- Foreign keys
- Associations (at DB level)
- Auto-increment IDs
- Schema builder
- Migrations
- Views
- Raw SQL
- Pessimistic locking

**Effort**: 30-40 hours
**Risk**: Medium — requires a non-`database/sql` backend, but the query mapping is well-defined
**Verdict**: Possible if there is user demand. The query capabilities are limited but real.

### Option C: Array Source Workaround (Zero Effort)

Neat already has an "array-backed source" feature (`array` driver) that allows querying in-memory data. Azure Table Storage entities could be loaded into memory and queried via the array source.

**How it would work**:
1. User fetches entities from Azure Table Storage using the Azure SDK directly
2. User converts entities to Go structs
3. User wraps the slice in `neat.NewArraySourceFrom(items)`
4. User queries via Neat's query builder (in-memory, no SQL)

**Effort**: 0 hours (already works via array source)
**Risk**: None
**Verdict**: This already works but provides no real Azure Table Storage integration — it's just in-memory querying after manual data loading.

### Option D: No Support (Also Valid)

Do not support Azure Table Storage. Document that Neat is a SQL ORM and Azure Table Storage is a NoSQL key-value store — they serve fundamentally different use cases.

**Effort**: Zero
**Risk**: None
**Verdict**: Valid if we don't see demand from users.

## Recommendation: Option B (OData Adapter) or Option D (No Support)

The recommendation depends on user demand:

### If there is user demand → Option B (OData Adapter)

Azure Table Storage's query capabilities are limited but real. The OData filter syntax maps well to Neat's `Where()` clauses, and basic CRUD operations (INSERT, UPDATE, DELETE, SELECT with filters) are all supported. An adapter could provide genuine value for users who want to use Neat's query builder API against Azure Table Storage.

The adapter would need to:
- Translate `Where()` to OData `$filter` strings
- Translate `Limit()` / `Take()` to `$top`
- Translate `Select()` to `$select`
- Handle pagination via continuation tokens (not OFFSET)
- Return clear errors for unsupported operations (JOINs, GROUP BY, aggregates, etc.)

This is more feasible than I initially assessed. The query mapping is well-defined and the Azure SDK handles the HTTP/REST layer.

### If there is no user demand → Option D (No Support)

If no users are asking for Azure Table Storage support, the effort (30-40 hours) is not justified. The array source workaround (Option C) already provides a way to query Azure Table Storage data in memory.

### Why Not Option A (Full Abstraction)

Option A (redesigning Neat to support both SQL and NoSQL backends) is not recommended regardless of demand. It would require 100+ hours and fundamentally change Neat's architecture. Option B provides most of the value at a fraction of the effort.

### What About Cosmos DB?

Azure Cosmos DB has a MongoDB-compatible API and a Cassandra API, but neither is SQL-compatible in the `database/sql` sense. The same recommendation applies: Neat is a SQL ORM, not a NoSQL ORM.

Cosmos DB for PostgreSQL (formerly Hyperscale/Citus) is a different story — it **is** PostgreSQL-compatible and would work with Neat's PostgreSQL driver. But that's just PostgreSQL, not Cosmos DB's native API.

## Documentation Update

If we go with Option D, we should:

1. **Add a note to `docs/driver-registration.html`** explaining that Neat is a SQL ORM and does not support NoSQL databases (Azure Table Storage, Cosmos DB native API, DynamoDB, MongoDB, etc.)

2. **Do not add Azure Table Storage to comparison tables** — it's not a SQL database and comparing it to SQL ORMs is misleading

3. **Document the array source workaround** in `docs/array-source.html` — users can load NoSQL data into memory and query it with Neat's query builder

## Estimated Effort

- **Option B (OData Adapter)**: 30-40 hours (adapter + tests + documentation)
- **Option C (Array Source Workaround)**: 0 hours (already works)
- **Option A (Full Abstraction)**: 100+ hours (not recommended)
- **Option D (No Support)**: 1-2 hours (documentation only)

## Success Criteria (Option D)

1. `docs/driver-registration.html` includes a note explaining Neat is SQL-only
2. No Azure Table Storage rows are added to comparison tables
3. `docs/array-source.html` mentions the workaround for NoSQL data

## Decision Required

**Should Neat support Azure Table Storage via an OData adapter (Option B), or should we not support it (Option D)?**

Arguments for Option B:
- Azure Table Storage has real query capabilities (OData filters, projection, limiting)
- The query concepts map well to Neat's `Where()`, `Select()`, and `Limit()` methods
- Basic CRUD (INSERT, UPDATE, DELETE, SELECT) is fully supported
- Would differentiate Neat from other Go ORMs (none support Azure Table Storage)
- Useful for users building on Azure who want ORM-like abstraction

Arguments for Option D:
- Requires a non-`database/sql` backend (architectural change)
- ~60% of Neat's features would not work (JOINs, aggregates, GROUP BY, associations, etc.)
- 30-40 hours of effort for a partial feature set
- Users may be better served by the Azure SDK directly
- No SQL ORM competitor supports it (not a competitive gap today)

**Recommendation**: Defer until there is user demand. If demand exists, implement Option B (OData adapter) — it provides real value and the query mapping is well-defined. Do not attempt Option A (full abstraction).

## References

- Azure Table Storage Go SDK: https://github.com/Azure/azure-sdk-for-go/tree/main/sdk/data/aztables
- Azure Table Storage Documentation: https://learn.microsoft.com/en-us/azure/storage/tables/
- Azure Cosmos DB for Table: https://learn.microsoft.com/en-us/azure/cosmos-db/table/
- Neat Array Source: `docs/array-source.html`
- Neat Driver Registration: `docs/driver-registration.html`
