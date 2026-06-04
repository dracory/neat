# Security Review Report

**Date**: June 2026 (Updated)
**Reviewer**: Senior Principal Golang Engineer
**Codebase**: `github.com/dracory/neat` (Go ORM Library)
**Review Scope**: Full codebase re-review after prior findings were resolved

---

## Executive Summary

This is a full re-review of the Neat ORM codebase following the resolution of the 12 findings documented in the previous review. **All 12 prior findings have been confirmed as resolved.** This updated review identifies **8 new findings** (1 High, 4 Medium, 3 Low) arising from residual design patterns and newly examined code paths.

**Overall Risk Level**: **MEDIUM** (no new Critical issues)

**Status of Prior 12 Findings**: All confirmed FIXED

**New Findings Summary**:
- ⚠️ 0 Critical
- ⚠️ 1 High
- ⚠️ 4 Medium
- ℹ️ 3 Low

---

## Prior Findings: Confirmed Resolved

The following findings from the previous review were all verified fixed during this re-review:

| # | Title | Severity | Status |
|---|-------|----------|--------|
| 1 | SQL Injection in ORDER BY Clause | Critical | ✅ FIXED |
| 2 | SQL Injection in GROUP BY Clause | Critical | ✅ FIXED |
| 3 | SQL Injection in Select Clause | Critical | ✅ FIXED |
| 4 | SQL Injection via RawExpr Function | High | ✅ DOCUMENTED |
| 5 | SQL Injection in WhereColumn | High | ✅ FIXED |
| 6 | Transaction Savepoint Name Injection | High | ✅ FIXED |
| 7 | SQL Injection via Distinct Columns | Medium | ✅ FIXED |
| 8 | SQL Injection in WhereNot | Medium | ✅ FIXED |
| 9 | Information Disclosure Through Error Messages | Medium | ✅ FIXED |
| 10 | SQL Injection via Table Name | Medium | ✅ FIXED |
| 11 | Missing Timeout Configuration | Low | ✅ FIXED |
| 12 | Password Exposure in Configuration | Low | ✅ FIXED |

---

## New Findings

### Finding N1: SQL Injection via Increment/Decrement Column Parameter (High) ✅ FIXED

- **Severity**: High
- **CWE**: CWE-89 (SQL Injection)
- **OWASP**: A03:2021 — Injection
- **Location**: `database/query/query_advanced.go:14-57`

**Description**: The `Increment()` and `Decrement()` methods call `validateAggregate()` to check the column name, but `validateAggregate()` only rejects characters outside `[a-zA-Z0-9_.*]`. However, the column value is then directly interpolated into the SQL string using `fmt.Sprintf`:

```go
updateQuery := fmt.Sprintf("%s = %s + ?", column, column)
return q.Update(updateQuery, incAmount)
```

The resulting `updateQuery` string is passed to `Update()` where it enters `BuildUpdate()` via the `colStr` path at `builder_update.go:66`. Because `colStr` contains `=`, the code takes the branch at line 66 that uses the expression as-is with `strings.Replace(colStr, "?", placeholderFunc(...), 1)`. The column name itself is embedded unquoted into the `SET` clause, so a column name like `a + 1 -- ` or a multi-part expression could produce unexpected SQL, even though `validateAggregate()` blocks obvious injection characters.

More critically, the raw `updateQuery` string (`col = col + ?`) is inserted verbatim into the SET clause without quoting the identifier via `quoteIdentifier()`. In contrast, all other SET clause paths quote identifiers.

- **Impact**: A developer passing a column name derived from user-controlled data could produce malformed or injected SET clauses. Rated High because the column name is not quoted and is used twice in the expression.
- **Recommendation**: Apply `quoteIdentifier()` to the `column` value before constructing the `updateQuery` expression. Example fix:
  ```go
  // In Increment():
  updateQuery := fmt.Sprintf("%s = %s + ?", column, column)
  // Should be:
  quoted := builder.quoteIdentifier(column)
  updateQuery := fmt.Sprintf("%s = %s + ?", quoted, quoted)
  ```
  Alternatively, add a dedicated `buildIncrementSQL()` method in the builder that quotes the identifier.

