# Lessons from SB SQL Builder

This document outlines key lessons and best practices that Neat ORM can learn from the SB SQL Builder project, based on a direct comparison of both codebases.

> Status markers: ✅ Already done in Neat | ⚠️ Partial | ❌ Not yet implemented

## Overview

SB SQL Builder demonstrates excellent practices in building a focused, production-ready library with sophisticated error handling, security-first design, and systematic documentation. These lessons help identify where Neat ORM is already strong, and where concrete improvements should be made.

## 1. Error Handling Excellence

### Zero-Panic Philosophy ⚠️

**SB Approach:**
- Eliminated all panics in runtime operations through comprehensive refactoring
- Only fundamental configuration errors (like invalid dialect in constructor) use panic
- All runtime validation errors use error returns or error collection

**Current Neat state (`database/query/query_errors_test.go`):**
Six test functions use `defer recover()` to absorb panics from the standard library when a nil `*sql.DB` is passed to `NewQuery`. This is not a test for correct behavior — it is masking a known panic:
```go
// TestNilDatabaseConnection, TestTransactionCommitError, TestTransactionRollbackError,
// TestDialectErrorHandlingMySQL, TestDialectErrorHandlingPostgreSQL, TestDialectErrorHandlingSQLite
defer func() {
    if r := recover(); r != nil {
        _ = r // Expected panic for nil database connection
    }
}()
```

Additionally, `query_scopes_test.go` has a `TestScopeErrorHandling` test with:
```go
t.Run("scope panic handling", func(t *testing.T) {
    // With the new error handling, panics in user functions are not caught
    defer func() {
        if r := recover(); r != nil { _ = r } // Expected panic - user code panicked
    }()
})
```
This explicitly tests that user-scope panics propagate — which is correct for user panics, but the nil-DB panics are internal and should be handled by Neat itself.

**Lesson for Neat:**
- Fix nil `*sql.DB` in `NewQuery` — check at entry and return `ErrNilDatabase` before any operation
- This eliminates the six `defer recover()` blocks and replaces them with proper `if err == nil { t.Error(...) }` assertions
- User-code panics (scope functions) do not need to be caught by Neat — that is the user's responsibility

### Error Collection Pattern ⚠️

**SB Approach:**
```go
// Errors collected during fluent chaining
func (b *Builder) Column(column Column) BuilderInterface {
    if column.Name == "" {
        b.sqlErrors = append(b.sqlErrors, ErrEmptyColumnName)
        return b
    }
    return b
}

// Validated when SQL is generated
func (b *Builder) Create() (string, error) {
    if err := b.validateAndReturnError(); err != nil {
        return "", err
    }
    // ... SQL generation
}
```

**Current Neat state:** Neat's query builder (`database/query/`) uses a direct terminal-method pattern: methods like `Find`, `First`, `Create`, `Update`, `Delete` always return `error`. There is no fluent-chaining layer that accumulates errors. `Where`, `OrderBy`, `Limit` etc. return `*Query` (not an error-returning interface), so the error collection pattern doesn't directly apply to Neat's architecture. Neat's approach is simpler and arguably cleaner for an ORM.

**Lesson for Neat:**
- Neat's terminal-method pattern is correct for an ORM — do not impose SB's error collection on it
- The lesson to apply is: ensure *all* terminal methods (`First`, `Find`, `Count`, `Pluck`, `Create`, `Update`, `Delete`) check for nil DB and empty table name before executing, rather than relying on the database driver to surface those errors with opaque messages
- A helper like `func (q *Query) validate() error` that runs before every terminal method would centralize this

### Structured Error Types ✅

**SB Approach:**
- Custom `BuilderError` struct with Type and Message fields
- Standard error variables: `ErrEmptyTableName`, `ErrEmptyColumnName`, `ErrNilSubquery`
- Helper functions: `NewValidationError()`, `NewConfigurationError()`

**Current Neat state (`errors/errors.go`):**
```go
type StructuredError struct {
    Type    string
    Message string
    Err     error
    Module  string
}

var (
    ErrNilDatabase      = &StructuredError{Type: "ValidationError", Message: "database connection cannot be nil"}
    ErrNilQuery         = &StructuredError{Type: "ValidationError", Message: "query cannot be nil"}
    ErrNotInTransaction = &StructuredError{Type: "ValidationError", Message: "operation requires an active transaction"}
    // ...
)
```

