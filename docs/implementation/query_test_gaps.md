# Test Coverage Gaps: database/query Package

**Date:** 2026-05-25  
**Package:** `database/query`  
**Overall Coverage:** ~60%

## Executive Summary

The test suite for the `database/query` package has moderate coverage with good breadth across public methods but limited depth in integration scenarios and edge cases. While most API surface is covered, critical integration paths and error handling scenarios lack comprehensive testing.

## Coverage by Category

### Well-Covered Areas ✅

#### Utility Functions (`utils_test.go`)
- String conversion (toCamelCase, camelToSnake)
- Primary key handling (getPrimaryKeyValue, setModelPrimaryKey)
- Struct field mapping (structFieldColumnName, getColumnToIndexPath)
- Scan operations (structScanDests, copyScanResults)

#### Transaction Hooks (`transaction_hooks_test.go`)
- BeforeCommit/AfterCommit callbacks
- BeforeRollback/AfterRollback callbacks
- Hook error handling
- Nested transaction commit/rollback

#### WHERE Clause Methods (`query_where_test.go`)
- All 25+ WHERE variants tested (WhereIn, WhereNotIn, WhereBetween, etc.)
- JSON WHERE clauses (WhereJsonContains, WhereJsonContainsKey, etc.)
- Column comparisons (WhereColumn, WhereAny, WhereAll, WhereNone)

#### Soft Delete (`query_soft_delete_test.go`)
- Filter injection logic for soft-deletable models
- WithTrashed/OnlyTrashed/WithoutTrashed methods
- Pointer field handling

#### Query Builder Methods (`query_builder_test.go`)
- Basic method chaining (Select, Join, Group, OrderBy, Limit, etc.)
- All builder methods return non-nil Query instances

#### Scan Operations (`query_scan_test.go`)
- Tag priority (db > neat > gorm > snake_case)
- Slice scanning
- Unmatched column handling
- Embedded struct field mapping

#### CRUD Operations (`query_crud_test.go`)
- InsertGetId dialect differences (Postgres RETURNING vs MySQL)
- ToSql generation for Create, Delete, First, Get, Update

#### Connection Management (`query_accessors_test.go`)
- Read/write connection routing
- Query logging (enable/disable/flush)
- Slow query threshold warnings
- Replica connection handling
- Clone propagation of replicas

#### Aggregate Methods (`query_aggregate_test.go`)
- SQL generation for Count, Pluck, Value, Avg, Max, Min, Sum

---

## Significant Gaps ❌

### Missing Integration Tests

#### 1. Nested Transactions and Savepoints ✅
**File:** `query_transaction.go`  
**Status:** COMPLETED (2026-05-25)
**Tests Added:**
- `TestSavePoint` - Basic savepoint creation
- `TestSavePointNotInTransaction` - Error handling when not in transaction
- `TestRollbackTo` - Rollback to a specific savepoint
- `TestRollbackToNotInTransaction` - Error handling when not in transaction
- `TestSavepointCreationAndRollback` - Full savepoint creation and rollback scenario
- `TestNestedSavepointLevels` - Nested savepoint levels with rollback
- `TestSavepointErrorHandling` - Invalid savepoint name error handling
- `TestRollbackToInvalidSavepoint` - Nonexistent savepoint error handling
- `TestBeginCreatesSavepointForNestedTransaction` - Nested transaction savepoint creation
- `TestCommitReleasesSavepoint` - Savepoint release on commit
- `TestRollbackReleasesSavepoint` - Savepoint release on rollback

**Impact:** Critical for complex transaction workflows

#### 2. Raw SQL Execution
**File:** `query_advanced.go`  
**Missing Tests:**
- `Raw()` method with subquery callbacks
- `Exec()` method execution
- Raw SQL with parameter binding
- Raw SQL error handling

**Impact:** High - Raw SQL is commonly used for complex queries