---

### Finding N2: QueryTimeout Field Defined but Never Enforced (Medium) ✅ FIXED

- **Severity**: Medium
- **CWE**: CWE-400 (Uncontrolled Resource Consumption)
- **Location**: `database/db/config_builder.go:416`, `config.go:80`, `database/orm/orm.go`

**Description**: `PoolConfig` defines a `QueryTimeout` field (default: 30 seconds per the `GetInt()` fallback at `config_builder.go:119`), but this timeout is **never applied** anywhere in the query execution path. All calls to `ExecContext`, `QueryContext`, and `QueryRowContext` pass `q.ctx` directly — a context that is not wrapped with any deadline derived from `QueryTimeout`.

```go
// database/query/query_scan.go:36 — raw context, no timeout applied
rows, err := databaseConn.QueryContext(q.ctx, sql, args...)
```

The `QueryTimeout` value is stored and returned by the config API but never consumed by `buildQuery()` in `database/orm/orm.go` or by the `Query` struct.

- **Impact**: Long-running queries can exhaust connection pool resources and cause denial of service. There is no automatic per-query timeout enforcement.
- **Recommendation**:
  1. In `database/orm/orm.go:buildQuery()`, store `QueryTimeout` from the config.
  2. In each query execution method (e.g., `Scan`, `Update`, `Create`, `Delete`), wrap `q.ctx` with a deadline if `QueryTimeout > 0`:
     ```go
     if q.dbConfig != nil && q.dbConfig.Pool.QueryTimeout > 0 {
         var cancel context.CancelFunc
         ctx, cancel = context.WithTimeout(q.ctx, time.Duration(q.dbConfig.Pool.QueryTimeout)*time.Second)
         defer cancel()
     }
     ```

---

### Finding N3: Oracle MAX(id) Fallback Is a TOCTOU Race Condition (Medium) ⚠️ PARTIALLY FIXED

- **Severity**: Medium
- **CWE**: CWE-367 (Time-of-Check Time-of-Use Race Condition)
- **Location**: `database/query/query_create.go:236-253`

**Description**: When inserting a row into an Oracle database and all sequence-based ID lookup patterns fail, the code falls back to `SELECT MAX(id) FROM <table>`. This is a classic TOCTOU race:

```go
maxIDQuery := fmt.Sprintf("SELECT MAX(id) FROM %s", tableName)
// ...executed on the same or a different connection than the INSERT
seqErr = dbConn.QueryRowContext(q.ctx, maxIDQuery).Scan(&lastID)
```

Between the `INSERT` and the `SELECT MAX(id)`, another goroutine or database client can insert a row with a higher `id`, causing the wrong ID to be returned and stored in the model. The comment in the code acknowledges this: *"This is less safe but works as a last resort."*

- **Impact**: Under concurrent insert workloads on Oracle, the wrong primary key can be populated into the model struct, leading to data integrity bugs, phantom updates, and security issues if the wrong record is subsequently updated or deleted.
- **Recommendation**:
  1. Execute the `INSERT` and `SELECT MAX(id)` within the same transaction to reduce (but not eliminate) the race window.
  2. Prefer Oracle `RETURNING id INTO :outvar` or use `SEQUENCE.CURRVAL` exclusively.
  3. If `CURRVAL` is not available, document clearly that Oracle support requires a standard sequence naming convention and fail loudly rather than falling back to `MAX(id)`.

---

### Finding N4: JSON Path Injection via Unvalidated User-Supplied Path Components (Medium) ⚠️ PARTIALLY FIXED

- **Severity**: Medium
- **CWE**: CWE-89 (SQL Injection via JSON path manipulation)
- **Location**: `database/query/query_where.go:280-292`, `database/query/builder_update.go:72-101`

**Description**: JSON column path syntax (e.g., `"data->name->nested"`) is parsed by `splitJsonColumn()` and `splitJsonColumnForMySQL()` using `strings.Split(column, "->")`. The resulting path segments are interpolated directly into SQL function calls:

