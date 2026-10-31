# Security Review Report

Date: 2026-10-31
Reviewer: Senior Principal Golang Engineer
Codebase: neat — `github.com/dracory/neat` (ORM, query builder, schema builder, migrations)

> This report supersedes the 2026-07-19 review. All findings from that review were verified as fixed: `quoteIdentifier` now escapes/rejects crafted pre-quoted identifiers (`isSafelyPreQuoted`, `builder_quote.go`), `logQuery` redacts bindings unless debug is on (`query_helpers.go`), and `sanitizeError` retains context errors. Scope and trust boundaries follow `SECURITY_MODEL.md` (file-based drivers, test/demo code, and developer-supplied `Raw()` SQL are out of scope / accepted risk).

## Executive Summary

neat is a library, not a deployed service; the trust boundary is host application → neat API → database. The codebase demonstrates strong defensive practice: bindings are parameterized throughout, identifier validation (`isSimpleIdentifier`, `isTableIdentifier`, `isValidColumnReference`, `isValidJSONPathSegment`) is applied consistently at untrusted entry points (`Table`, `OrderBy`, `Group`, `Distinct`, `WhereIn`, savepoints, JSON paths, migration table names), LIKE wildcards are escaped with `ESCAPE '!'`, slow-query logs redact bindings, and CI runs `gosec` + `golangci-lint`.

The residual risk is concentrated in **DSN handling** (credential interpolation without escaping, and a redaction gap in `redactDSN` for non-URL DSN formats) and one **error-sanitization flaw** where the "sanitized" error still embeds the raw SQL error text.

**Overall risk: Medium.** 0 Critical, 1 High, 4 Medium, 4 Low findings.

---

## Critical Findings (Severity: Critical)

None.

---

## High Severity Findings

### Finding #1: `redactDSN` Does Not Redact Non-URL DSN Formats — Credentials Leak Into Error Messages

- **Location**: `database/db.go:359-392`, used at `database/db.go:210` and `database/db.go:514`
- **Description**: `redactDSN` only strips credentials when the DSN contains `://` (URL-style) or a `mysql://` prefix. Two supported DSN families pass through **unredacted**:
  1. **PostgreSQL keyword/value DSNs** — `buildPostgresDSN` produces `host=… user=… password=SECRET dbname=…` (`config_builder.go:344`). This format has no `://`, so `redactDSN` returns it verbatim.
  2. **go-sql-driver MySQL DSNs** — `user:pass@tcp(host:port)/db?params` has no scheme prefix; the `mysql://` branch can never match a real driver DSN (a string starting with `mysql://` is already caught by the `://` branch above it, making that branch dead code).
  If DSN parsing fails (typo, unreachable host validation, etc.), `fmt.Errorf("failed to parse DSN %s: %w", redactDSN(dsn), err)` embeds the plaintext password in the returned error — which host applications commonly log or surface.
- **Impact**: CWE-532 (Insertion of Sensitive Information into Log Files) / CWE-209. Database credentials exposed through error messages, potentially reaching logs, crash reporters, or end users.
- **Code Example**:
```go
// database/db.go:380-392 — falls through unredacted
if strings.HasPrefix(dsn, "mysql://") {          // dead branch for real driver DSNs
    rest := strings.TrimPrefix(dsn, "mysql://")
    ...
}
return dsn  // "user:pass@tcp(h:3306)/db" or "host=x password=SECRET ..." returned as-is
```
- **Recommendation**: Redact all supported DSN shapes:
```go
// Keyword/value DSNs (postgres): redact password= and any key ending in password
re := regexp.MustCompile(`(?i)(password\s*=\s*)('[^']*'|\S+)`)
dsn = re.ReplaceAllString(dsn, "${1}[REDACTED]")

// go-sql-driver DSNs: user:pass@ before first "/" or after scheme detection
// Match ":<non-empty>@tcp(" / "@(" style userinfo regardless of scheme
```
  Also parse with `url.Parse` where possible instead of string surgery, and add unit tests asserting redaction for every DSN family the builder can emit.

---

## Medium Severity Findings

### Finding #2: `buildPostgresDSN` Allows libpq Parameter Injection via Unquoted Values

- **Location**: `database/db/config_builder.go:344-360`
- **Description**: The keyword/value DSN is assembled with `fmt.Sprintf` and raw values. libpq treats the string as space-separated `key=value` pairs; a value containing whitespace or a quote breaks out of its pair. `Password`, `Schema` (`search_path`), and `Timezone` are concatenated unquoted:
```go
dsn += " search_path=" + b.config.Schema
dsn += " TimeZone=" + b.config.Timezone
```
  A password or schema value such as `x sslmode=disable` injects a new connection parameter — silently downgrading TLS, changing `connect_timeout`, or redirecting `search_path` to an attacker-controlled schema. Per `SECURITY_MODEL.md`, DSN components are explicitly **untrusted** (multi-tenant apps build them from tenant data).
- **Impact**: CWE-88 (Argument/Parameter Injection). TLS downgrade (`sslmode=disable`), credential-relevant parameter injection, schema search-path hijack.
- **Recommendation**: Quote every value per libpq rules (`'` wrapping, `\'`/`\\` escaping), or build the DSN via `url.URL` (the `postgres://` URL form) with `url.UserPassword`/`QueryEscape` for all components. Validate `Schema`/`Timezone`/`SSLMode` against an allowlist (e.g. `isTableIdentifier` for schema names; fixed enum for `sslmode`).