**Status:** ✅ Completed
- Added `TestRawWithSimpleSQL`, `TestRawWithMultipleParameters`, `TestRawWithoutParameters`
- Added `TestExecMethod`, `TestExecWithParameterBinding`, `TestExecWithMultipleParameters`
- Added `TestExecInTransaction`, `TestExecWithUpdate`, `TestExecWithDelete`
- Fixed nil pointer bug in `Exec()` implementation
- Note: `Raw()` only accepts string SQL, not callbacks (callbacks are supported by `Select()`)

#### 3. Relation Loading (Eager Loading)
**File:** `query_relations.go`  
**Missing Tests:**
- Actual eager loading execution (only initialization tested)
- `loadRelations()` method
- Relation constraint callbacks
- Foreign key resolution
- Recursive relation loading prevention
- Relation loading with different model types

**Impact:** High - Eager loading is a core ORM feature

#### 4. Cursor Streaming
**File:** `query_scan.go`  
**Missing Tests:**
- `Cursor()` method
- Cursor channel creation and consumption
- Cursor error handling
- Cursor with transactions
- Cursor with WHERE clauses

**Impact:** Medium - Important for large dataset processing

#### 5. Chunk Processing
**File:** `query_scan.go`  
**Missing Tests:**
- `Chunk()` method
- Chunk callback execution
- Chunk size variations
- Chunk with typed slices
- Chunk error handling

**Impact:** Medium - Important for memory-efficient processing

#### 6. UpdateOrInsert Pattern
**File:** `query.go`  
**Missing Tests:**
- `UpdateOrInsert()` method
- Map and struct attributes
- Merge logic for insert
- Error scenarios

**Impact:** High - Common upsert pattern

#### 7. FirstOr/FirstOrCreate/FirstOrNew Patterns
**File:** `query_scan.go`  
**Missing Tests:**
- `FirstOr()` with callback
- `FirstOrCreate()` execution
- `FirstOrNew()` preparation
- Error handling for not found scenarios

**Impact:** High - Common patterns in application code

#### 8. UpdateOrCreate Pattern
**File:** `query_scan.go`  
**Missing Tests:**
- `UpdateOrCreate()` method
- Update vs create logic
- Attribute matching
- Error scenarios

**Impact:** High - Common upsert pattern

#### 9. Pagination
**File:** `query_scan.go`  
**Missing Tests:**
- `Paginate()` method
- Total count calculation
- Offset calculation
- Paginate with WHERE clauses
- Paginate error handling

**Impact:** High - Essential for web applications

#### 10. Scopes
**File:** `query.go`  
**Missing Tests:**
- `Scopes()` method
- Scope application order
- Scope with query chaining
- Scope error handling

**Impact:** Medium - Important for query reusability

#### 11. Context Handling
**File:** `query_accessors.go`  
**Missing Tests:**
- `WithContext()` method
- Context propagation
- Context cancellation
- Context with transactions

**Impact:** Medium - Important for request-scoped queries

#### 12. Observer Registration
**File:** `query_accessors.go`  
**Missing Tests:**
- `Observe()` method
- Observer registration
- Observer dispatch during CRUD operations
- Multiple observers

**Impact:** Medium - Important for event-driven architecture

#### 13. Distinct with Columns
**File:** `query_builder.go`  
**Missing Tests:**
- `Distinct(args...)` with column arguments
- Distinct with aggregate functions
- Distinct SQL generation verification

**Impact:** Low - Edge case but should be tested

---

### Limited Depth Issues

#### 1. WHERE Clause Tests
**File:** `query_where_test.go`  
**Issue:** Tests only check non-nil return, not SQL generation or execution  
**Missing:**
- SQL output verification
- Argument binding verification
- Dialect-specific SQL differences
- Complex WHERE clause combinations

**Impact:** Medium - WHERE clauses are critical for query correctness

#### 2. Builder Tests
**File:** `query_builder_test.go`  
**Issue:** Only basic chaining tested, no SQL output verification  
**Missing:**
- Actual SQL generation verification
- Argument order verification
- Complex query building scenarios
- Error cases in building

**Impact:** High - Builder is core to query generation

