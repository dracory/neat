# Lessons from SB SQL Builder

This document outlines key lessons and best practices that Neat ORM can learn from the SB SQL Builder project, based on a direct comparison of both codebases.

> Status markers: ✅ Already done in Neat | ⚠️ Partial | ❌ Not yet implemented

## Overview

SB SQL Builder demonstrates excellent practices in building a focused, production-ready library with sophisticated error handling, security-first design, and systematic documentation. These lessons help identify where Neat ORM is already strong, and where concrete improvements should be made.

## 1. Code Quality Practices

### Standard Documentation ✅

**SB Approach:**
- Standard Go godoc format consistently
- Practical examples in method documentation
- Database-specific behavior documentation
- Parameter explanations with usage notes

**Current Neat state:** Documentation has been significantly improved:
- `contracts/database/orm/orm.go` — Added comprehensive godoc to 10 most commonly used Query methods (Find, First, Create, Update, Delete, Where, Select, Table, Transaction, Count) with parameter descriptions and usage examples
- `database/db_test.go` — Example functions already exist for New, NewFromDSN, and Transaction (multiple variants each)
- `database/driver/driver.go` — excellent interface documentation
- `errors/errors.go` — good type and variable documentation

**Implementation:** Added godoc documentation to `contracts/database/orm/orm.go` for:
- **Find** - Retrieving records with conditions
- **First** - Retrieving first matching record
- **Create** - Inserting new records
- **Update** - Updating records with column/values
- **Delete** - Deleting records with soft delete support
- **Where** - Adding WHERE clauses
- **Select** - Specifying columns to retrieve
- **Table** - Specifying table name
- **Transaction** - Executing functions within transactions
- **Count** - Counting matching records

Each method now includes parameter descriptions, return value explanations, and practical code examples.

**No action required.** This is already implemented.

## 2. Testing Strategy ⚠️

**SB Approach:**
- Focused test coverage with 32+ tests for advanced features
- Clear test patterns for dialect-specific behavior
- Integration tests for multiple databases
- All tests passing with comprehensive coverage

**Current Neat state:** Neat has strong unit test coverage in `database/query/` (35+ error scenario tests, scope tests, association tests, cursor tests). Integration tests exist for MySQL and Oracle in `integration_tests/`. 

**Three specific gaps:**
1. ✅ Six tests use `defer recover()` for expected panics — these should be replaced with nil-DB guards (see Section 1) - **COMPLETED**: The `validate()` helper now checks for nil DB and empty table, and all terminal methods call it before execution.
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
1. ✅ Add nil `*sql.DB` guard at the start of every terminal query method - **COMPLETED**
2. ✅ Return `ErrNilDatabase` instead of panicking - **COMPLETED**
3. ✅ Remove six `defer recover()` blocks from `query_errors_test.go` - **COMPLETED**
4. ✅ Replace them with `if err == nil { t.Error("expected ErrNilDatabase") }` assertions - **COMPLETED**

### Priority 2: Security Documentation (Medium Impact, Low Effort)
1. ✅ Create `docs/security.md` covering parameterized queries, DSN redaction, Raw() risks - **COMPLETED**
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