### Finding #3: URL-Style DSNs Interpolate Credentials Without Escaping (MySQL, SQL Server, Oracle)

- **Location**: `database/db/config_builder.go:314` (MySQL), `:412` (SQL Server), `:445` (Oracle)
- **Description**: Credentials and database names are placed into userinfo/path positions without escaping:
```go
dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, host, port, database)
dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", ...)
return fmt.Sprintf("oracle://%s:%s@%s:%d/%s", ...)
```
  A password containing `@`, `/`, `:` (MySQL), or `?`, `#`, `%` (URL DSNs) corrupts the DSN or shifts parsing — e.g. a MySQL password containing `@tcp(` injects a second network address; a SQL Server database name containing `&encrypt=false` appends a query parameter that disables encryption. Turso correctly uses `url.QueryEscape` for its token (`:436`) — the inconsistency shows the omission is accidental.
- **Impact**: CWE-88. Connection-string injection: TLS/encryption downgrade, wrong-host connection (credential phishing to attacker server), or persistent connection failure (DoS).
- **Recommendation**: For URL DSNs use `(&url.URL{Scheme, User: url.UserPassword(u,p), Host: net.JoinHostPort(h,port), Path: db}).String()`; for go-sql-driver, prefer `mysql.Config{}.FormatDSN()` which handles escaping correctly.

### Finding #4: `sanitizeError` "Sanitized" Message Still Contains Full SQL Error Detail

- **Location**: `database/query/error_sanitizer.go:38-40`
- **Description**: When an error matches a SQL keyword, production mode returns `fmt.Errorf("database operation failed: %w", err)`. `%w` **embeds the original message text** in the rendered error — `err.Error()` still contains the full `pq: syntax error at or near …`, `Error 1064 …`, constraint names, column names, etc. The wrap only adds a prefix; nothing is actually redacted. Host apps that return/log `err.Error()` leak exactly what the sanitizer claims to hide.
- **Impact**: CWE-209 (Information Exposure Through an Error Message). Schema disclosure, SQL syntax feedback to attackers probing for injection, table/column enumeration.
- **Recommendation**: Return a static message in production while preserving the chain for `errors.Is` via a wrapper type, e.g.:
```go
type sanitizedError struct{ inner error }
func (e sanitizedError) Error() string { return "database operation failed" }
func (e sanitizedError) Unwrap() error { return e.inner }
```
  …or keep `%w` but document clearly that the detail remains visible and the function only adds a generic prefix — currently the code and comment ("return generic error messages") misrepresent the behavior.

### Finding #5: `Query.Cursor` Goroutine Leaks When Consumer Abandons the Channel

- **Location**: `database/query/query_cursor.go:44-78`
- **Description**: The cursor goroutine sends on `cursorChan` (buffer 10) with `select { case cursorChan <- …: case <-ctx.Done(): }`. When `Pool.QueryTimeout == 0`, `timeoutContext()` returns `context.Background()` (`query_helpers.go:85-88`) — `ctx.Done()` never fires. If the caller stops draining the channel without canceling its own context, the goroutine blocks forever holding `rows` open — a leaked goroutine plus a leaked connection-pool slot.
- **Impact**: CWE-772 / CWE-400 (Uncontrolled Resource Consumption). Repeated abandoned cursors exhaust the `database/sql` connection pool → host app DoS.
- **Recommendation**: Document that callers must drain or cancel, return a `Cancel`/`Close` handle, or default to a bounded timeout when `QueryTimeout` is unset. At minimum, derive the goroutine's context from the caller's `q.ctx` *and* a sane default timeout rather than `Background()`.

---

## Low Severity Findings

### Finding #6: Exported `QuoteIdentifier`/`QuoteValue` Helpers Are Unsafe for MySQL/SQL Server

- **Location**: `database/db/expression.go:122-146`
- **Description**: `QuoteIdentifier` wraps identifiers without escaping embedded quote characters (`` ` `` for MySQL, `]` for SQL Server — `[a]]b` breaks out). `QuoteValue` escapes only `'`, but under MySQL's default SQL mode `\` also escapes inside string literals, so `\'` injection remains possible. These helpers are exported and encourage building SQL by concatenation.
- **Impact**: CWE-89 if a host app feeds user input through them. Not used internally — API foot-gun.
- **Recommendation**: Escape the dialect quote char inside identifiers; deprecate `QuoteValue` in favor of bound parameters, or document it as debug-only.

### Finding #7: `Paginate` Accepts Negative `page`/`limit` Without Validation

