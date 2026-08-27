# Benchmark Suite

**Date**: June 1, 2026
**Last Reviewed**: August 27, 2026
**Status**: Done
**Priority**: Medium
**Impact**: Medium
**Effort**: Low
**ROI**: High

## Description

Create comprehensive benchmark suite.

## Rationale

- Performance comparisons with GORM
- Database-specific performance tests
- Memory usage profiling
- Regression detection

## Implemented

- `database/query/query_bench_test.go` with 6 benchmark functions
- `database/query/query_performance_test.go` with 15 benchmark functions covering CRUD and query building
- `benchmarks/gorm_comparison_test.go` with 12 GORM vs neat comparison benchmarks (Where chain, Select, OrderBy+Limit+Offset, complex query, ToSql, fresh-build-per-iteration)
- `integration_tests/benchmarks/engine_benchmarks_test.go` with 20 per-engine benchmarks covering MySQL, PostgreSQL, SQLite, and SQL Server (Select, Where, Insert, Update, ToSql per engine)
- `.github/workflows/benchmarks.yml` CI workflow for continuous benchmarking with:
  - Unit benchmarks on every push/PR (`./database/query/...` and `./benchmarks/...`)
  - Integration benchmarks matrix for MySQL, PostgreSQL, and SQLite
  - `benchstat` comparison against cached baselines
  - Regression detection (warns on >=10% slowdown)
  - Automatic baseline updates on main branch pushes
  - Manual baseline refresh via `workflow_dispatch`
  - 90-day artifact retention for trend analysis
- Integration test suite expanded to cover 10 database engines (MySQL, PostgreSQL, SQLite, SQL Server, Oracle, Turso, CockroachDB, TiDB, CSVDB, JSONDB, XMLDB, GODB)
- Embedded filesystem integration tests for CSVDB, JSONDB, XMLDB

## Remaining Work

- Connection pooling efficiency benchmarks (requires connection-pool-aware test harness)
- Eager loading performance benchmarks (requires multi-table benchmark setup)
- Transaction overhead benchmarks (requires nested-savepoint benchmark scenarios)
- Oracle, TiDB, CockroachDB per-engine benchmarks (CI services for these engines can be added to the matrix)

## Proposed Benchmarks

- ~~Query building overhead~~ — done (GORM comparison + per-engine ToSql)
- Connection pooling efficiency
- Eager loading performance
- Transaction overhead
- ~~Memory allocation patterns~~ — done (b.ReportAllocs on all benchmarks)

## Implementation Details

- Go benchmark framework (`testing.B` with `b.Loop`, `b.ReportAllocs`, `b.ReportMetric`)
- Continuous benchmarking via GitHub Actions (`benchmarks.yml`)
- Performance regression detection via `benchstat` baseline comparison
- Database-specific benchmarks via integration test suite with build tag
- GORM comparison uses DryRun mode (MySQL dialector, no real DB connection needed)
- Per-engine benchmarks use real database connections with 1000-row seeded tables

## References

- Extracted from `docs/proposals/feature-requests.md` (item #8)
- GORM comparison: `benchmarks/gorm_comparison_test.go`
- Per-engine benchmarks: `integration_tests/benchmarks/engine_benchmarks_test.go`
- CI workflow: `.github/workflows/benchmarks.yml`