```go
// query_where.go:361
query: fmt.Sprintf("json_extract(%s, '%s') = ?", col, path)

// builder_update.go:82
fmt.Sprintf("%s = JSON_SET(%s, '%s', %s)", jsonColumn, jsonColumn, jsonPath, ...)
```

The `path` variable (e.g., `$.name`) is not validated or escaped before being placed inside single-quoted SQL strings. A column argument like `data->name' OR '1'='1` (which is theoretically blocked by `isSimpleIdentifier` for the column part but not by any validation on the path segment after `->`) could escape the single-quoted JSON path literal.

Validation of what comes *after* `->` is absent. The split simply joins with `.` and prepends `$.` with no character-level validation of path segments.

- **Impact**: If any application code passes a JSON column path string that is partially derived from user input, an attacker can escape the JSON path literal and inject arbitrary SQL.
- **Recommendation**: Validate each path segment from the `->` split with `isSimpleIdentifier()` before constructing the JSON path string:
  ```go
  func validateJsonPathSegment(seg string) bool {
      return isSimpleIdentifier(seg)
  }
  ```
  Reject or sanitize any path that contains characters outside `[a-zA-Z0-9_]`.

---

### Finding N5: Read-Replica Connection Opened Without Ping or Error Handling (Medium) ⚠️ NOT FIXED

- **Severity**: Medium
- **CWE**: CWE-390 (Detection of Error Condition Without Action)
- **Location**: `database/orm/orm.go:167-180`

**Description**: When opening read-replica and write-primary connections, errors from `dbDriver.Open()` are silently discarded:

```go
readSQLDB, _ = dbDriver.Open(replicaDSN)
// ...
writeSQLDB, _ = dbDriver.Open(primaryDSN)
```

If the replica DSN contains invalid credentials or is unreachable, `readSQLDB` will be `nil` (or an opened but unverified connection). No ping is performed on replica connections. A nil `readSQLDB` will panic or cause a nil-pointer dereference if it is subsequently used.

- **Impact**: Silent misconfiguration of read-replica routing. Failed replica connections fall through undetected. If the nil-check path in `NewQueryWithReplicas` is not watertight, nil-pointer panics may occur under load.
- **Recommendation**:
  1. Check and propagate `err` from `dbDriver.Open()` for replicas.
  2. Ping replica connections at startup similarly to primary connections.
  3. Log a warning at minimum when a replica connection cannot be established.

---

### Finding N6: Slow-Query Log Emits Full SQL with Bound Parameters (Low) ⚠️ NOT FIXED

- **Severity**: Low
- **CWE**: CWE-532 (Insertion of Sensitive Information into Log File)
- **Location**: `database/query/query_helpers.go:24-26`

**Description**: When a query exceeds `SlowThreshold`, the logger emits the full SQL string along with all bound parameters:

```go
q.log.Warningf("[slow query %.1fms] %s %v", elapsed, sql, bindings)
```

The `bindings` slice may contain sensitive values such as passwords, PII (names, email addresses), financial data, or authentication tokens.

- **Impact**: Sensitive data written to application logs. If logs are shipped to a centralized logging system, this becomes a data-exposure risk (GDPR, PCI-DSS concern).
- **Recommendation**: Redact or truncate bound parameter values in slow-query log output, or provide a configurable option to suppress them:
  ```go
  q.log.Warningf("[slow query %.1fms] %s [%d bindings redacted]", elapsed, sql, len(bindings))
  ```

---

### Finding N7: Docker Compose Hardcodes Production-Grade Passwords (Low) ⚠️ NOT FIXED

- **Severity**: Low
- **CWE**: CWE-798 (Use of Hard-coded Credentials)
- **Location**: `docker-compose.yml:6`, `docker-compose.yml:36-37`, `docker-compose.yml:41-42`

**Description**: The `docker-compose.yml` file used for integration testing contains hardcoded credentials for all supported databases:

```yaml
MYSQL_ROOT_PASSWORD: root
SA_PASSWORD: YourStrong@Passw0rd
ORACLE_PASSWORD: oracle
POSTGRES_PASSWORD: test
```