- **Location**: `database/query/query_paginate.go:9-13`
- **Description**: `offset := (page - 1) * limit` — negative or zero inputs produce negative `LIMIT`/`OFFSET` and (with extreme values) int overflow of the offset computation. Results are dialect-dependent syntax errors or silently wrong windows.
- **Impact**: Low — malformed queries, no injection (values are formatted via `%d`).
- **Recommendation**: Clamp to `page >= 1`, `limit >= 0` (or return a validation error).

### Finding #8: Panic-Based Error Handling in `support/*source` Parsers

- **Location**: `support/csvsource/csv_source.go:40-127`, `support/jsonsource/json_source.go:56-125`, `support/xmlsource/xml_source.go:59-108`, `support/arraysource/*`
- **Description**: Malformed input (`failed to parse …`) triggers `panic` rather than an error return. These are file/local-dev drivers per `SECURITY_MODEL.md` (out of scope), but a host app that ever parses untrusted uploads through them has a crash-level DoS.
- **Impact**: CWE-248 (Uncaught Exception) — process crash on bad input.
- **Recommendation**: Prefer returning `error`; at minimum document the panic contract prominently.

### Finding #9: No `recover()` Around Observer Callbacks

- **Location**: `database/observer/dispatcher.go:62-90`
- **Description**: Host-registered observer hooks (`Created`, `Creating`, …) run synchronously; a panicking observer propagates up through `Create`/`Update`/etc. and can crash the host process mid-transaction.
- **Impact**: Availability — a buggy third-party observer kills the request/process.
- **Recommendation**: Wrap observer invocation in a `recover` and convert to error, consistent with the existing "continue with other observers even if one fails" policy.

---

## Best Practice Recommendations

1. **DSN handling (highest value)**: centralize DSN construction/redaction behind a single tested module — escape on build, redact on error, and add table-driven tests over every driver family.
2. **Fail the build on gosec**: `.github/workflows/tests.yml` runs `gosec ./... || true` — findings never block merges. Remove `|| true` once the baseline is clean, and pin tool versions (`golangci-lint@vX.Y`, `gosec@vX.Y`) instead of `@latest` for supply-chain integrity.
3. **Debug mode guardrail**: `IsDebug` gates error detail, query bindings, and logging — add a one-line warning in docs/debug paths reminding that debug must be off in production (QueryLog bindings, full SQL errors).
4. **Default query timeout**: `Pool.QueryTimeout == 0` means no timeout anywhere; consider a safe default (e.g. 30s) so abandoned queries/cursors can't pin connections forever.
5. **`Join`/`Having`/`Where` raw clause strings** are pass-through SQL by design — ensure `docs/security.html` explicitly calls these "never concatenate user input" surfaces, matching the threat model.
6. **Commit hygiene**: `docker-compose.yml` contains test credentials (`YourStrong@Passw0rd`, `root`) — acceptable for throwaway test containers, but keep them out of any non-test compose files.

## Dependencies Analysis

- Total dependencies: ~14 direct, ~25 indirect (39 modules in `go.mod`)
- Dependencies with known vulnerabilities: none confirmed (govulncheck not run in this review — recommend `govulncheck ./...` in CI)
- Outdated/unmaintained dependencies:
  - `github.com/lib/pq` — officially unmaintained/deprecated; recommend migrating to `jackc/pgx` long-term.
  - `github.com/golang-sql/civil` — last commit 2022 (indirect via mssqldb; low risk).
  - `github.com/dracory/aztablessql` — first-party; track its own security posture.
- All `golang.org/x/*` and driver versions are current as of this review.

## Compliance Considerations

- **GDPR/PII**: bound values (often PII) are stored in `QueryLog.Bindings` only when `IsDebug()` is on — good default; document that enabling debug mode in production creates a PII-in-logs exposure.
- **PCI-DSS**: not in scope (no card data paths), but credential handling findings (#1–#3) are the relevant control gaps.
- Library code carries **no authn/authz surface** — items like JWT, session management, rate limiting, CORS, and CSRF are the host application's responsibility and were not applicable.

## Summary Statistics

- Total issues found: 9
- Critical: 0
- High: 1
- Medium: 4
- Low: 4

## Next Steps

1. Fix `redactDSN` for keyword/value and driver-native MySQL DSNs (Finding #1); add per-driver redaction tests.
2. Quote/escape all values in `buildPostgresDSN` and validate `SSLMode`/`Schema`/`Timezone` (Finding #2); switch URL DSNs to `url.URL` + `url.UserPassword` / `mysql.Config.FormatDSN` (Finding #3).
3. Correct `sanitizeError` so production errors genuinely hide SQL detail (Finding #4).
4. Bound cursor goroutine lifetime — default timeout or explicit cancellation API (Finding #5).
5. Harden `QuoteIdentifier`/`QuoteValue` or mark `QuoteValue` deprecated (Finding #6); validate `Paginate` inputs (#7).
6. CI: remove `|| true` from gosec, pin scanner versions, add `govulncheck`.