Neat already has `StructuredError`, `ErrNilDatabase`, `ErrNotInTransaction`, `NewValidationError()`, `NewArgumentError()`, and `NewConfigurationError()`. This is fully equivalent to SB's pattern and includes the `Module` field for better context.

**No action required.** The structured error infrastructure is in place. The gap is in *using* it consistently (see Zero-Panic section above).

## 2. Security First Approach

### Parameterized Queries by Default ✅

**SB Approach:**
- Parameterized queries with SQL injection protection as default
- Database-specific placeholder support (`?`, `$1`, `@p1`)
- Legacy mode available via `WithInterpolatedValues()` for backward compatibility

**Current Neat state:** Neat's `Driver.Placeholder(n int) string` method (defined in `database/driver/driver.go`) returns the appropriate placeholder per dialect (`?` for MySQL/SQLite, `$1` for Postgres, `@p1` for SQL Server). The query layer uses these placeholders with parameterized execution. This is the correct, secure approach and it is already implemented.

**No action required** for parameterized queries. Neat is ahead of many ORMs here.

**One gap:** DSN redaction. `redactDSN()` exists in `database/db.go` to mask passwords in log output — this is good. Ensure it is called consistently in all log statements that include DSN information, not just in `NewFromDSN`.

### Security Documentation ❌

**SB Approach:**
- Dedicated security guide with best practices
- Clear examples of safe vs unsafe patterns
- SQL injection protection documentation

**Current Neat state:** There is no dedicated security documentation. The docs directory contains comparison docs, API reference, associations, and soft-delete proposals, but nothing on security.

**Lesson for Neat:**
- Add a `docs/security.md` covering: parameterized queries (already done), DSN redaction in logs, SQL injection prevention when using `Raw()`, safe handling of user-supplied table/column names
- Document that `q.Where("column = ?", value)` is safe, and `q.Raw("SELECT * FROM " + userInput)` is not
- Document that `redactDSN` is used in log output and link to the implementation

## 3. Focused Architecture

### Scope Management ⚠️

**SB Approach:**
- Focused scope with clear boundaries
- Single-purpose library for SQL generation
- Minimal dependencies (8 direct dependencies)

**Current Neat state:** Neat is an ORM, not a SQL builder, so broader scope is intentional and correct. However, Neat's `go.mod` should be audited for unnecessary dependencies. The key discipline is: new features should not require new dependencies unless absolutely necessary.

**Lesson for Neat:**
- Run `go mod tidy` and `go mod why <package>` periodically to prune unused or indirect dependencies
- For each new feature proposal, explicitly evaluate whether it can be implemented without adding a new module dependency
- Document the rationale for each significant direct dependency in a comment or in the README

### Separation of Concerns ✅

**SB Approach:**
- Clear separation between SQL generation and execution
- `schema/` sub-package for execution functions
- Builder methods for SQL generation only

**Current Neat state:** Neat already has well-separated packages:
- `database/query/` — query execution
- `database/schema/` — schema operations
- `database/migration/` — migrations
- `database/driver/` — driver abstraction
- `database/association/` — relationships
- `contracts/database/orm/` — interfaces only

This is good architecture. The interface-in-`contracts/`, implementation-in-`database/` pattern cleanly separates concerns.

**One gap:** The `database/cursor/` package couples SQL generation with execution for cursor-based pagination. Verify it uses the same driver interface rather than building raw SQL directly.

## 4. Production Readiness

### Migration Strategy ⚠️

**SB Approach:**
- Clear migration guides for breaking changes
- Well-documented deprecation timelines
- Examples of before/after code patterns
- Multiple migration strategies (quick, gradual)

**Current Neat state:** Neat has a `CHANGELOG.md` but no migration guides. The `docs/proposals/` directory contains implementation plans but not upgrade guides for users of Neat itself.

**Lesson for Neat:**
- Establish a versioning policy: document which package version introduced which breaking change
- For any change to the `contracts/database/orm/` interfaces (which is the public API surface), provide a concrete migration example
- Add a `BREAKING_CHANGES.md` or section in `CHANGELOG.md` specifically for interface changes
- Note: most of Neat's recent development has been additive — this becomes critical when the first major API refactor happens