#### 3. Transaction Tests
**File:** `transaction_hooks_test.go`  
**Issue:** No nested transaction scenarios beyond basic commit/rollback  
**Missing:**
- Nested savepoint scenarios
- Transaction isolation levels
- Transaction timeout handling
- Deadlock scenarios

**Impact:** Medium - Important for complex workflows

#### 4. Soft Delete Tests
**File:** `query_soft_delete_test.go`  
**Issue:** No end-to-end delete/restore tests  
**Missing:**
- Actual DELETE execution with soft delete
- Restore execution
- ForceDelete execution
- Soft delete with relations

**Impact:** High - Soft delete is a core feature

#### 5. JSON WHERE Clauses
**File:** `query_where_test.go`  
**Issue:** No dialect-specific SQL verification  
**Missing:**
- SQLite json_extract() SQL verification
- MySQL/Postgres JSON_CONTAINS() SQL verification
- JSON path handling
- JSON operator differences

**Impact:** Medium - Important for cross-database compatibility

#### 6. Lock Clauses
**File:** `query_advanced.go`  
**Issue:** No tests for LockForUpdate/SharedLock SQL generation  
**Missing:**
- `LockForUpdate()` SQL generation
- `SharedLock()` SQL generation
- Lock clause with WHERE clauses
- Dialect-specific lock syntax

**Impact:** Medium - Important for concurrent access control

---

### Missing Edge Cases

#### 1. Error Handling
**Missing:**
- Database connection errors
- Query execution errors
- Transaction errors
- Scan errors with mismatched types
- Timeout errors
- Constraint violation errors

**Impact:** High - Error handling is critical for production

#### 2. Nil/Zero Values
**Missing:**
- Nil model handling
- Zero value primary keys
- Nil pointer fields
- Empty slices in bulk operations
- Empty WHERE clauses

**Impact:** Medium - Edge cases that cause panics

#### 3. Complex Types
**Missing:**
- Nested struct fields
- Embedded struct fields
- Pointer to pointer fields
- Custom types implementing Scanner/Valuer
- JSON/JSONB fields
- Array fields

**Impact:** Medium - Common in real-world models

#### 4. Dialect Differences
**Missing:**
- MySQL-specific tests (LIMIT/OFFSET syntax, quoting)
- PostgreSQL-specific tests (RETURNING, array types)
- SQL Server-specific tests
- Turso-specific tests
- Cross-dialect compatibility tests

**Impact:** High - Package supports multiple databases

#### 5. Bulk Operations
**Missing:**
- Bulk insert with many records
- Bulk update scenarios
- Bulk delete scenarios
- Bulk operation error handling
- Bulk operation performance

**Impact:** Medium - Important for data migration

#### 6. Concurrent Operations
**Missing:**
- Concurrent query execution
- Concurrent transaction handling
- Thread-safety of Query cloning
- Race condition tests

**Impact:** Low - Important for high-concurrency applications

---

## Recommendations

### High Priority (Critical for Production)

1. **Add integration tests for nested transactions and savepoints**
   - Test SavePoint() and RollbackTo() methods
   - Test nested savepoint levels
   - Test savepoint error scenarios

2. **Add end-to-end tests for relation loading**
   - Test actual eager loading execution
   - Test relation constraint callbacks
   - Test foreign key resolution
   - Test with different model types

3. **Add tests for Cursor() and Chunk() streaming methods**
   - Test cursor channel creation and consumption
   - Test chunk callback execution
   - Test error handling

4. **Add tests for UpdateOrInsert, FirstOrCreate, UpdateOrCreate patterns**
   - Test upsert logic
   - Test attribute matching
   - Test error scenarios

5. **Add SQL generation verification for WHERE clause methods**
   - Verify actual SQL output
   - Verify argument binding
   - Test dialect-specific differences

6. **Add tests for Raw() and Exec() methods**
   - Test raw SQL execution
   - Test parameter binding
   - Test error handling

7. **Add comprehensive error handling tests**
   - Test database connection errors
   - Test query execution errors
   - Test transaction errors
   - Test scan errors

### Medium Priority (Important for Robustness)

8. **Add Paginate() tests with total count verification**
   - Test offset calculation
   - Test total count accuracy
   - Test with WHERE clauses