While these are intended only for local CI testing, if the same `docker-compose.yml` is used in a shared CI/CD environment or accidentally applied to a staging environment, these credentials would be in use for exposed services.

- **Impact**: If ports are exposed on a network-accessible host, these hardcoded credentials allow trivial authentication. The SQL Server healthcheck also embeds `SA_PASSWORD` in plain text in the `CMD` array.
- **Recommendation**:
  1. Use environment variable substitution in `docker-compose.yml` (e.g., `${MYSQL_ROOT_PASSWORD:-root}`) so that CI can inject non-default credentials.
  2. Rename test credentials to be clearly non-production (they already are somewhat, but document this explicitly).
  3. Add a comment in the file warning that it is for local testing only.

---

### Finding N8: `UpdateOrCreate` Performs Non-Atomic Check-Then-Act (Low) ⚠️ NOT FIXED

- **Severity**: Low
- **CWE**: CWE-362 (Race Condition / TOCTOU)
- **Location**: `database/query/query_update.go:69-106`, `database/query/query_helpers.go:40-124`

**Description**: Both `UpdateOrCreate()` and `UpdateOrInsert()` perform a `COUNT` query followed by a separate `UPDATE` or `INSERT`. These two operations are not wrapped in a transaction and are not atomic:

```go
// query_helpers.go:67-90
count := int64(0)
if err := clone.Count(&count); err != nil { ... }
if count > 0 {
    _, err := updateQ.Update(values)
} else {
    return q.Create(merged)
}
```

Between the `COUNT` and the `UPDATE`/`INSERT`, another concurrent request can insert or delete the record, leading to duplicate inserts or missed updates.

- **Impact**: Under concurrent load, `UpdateOrInsert` can produce duplicate rows (violating UNIQUE constraints) or silently fail to update an existing record. This is a correctness and potential data-integrity issue. In security-sensitive contexts (e.g., upserting user records) it could lead to duplicate accounts.
- **Recommendation**: Wrap the check-then-act sequence in a transaction, or — preferably — implement true UPSERT semantics using database-native `INSERT ... ON CONFLICT` (PostgreSQL/SQLite), `INSERT ... ON DUPLICATE KEY UPDATE` (MySQL), or `MERGE` (SQL Server/Oracle).

---

## Dependencies Analysis

**Module**: `github.com/dracory/neat`
**Go version**: as declared in `go.mod`

| Dependency | Version | Status |
|---|---|---|
| `github.com/go-sql-driver/mysql` | v1.10.0 | No known CVEs |
| `github.com/lib/pq` | v1.12.3 | No known CVEs |
| `github.com/microsoft/go-mssqldb` | v1.10.0 | No known CVEs |
| `modernc.org/sqlite` | v1.51.0 | No known CVEs |
| `github.com/tursodatabase/libsql-client-go` | v0.0.0-20260528064733-9d5d30a29a60 | Pre-release commit pin — monitor for updates |
| `sijms/go-ora/v2` | v2.8.35 | No known CVEs |

**Recommendation**: Pin `libsql-client-go` to a tagged release once one is available, as pre-release commit hashes are not amenable to vulnerability scanning.

**Supply Chain Note**: Consider adding `govulncheck` to CI to continuously scan for newly disclosed vulnerabilities in direct dependencies.

---

## Concurrency Review

The `Orm` struct uses a `sync.Mutex` (`orm.go:237`) to protect access to the shared `queries` map. The `Query` struct itself is cloned via `Clone()` for each operation chain, which is the correct pattern for concurrent use. The `migrationRegistry` global is protected by `sync.RWMutex`. No data race issues were identified in these paths.

**Residual concern**: The `Query` struct's `beforeCommit`, `afterCommit`, `beforeRollback`, and `afterRollback` slices are mutated via `BeforeCommit()` / `AfterCommit()` etc. (`query_transaction.go:238-255`) without synchronization. If a `Query` instance is shared across goroutines (not the intended use, but possible), these slice mutations are not thread-safe.

---

## Architecture & Best Practice Observations