### Architectural Decision Records ❌

**Observation:** Neat's `docs/proposals/` directory contains good proposal documents but lacks records of *why* key architectural decisions were made. This creates maintenance risk — future contributors don't know why the callback `Transaction(func(tx Query) error)` pattern was chosen over explicit Begin/Commit, or why the `contracts/` package exists separately from implementations.

**Lesson for Neat:**
- Add a lightweight `docs/decisions/` folder with short ADRs (Architecture Decision Records)
- Each ADR covers: context, decision, rationale, consequences
- Start with the most non-obvious decisions: callback transactions, interface-in-contracts pattern, driver registration system
- ADRs do not need to be exhaustive — 3-4 sentences per decision is sufficient

## 5. Code Quality Practices

### Standard Documentation ⚠️

**SB Approach:**
- Standard Go godoc format consistently
- Practical examples in method documentation
- Database-specific behavior documentation
- Parameter explanations with usage notes

**Current Neat state:** Documentation quality is uneven across packages:
- `database/driver/driver.go` — excellent interface documentation
- `errors/errors.go` — good type and variable documentation
- `database/db.go` — methods documented but no `Example*` functions
- `database/query/` — many public methods lack godoc entirely
- `contracts/database/orm/` — interface methods have no documentation

**Lesson for Neat:**
- Prioritize documenting the `contracts/database/orm/` interface methods — these are the primary API surface that users program against
- Add `Example*` test functions in `database/db_test.go` for `New`, `NewFromDSN`, `Transaction`
- The `database/association/` package has the least documentation relative to its complexity — prioritize that

### Testing Strategy ⚠️

**SB Approach:**
- Focused test coverage with 32+ tests for advanced features
- Clear test patterns for dialect-specific behavior
- Integration tests for multiple databases
- All tests passing with comprehensive coverage

**Current Neat state:** Neat has strong unit test coverage in `database/query/` (35+ error scenario tests, scope tests, association tests, cursor tests). Integration tests exist for MySQL and Oracle in `integration_tests/`. 

**Three specific gaps:**
1. Six tests use `defer recover()` for expected panics — these should be replaced with nil-DB guards (see Section 1)
2. The `database/db/config_builder.go` `BuildDSN()` function has no unit tests — this is DSN generation for 6 drivers with no coverage
3. Pool configuration logic has no tests — there is no verification that `NewFromDSN("sqlite://...")` actually uses single-connection pool settings

**Lesson for Neat:**
- Add a `config_builder_dsn_test.go` covering all 6 `BuildDSN()` paths
- Add pool-default tests once driver-aware defaults are implemented (see `lessons-from-database-package.md`, Priority 1)

## 6. License Considerations

### Permissive Licensing ⚠️

**SB Approach:**
- MIT license for maximum compatibility
- Suitable for commercial production use
- Minimal restrictions on usage and distribution

**Current Neat state:** Neat uses a dual-license model (open-source + commercial). This is a valid business decision and not a technical anti-pattern. The implication is that commercial users must obtain a separate license, which means the adoption curve may be slower than MIT-licensed alternatives.

**Lesson for Neat:**
- Ensure `LICENSE` and `LICENSE_COMMERCIAL.txt` are clear about which use cases require a commercial license
- Add a concise "License" section to the README that answers: "Can I use Neat in a commercial closed-source product without paying?"
- This is a business decision, not a code quality issue — but clarity in the documentation reduces friction

## 7. API Design

### Fluent API Preservation ✅

**SB Approach:**
- Maintains fluent API while adding comprehensive error handling
- Error collection doesn't break method chaining
- Methods return `BuilderInterface` for chaining
- Errors validated at build time

**Current Neat state:** Neat uses a different but equally valid pattern: `Where`, `OrderBy`, `Limit`, `Select`, `Scopes` all return `*Query` for chaining, and errors are returned only at terminal methods (`Find`, `First`, `Create`, etc.). This is the standard ORM pattern used by GORM, Ent, and others. It does not need SB's error-collection approach.

**No action required.** The fluent API is already well-designed for an ORM context.

### Dialect-Specific Optimizations ⚠️

**SB Approach:**
- GIN indexes for PostgreSQL, FULLTEXT for MySQL
- Partial indexes, covering indexes with proper dialect support
- Clear documentation of dialect capabilities