9. **Add Scopes() tests**
   - Test scope application
   - Test scope chaining
   - Test scope error handling

10. **Add WithContext() tests**
    - Test context propagation
    - Test context cancellation
    - Test with transactions

11. **Add Distinct with columns tests**
    - Test Distinct(args...)
    - Test with aggregates
    - Verify SQL generation

12. **Add LockForUpdate/SharedLock SQL generation tests**
    - Test lock clause generation
    - Test with WHERE clauses
    - Test dialect differences

13. **Add MySQL/Postgres dialect-specific tests**
    - Test MySQL-specific syntax
    - Test PostgreSQL-specific syntax
    - Test cross-dialect compatibility

14. **Add observer registration tests**
    - Test Observe() method
    - Test observer dispatch
    - Test multiple observers

15. **Add soft delete end-to-end tests**
    - Test actual DELETE with soft delete
    - Test Restore execution
    - Test ForceDelete execution

### Low Priority (Nice to Have)

16. **Add more edge case tests**
    - Test nil/zero value handling
    - Test complex nested types
    - Test custom Scanner/Valuer types

17. **Add bulk operation tests**
    - Test bulk insert with many records
    - Test bulk update scenarios
    - Test performance characteristics

18. **Add concurrency tests**
    - Test concurrent query execution
    - Test thread-safety
    - Test race conditions

19. **Add performance regression tests**
    - Beyond existing benchmarks
    - Test query performance with large datasets
    - Test memory usage patterns

---

## Test File Organization

### Current Test Files
- `utils_test.go` - Utility functions
- `transaction_hooks_test.go` - Transaction lifecycle hooks
- `to_sql_test.go` - ToSql interface
- `query_where_test.go` - WHERE clause methods
- `query_transaction_test.go` - Transaction methods
- `query_soft_delete_test.go` - Soft delete methods
- `query_scan_test.go` - Scan operations
- `query_relations_test.go` - Relation methods
- `query_crud_test.go` - CRUD operations
- `query_builder_test.go` - Query builder methods
- `query_bench_test.go` - Benchmarks
- `query_aggregate_test.go` - Aggregate methods
- `query_advanced_test.go` - Advanced query methods
- `query_accessors_test.go` - Accessor methods
- `helpers_test.go` - Test helpers
- `export_test.go` - Test exports
- `clause_test.go` - Clause builders
- `builder_test.go` - SQL builder

### Suggested New Test Files
- `query_transaction_integration_test.go` - Nested transactions, savepoints
- `query_raw_test.go` - Raw SQL execution
- `query_relations_integration_test.go` - Eager loading integration
- `query_streaming_test.go` - Cursor and Chunk methods
- `query_upsert_test.go` - UpdateOrInsert, FirstOrCreate, UpdateOrCreate
- `query_pagination_test.go` - Paginate method
- `query_scopes_test.go` - Scopes method
- `query_context_test.go` - WithContext method
- `query_observers_test.go` - Observer registration
- `query_errors_test.go` - Error handling scenarios
- `query_dialect_test.go` - Dialect-specific tests
- `query_edge_cases_test.go` - Edge case scenarios
- `query_bulk_test.go` - Bulk operations
- `query_concurrency_test.go` - Concurrent operations

---

## Metrics

### Current Test Count
- Total test files: 19
- Total test functions: ~100
- Test lines of code: ~2,500

### Estimated Test Count After Coverage
- Total test files: 33 (+14)
- Total test functions: ~250 (+150)
- Test lines of code: ~8,000 (+5,500)

### Coverage Target
- Current: ~60%
- Target: ~85%
- Gap: +25%

---

## Conclusion

The `database/query` package has a solid foundation with good coverage of basic functionality. However, critical integration paths, error handling, and edge cases need additional testing to ensure production readiness. The recommended tests should be prioritized based on the likelihood of usage and potential impact of failures.

The test suite would benefit from better organization, with integration tests separated from unit tests, and more comprehensive end-to-end scenarios that verify actual database behavior rather than just method chaining.