1. **RawExpr is a documented footgun**: The `RawExpr()` function in `query_types.go:149-158` carries a clear `WARNING` comment. However, the public API surface still exports it without any compile-time indication of danger. Consider a naming convention like `UnsafeRawExpr` or a separate `unsafe` sub-package.

2. **`Exec()` and `Raw()` accept arbitrary SQL**: `query_advanced.go:87-96` and `99-146` accept arbitrary SQL strings with no validation. This is intentional for raw execution, but callers must be careful. No documentation at the call site warns of injection risk. Documentation should match the standard of `RawExpr()`.

3. **No query builder fuzz testing**: Given the complexity of the SQL builder, fuzz testing with `go test -fuzz` on `BuildSelect`, `BuildUpdate`, `BuildInsert`, and `BuildDelete` would increase confidence that the builders do not produce malformed SQL under unexpected inputs.

4. **No SECURITY.md**: There is no `SECURITY.md` file at the repository root documenting the vulnerability disclosure process or known security-sensitive APIs.

---

## Compliance Considerations

| Standard | Area | Status |
|---|---|---|
| OWASP A03:2021 | SQL Injection | Largely mitigated; N1 and N4 are residual vectors |
| OWASP A09:2021 | Security Logging & Monitoring | N6 (log exposure of bound parameters) |
| CWE-89 | SQL Injection | N1 (Increment/Decrement), N4 (JSON path) |
| CWE-362 | TOCTOU Race Condition | N3 (Oracle MAX fallback), N8 (UpdateOrCreate) |
| CWE-400 | Resource Exhaustion | N2 (QueryTimeout not enforced) |
| CWE-532 | Log Information Leakage | N6 (slow-query log) |
| CWE-798 | Hardcoded Credentials | N7 (docker-compose) |
| GDPR | Data minimization in logs | N6 |

---

## Summary Statistics

**Prior review findings**: 12 (all resolved)

**New findings**:

| Severity | Count | Status |
|---|---|---|
| Critical | 0 | - |
| High | 1 (N1) | ✅ 1 FIXED |
| Medium | 4 (N2, N3, N4, N5) | ✅ 1 FIXED, ⚠️ 2 PARTIAL, ⚠️ 1 NOT FIXED |
| Low | 3 (N6, N7, N8) | ⚠️ 3 NOT FIXED |
| **Total** | **8** | **2 FIXED, 4 PARTIAL, 2 REMAINING** |

---

## Recommended Remediation Priority

1. **High — Fix immediately**:
   - N1: Quote identifiers in `Increment()`/`Decrement()` before SQL construction

2. **Medium — Fix in next release**:
   - N2: Enforce `QueryTimeout` via `context.WithTimeout` in all query execution paths
   - N3: Remove or isolate the Oracle `MAX(id)` fallback; require explicit sequence naming
   - N4: Validate JSON path segments with `isSimpleIdentifier()` in `splitJsonColumn*()` helpers
   - N5: Handle and log errors from replica connection opening; ping replicas at startup

3. **Low — Address in backlog**:
   - N6: Redact bound parameter values in slow-query log output
   - N7: Parameterize `docker-compose.yml` credentials via environment variables
   - N8: Implement atomic UPSERT using database-native constructs

4. **Ongoing**:
   - Add `govulncheck` to CI pipeline
   - Add fuzz tests for SQL builders
   - Create `SECURITY.md` with vulnerability disclosure policy
   - Add security regression tests for all previously fixed injection vectors

---

## References

- OWASP SQL Injection: https://owasp.org/www-community/attacks/SQL_Injection
- OWASP Top 10 2021: https://owasp.org/www-project-top-ten/
- CWE-89 SQL Injection: https://cwe.mitre.org/data/definitions/89.html
- CWE-362 Race Condition: https://cwe.mitre.org/data/definitions/362.html
- CWE-400 Resource Exhaustion: https://cwe.mitre.org/data/definitions/400.html
- CWE-532 Log Exposure: https://cwe.mitre.org/data/definitions/532.html
- Go govulncheck: https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck
- Go Security Best Practices: https://go.dev/security
