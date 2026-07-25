# Security Review Report
Date: 2026-07-19
Reviewer: Senior Principal Golang Engineer (Cascade)
Codebase: neat — full codebase review

## Executive Summary

This review covers the entire neat ORM codebase — query builder, error handling, logging, CI/CD, configuration, concurrency, and dependencies.

The codebase demonstrates strong security practices: parameterized queries are used consistently, identifier validation (`isSimpleIdentifier`) is applied at key trust boundaries, error sanitization prevents SQL detail leakage in production, and slow-query logging redacts binding values. However, several issues were identified that warrant attention.

**Summary**: 0 Critical, 0 High, 2 Medium, 1 Low findings.

---

## Medium Severity Findings

### Finding #1: `quoteIdentifier` Does Not Sanitize Quote Characters Within Identifiers ✅ FIXED

- **Location**: `database/query/builder_quote.go:8-58`
- **Description**: The `quoteIdentifier` function wraps identifiers in quote characters (`"` or `` ` ``) but does not check for or escape embedded quote characters within the identifier name. If an identifier contains a quote character (e.g., `user"name` or `user\`name`), the resulting SQL will have unbalanced quotes, potentially allowing SQL injection.
- **Impact**: Defense-in-depth issue. `Table()` validates via `isSimpleIdentifier`, and other code paths (e.g., `Select`, `OrderByRaw`) are developer-controlled, not user input. Not directly exploitable, but escaping embedded quotes prevents potential misuse if identifiers ever reach `quoteIdentifier` without prior validation. CWE-89.
- **Code**:
```go
// database/query/builder_quote.go:57
return fmt.Sprintf("%s%s%s", quoteChar, name, quoteChar)
// If name contains quoteChar, this produces: "user"name" — broken quoting
```
- **Recommendation**: Add escaping of the quote character within the identifier:
```go
// Escape embedded quote characters
name = strings.ReplaceAll(name, quoteChar, quoteChar+quoteChar)
return fmt.Sprintf("%s%s%s", quoteChar, name, quoteChar)
```

---

## Medium Severity Findings

### Finding #2: `logQuery` Stores Full SQL in QueryLog

- **Location**: `database/query/query_helpers.go:14-22`
- **Description**: The `logQuery` function stores the full SQL query string and bindings in the query log (`q.queryLog`). While bindings are redacted in the slow-query warning log, they are stored unredacted in the `QueryLog` struct.
- **Impact**: If the query log is serialized, printed, or exposed through an API, binding values (which may contain user PII or sensitive data) are exposed. CWE-532 (Insertion of Sensitive Information into Log File).
- **Code**:
```go
// database/query/query_helpers.go:17-21
*q.queryLog = append(*q.queryLog, contractsorm.QueryLog{
    Query:    sql,
    Bindings: bindings,  // unredacted bindings stored
    Time:     elapsed,
})
```
- **Recommendation**: The slow-query warning already redacts bindings (`[%d bindings redacted]`). Apply the same treatment to the `QueryLog` struct, or add a `RedactedBindings` field. If full bindings are needed for debugging, ensure the query log is only accessible in debug mode.

---

## Low Severity Findings

### Finding #3: `sanitizeError` Keyword Check Is Too Narrow

- **Location**: `database/query/error_sanitizer.go:36-40`
- **Description**: The error sanitizer checks for "sql", "query", and "syntax" keywords to decide whether to suppress error details. Table and column names in error messages aren't typically secret, and the sanitizer catches the most common SQL error patterns. Expanding the keyword list would be hardening, not fixing a vulnerability.
- **Impact**: Minimal — database schema names in error messages are low sensitivity.
- **Code**:
```go
// database/query/error_sanitizer.go:36-40
if strings.Contains(strings.ToLower(errMsg), "sql") ||
    strings.Contains(strings.ToLower(errMsg), "query") ||
    strings.Contains(strings.ToLower(errMsg), "syntax") {
    return fmt.Errorf("database operation failed")
}
```
- **Recommendation**: Optionally expand the keyword list to include "column", "table", "constraint" for additional hardening. Low priority.

---

## Positive Security Practices Observed

1. **Parameterized queries**: All data values use `?` placeholders with `args` — no string interpolation of user data in the core query builder.
2. **`isSimpleIdentifier` validation**: Applied to `Table()`, `OrderBy()`, `OrderByDesc()`, `Group()`, `Distinct()` — rejects SQL keywords, dots, parentheses, and non-alphanumeric characters.
3. **Error sanitization**: `sanitizeError` prevents SQL detail leakage in production while preserving context errors (`context.Canceled`, `context.DeadlineExceeded`).
4. **Slow-query logging redacts bindings**: `logQuery` logs `[%d bindings redacted]` — binding values are not exposed in log warnings.
5. **`timeoutContext`**: Provides query-level timeout support via `context.WithTimeout` — prevents indefinite query execution.
6. **Context error preservation**: `sanitizeError` explicitly passes through `context.Canceled` and `context.DeadlineExceeded` — callers can properly handle cancellation.
7. **Debug-gated error logging**: Full error details are only logged when `q.IsDebug()` is true.
8. **CI/CD includes `gosec`**: The workflow runs `gosec` for automated security scanning.
9. **CI/CD includes `golangci-lint`**: Static analysis catches common bugs and security issues.
10. **No hardcoded secrets in Go code**: No API keys, passwords, or tokens found in non-test Go source files.

---

## Dependency Analysis (Full)

| Dependency | Version | Status | Notes |
|---|---|---|---|
| `github.com/go-sql-driver/mysql` | v1.10.0 | Current | No known CVEs |
| `github.com/lib/pq` | v1.12.3 | Maintenance mode | Consider migrating to `pgx` |
| `github.com/microsoft/go-mssqldb` | v1.10.0 | Current | No known CVEs |
| `github.com/sijms/go-ora/v2` | v2.9.0 | Current | No known CVEs |
| `github.com/google/uuid` | v1.6.0 | Current | No known CVEs |
| `github.com/samber/lo` | v1.53.0 | Current | No known CVEs |
| `github.com/spf13/cast` | v1.10.0 | Current | No known CVEs |
| `github.com/dromara/carbon/v2` | v2.6.16 | Current | No known CVEs |
| `modernc.org/sqlite` | v1.53.0 | Current | Pure-Go, no CGo |
| `golang.org/x/crypto` | v0.54.0 | Current | No known CVEs |
| `golang.org/x/text` | v0.40.0 | Current | No known CVEs |
| `golang.org/x/exp` | v0.0.0-20260709172345 | Experimental | OK for non-production features |

---

## Summary Statistics

| Severity | Count |
|---|---|
| Critical | 0 |
| High | 0 |
| Medium | 2 |
| Low | 1 |
| **Total** | **3** |

---

## Prioritized Remediation Plan

1. **Finding #1** (quoteIdentifier escaping) — Add quote character escaping. One-line fix, high impact.
2. **Finding #2** (QueryLog stores unredacted bindings) — Redact bindings in query log or gate behind debug mode.
3. **Finding #3** (sanitizeError keyword list) — Optionally expand keyword list. Low priority hardening.