**Current Neat state:** Neat's `database/schema/` supports column operations per dialect, but there is no documented feature matrix showing which schema features are supported per driver. Looking at `database/driver/` there are driver-specific implementations, but without reading each file it is not clear which features fall back silently vs. which return errors.

**Lesson for Neat:**
- Add a driver feature matrix table to `docs/driver-registration.md` or a new `docs/driver-capabilities.md`
- Columns: driver | transactions | savepoints | JSON operators | full-text search | partial indexes | read replicas
- Unsupported features should return a descriptive error (e.g. `ErrUnsupportedFeature`) rather than silently executing a no-op or panicking

## 8. Implementation Priority

### Core Feature Completeness ⚠️

**SB Approach:**
- Focused on completing core features before expanding scope
- Subqueries, JOINs, indexes fully implemented and tested
- Clear success criteria for each feature
- Comprehensive testing before moving to next feature

**Current Neat state:** Neat has broad feature coverage (associations, soft deletes, scopes, cursors, migrations, seeding) but the depth of test coverage is uneven. The `integration_tests/common/` directory has helpers for creates, finds, aggregates, and chunks, suggesting good coverage for read operations, but write-operation integration coverage is less visible.

**Lesson for Neat:**
- Before adding new features, audit test coverage for existing ones
- The `database/association/` package is complex and high-risk — verify integration test coverage for all association types (belongs_to, has_one, has_many, many_to_many)
- Avoid adding experimental features (e.g. proposals in `docs/proposals/`) to the stable API until they have integration test coverage

## Implementation Roadmap

### Priority 1: Nil Safety (High Impact, Low Risk)
1. Add nil `*sql.DB` guard at the start of every terminal query method
2. Return `ErrNilDatabase` instead of panicking
3. Remove six `defer recover()` blocks from `query_errors_test.go`
4. Replace them with `if err == nil { t.Error("expected ErrNilDatabase") }` assertions

### Priority 2: Security Documentation (Medium Impact, Low Effort)
1. Create `docs/security.md` covering parameterized queries, DSN redaction, Raw() risks
2. Audit all `db.go` log calls to confirm `redactDSN` is used consistently
3. Add security notice to `Where` vs `Raw` comparison in docs

### Priority 3: Test Coverage Gaps (Medium Impact, Medium Effort)
1. Add `config_builder_dsn_test.go` covering all 6 `BuildDSN()` driver paths
2. Add unit tests for `parseDSN` edge cases (empty, too long, malformed URLs)
3. Add association integration tests for all four relationship types

### Priority 4: Documentation Completeness (Low Risk, Medium Effort)
1. Document all `contracts/database/orm/` interface methods
2. Add `Example*` functions for primary constructors
3. Add `docs/driver-capabilities.md` feature matrix
4. Add `docs/decisions/` with 3-4 ADRs for key architectural choices
5. Clarify license requirements in README

## What Neat Already Does Better Than SB

| Feature | SB SQL Builder | Neat ORM |
|---------|---------------|----------|
| Error types | `BuilderError` struct | `StructuredError` with Module field |
| Parameterized queries | Opt-in initially | Always on |
| DSN security | Not applicable | `redactDSN()` in log output |
| Architecture | Single package | Well-separated sub-packages |
| Scope | SQL generation only | Full ORM with associations |
| Context support | Not built-in | Full `context.Context` propagation |
| Transaction safety | Manual Begin/Commit | Callback pattern prevents leaks |

## Success Metrics

- **Nil safety**: Zero `defer recover()` blocks in all test files
- **Security doc**: `docs/security.md` exists and covers the four key topics
- **DSN coverage**: All 6 `BuildDSN()` paths have unit tests
- **Interface docs**: All `contracts/database/orm/` methods have godoc
- **Feature matrix**: Driver capabilities documented in one place

## Conclusion

SB SQL Builder demonstrates how to build a focused, production-ready library with sophisticated error handling and systematic documentation. The most applicable lessons for Neat are not architectural — Neat's architecture is already solid — but operational: fixing the nil-DB panic surface, adding the security documentation that is clearly missing, and closing the test coverage gap in DSN building and associations.

Neat is already ahead of SB in structured error types, parameterized query safety, context propagation, and architectural separation. The gap is primarily in documentation completeness and a small number of nil-safety defects.
