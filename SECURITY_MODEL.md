# Security model: dracory/neat

## System overview
neat is a Go ORM, query builder, schema-builder and migration library (`github.com/dracory/neat`). It is not deployed on its own. It is linked into host applications, which pass it SQL values, identifiers, DSNs and, for the file-based drivers, file paths. It talks to MySQL, PostgreSQL, SQLite (modernc), SQL Server, Oracle, Turso/libsql and Azure Table Storage (via `dracory/aztablessql`). It also ships file-backed drivers (CSV/JSON/XML/godb/array).

There is no infrastructure repository: the user confirmed it is a library. Deployment controls (TLS termination, DB network ACLs, secrets storage) belong to the host application and are out of view.

**Trust boundaries**
1. Host application → neat's public API (`contracts/`, `database/query`, `database/orm`, `database/schema`, `database/migrator`). The main boundary: a value or identifier the host app takes from its end users crosses it.
2. neat → database server/driver (network, DSN credentials, TLS settings such as `sslmode`, defaulting to `require` for PostgreSQL in `database/db.go`).
3. neat → host logs/errors (query log, slow-query warnings, `sanitizeError`, DSN redaction).
4. neat → local filesystem (file-based drivers and sources). Out of scope by user decision, see below.

## Components
| Component | Covers |
|---|---|
| **Query Builder** | `database/query/` (builder_*.go, query_*.go, utils.go `isSimpleIdentifier`/`isValidColumnReference`, builder_quote.go `quoteIdentifier`, error_sanitizer.go, to_sql.go) |
| **Connections & Drivers** | `database/db.go` (NewFromDSN, `redactDSN`), `database/db/` (dsn.go, config_builder.go, expression.go), `database/driver/{driver,placeholder,mysql,postgres,sqlite,sqlserver,oracle,turso,aztables}.go` |
| **Schema & Migrations** | `database/schema/` (blueprint, grammars/*, processors), `database/migrator/` (migrator.go, tracker.go, validation.go `ValidateTableName`), `database/seeder/` |
| **ORM & Models** | `database/orm/`, `database/association/`, `database/observer/`, `database/soft_delete/`, `database/cursor/` |
| **File-based Drivers** | `database/driver/{csvdb*,jsondb*,xmldb*,godb*,array*,schemainfer}.go`, `support/{csvsource,jsonsource,xmlsource,arraysource}/`, `array_source.go` |
| **Support & Config** | `support/{collect,convert,str,structref,env,database}/`, `config.go`, `constants.go`, `event.go`, `errors/`, `contracts/` |
| **Tests & Examples** | `examples/`, `integration_tests/`, `benchmarks/`, `**/*_test.go`, `docker-compose.yml`, `.github/workflows/` |

## Entry points and who can reach them
- **Values** (bindings to `Where(...,"?", v)`, `Create`/`Update`/`Save` model fields, `WhereIn`, `WhereLike`, `WhereJsonContains`, pagination/limit arguments): reachable by the host app's end users. **Untrusted** (user).
- **Identifiers**: table names (`Table`), columns in `Select`, `OrderBy`/`OrderByDesc`, `Group`, `Distinct`, `WhereIn(column,…)`, `WhereColumn`, `WhereAny/All/None` columns, JSON paths, schema blueprint column/table/index names, migration table names. **Untrusted** (user): a host app may pass a user-chosen sort column or filter field.
- **DSNs and connection config** (`NewFromDSN`, `ParseDSN`, `config_builder.go`): **untrusted** (user), e.g. multi-tenant apps that build DSNs from tenant data.
- **Raw SQL**: `Raw()`, `db.NewRawExpression`, `OrderByRaw`/`SelectRaw`-style raw fragments. **Trusted / developer-controlled** (user). Injection that needs attacker control of a Raw SQL string is accepted risk.
- **SQL clause strings** with placeholders (`Where("name = ? AND …")`, `Having`, `Join` query strings): **untrusted** (user). An app might build them from user input, so only `Raw()`/`RawExpression` SQL is developer-controlled.

## Assets and data sensitivity
1. Integrity and confidentiality of the host application's database: injection would allow reading or modifying any row the connection can reach. Highest.
2. Database credentials and tokens in DSNs (passwords, Turso `authToken`/`jwt`, Azure keys). Must not leak via logs or errors.
3. Bound values (often PII) in the query log (`QueryLog.Bindings`) and in error messages.
4. Migration history integrity (`migration_tracker`, append-only by design) and schema state.
5. Availability of the host app's DB connection pool (timeouts via `timeoutContext`, transactions/savepoints, connection leaks).

## Attacker profiles
- **A1: Remote end user of a host application.** Controls request data the host app passes into neat as values, identifiers, JSON paths, pagination and, per the user's answer, DSN components. Primary attacker.
- **A2: Malicious or compromised database server / network position.** Returns crafted results or intercepts traffic (TLS defaults, result parsing, Oracle ID retrieval). Secondary attacker.
- **A3: Supply chain.** Vulnerable dependency versions in `go.mod` (drivers, `golang.org/x/*`, gorm). Secondary attacker.
- Not considered: the host-app developer writing Raw SQL (trusted).

## Existing controls
- Parameterized bindings throughout the builder (`?` placeholders rewritten per dialect in `database/driver/placeholder.go`).
- `isSimpleIdentifier` / `isTableIdentifier` / `isValidColumnReference` in `database/query/utils.go` are applied to `Table`, `OrderBy`, `Group`, `Distinct` and savepoint names. `quoteIdentifier` (`builder_quote.go`) escapes embedded quote characters (fixed per `docs/security-review.md` #1).
- `ValidateTableName` in `database/migrator/validation.go`.
- `sanitizeError` hides SQL details unless debug mode is on. Context errors are passed through.
- `redactDSN` / `redactQueryParams` in `database/db.go` redact userinfo plus `authToken`/`auth_token`/`jwt`.
- Slow-query warnings redact bindings. The QueryLog binding issue is marked fixed in `docs/security-review.md` #2.
- CI runs `gosec` and `golangci-lint` (`.github/workflows/tests.yml`).
- Prior internal review: `docs/security-review.md` (2026-07-19); user guide: `docs/security.html`.

## Out of scope / accepted risk (user decisions)
- **Test and demo code**: `examples/`, `integration_tests/`, `benchmarks/`, `*_test.go`. Never shipped.
- **File-based drivers and sources** (csvdb, jsondb, xmldb, godb, array, and the `support/*source` parsers): only for local/dev data. This covers path traversal and parsing issues in them.
- **Raw SQL** passed to `Raw()` / `RawExpression`: developer-controlled.

## Open questions
- Do host apps run with debug mode (`IsDebug`) enabled in production? That would expose full errors and queries in logs.
- Which dialects matter most in production? Findings specific to Oracle, SQL Server or Azure Tables could be weighted accordingly.
