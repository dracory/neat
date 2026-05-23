# Neat Implementation Plan

**Date:** May 21, 2026
**Last reviewed:** May 23, 2026
**Status**: ✅ Implementation Complete - Historical Planning Document

## Overview

This document provides the detailed implementation plan for building **neat**, a Go ORM and database abstraction layer that provides the same functionality as eloquent but without GORM dependency. Neat was implemented from scratch using native `database/sql` with a custom query builder and SQL generation.

**Note**: This is a historical planning document. All 14 phases have been completed. For current implementation gaps and remaining work, see `gaps.md`.

## Project Goals

1. **Feature Parity**: Implement all features currently available in eloquent
2. **No GORM Dependency**: Use native `database/sql` and custom SQL builder
3. **Clean Architecture**: Maintain the same contract-based interface design
4. **Performance**: Optimize for better performance than GORM-based implementation
5. **Maintainability**: Simplify codebase by removing GORM abstraction layer

## Eloquent Feature Analysis

Based on the current eloquent implementation, neat must implement:

### Core ORM Features (contracts/database/orm/orm.go)
- **80+ Query Methods**: Where, Select, Order, Limit, Offset, Join, Group, Having, etc.
- **CRUD Operations**: Create, Read, Update, Delete with various methods
- **Advanced Query**: WhereIn, WhereBetween, WhereColumn, WhereExists, WhereJson, etc.
- **Aggregations**: Count, Sum, Avg, Min, Max, Exists
- **Result Retrieval**: Get, First, Find, Pluck, Value, Scan, Chunk, Paginate, Cursor
- **Transactions**: Begin, Commit, Rollback, Transaction with callbacks and savepoints
- **Model Lifecycle**: Observer pattern with Creating, Created, Updating, Updated, etc.
- **Soft Deletes**: WithTrashed, OnlyTrashed, WithoutTrashed, Restore, ForceDelete
- **Associations**: Belongs-to, Has-many, Has-one with eager and lazy loading
- **Query Logging**: EnableQueryLog, DisableQueryLog, GetQueryLog, FlushQueryLog
- **ToSql Interface**: Generate SQL without execution for debugging

### Schema Builder (contracts/database/schema/)
- **Blueprint Interface**: 50+ methods for table/column definition
- **Column Types**: Integer, BigInteger, String, Text, Boolean, Json, Decimal, etc.
- **Indexes**: Primary, Unique, Index, FullText
- **Foreign Keys**: Foreign constraints with onUpdate/onDelete
- **Table Operations**: Create, Drop, Rename, Modify
- **Column Operations**: Add, Drop, Rename, Modify
- **Database-specific Grammars**: MySQL, PostgreSQL, SQLite, SQL Server
- **Database-specific Processors**: Type conversion and processing

### Configuration System (config.go)
- **DBConfig**: Default connection, connection map, pool config
- **ConnectionConfig**: Driver, host, port, database, username, password, SSL, etc.
- **PoolConfig**: MaxIdleConns, MaxOpenConns, ConnMaxLifetime, ConnMaxIdleTime
- **DSN Parsing**: Support postgres://, mysql://, sqlite://, turso:// formats

### Support Utilities (support/)
- **collect**: Collection utilities for data manipulation
- **convert**: Type conversion utilities
- **database**: Database-specific utilities
- **env**: Environment variable handling
- **str**: String manipulation utilities

### Event System (event.go)
- **EventBus**: Event dispatching and handling
- **Event Types**: Model lifecycle events
- **Observer Registration**: Register observers for models

### Factory Pattern
- **Factory Interface**: Create models with default attributes
- **Count**: Set number of models to generate
- **Create**: Create and persist models
- **Make**: Create models without persistence
- **CreateQuietly**: Create without firing events

## Architecture Design

### Package Structure

```
neat/
├── config.go                 # Configuration structures and parsing
├── db.go                     # Main Database entry point
├── event.go                  # Event bus implementation
├── contracts/                # Interface definitions (copied from eloquent)
│   ├── config/
│   ├── database/
│   │   ├── orm/
│   │   ├── schema/
│   │   ├── factory/
│   │   └── migration/
│   ├── errors/
│   ├── factory/
│   ├── log/
│   ├── migration/
│   ├── schema/
│   └── seeder/
├── database/
│   ├── db/                   # Database connection management
│   │   ├── config_builder.go
│   │   ├── dsn.go
│   │   └── expression.go
│   ├── orm/                  # ORM abstraction layer
│   │   ├── orm.go           # Orm implementation
│   │   ├── model.go         # Model types (Model, SoftDeletes, Timestamps)
│   │   └── factory.go       # Factory implementation
│   ├── query/                # Query builder implementation
│   │   ├── query.go         # Main Query struct
│   │   ├── builder.go       # SQL builder
│   │   ├── clause.go        # Query clauses
│   │   ├── where.go         # Where clause handling
│   │   ├── join.go          # Join handling
│   │   ├── aggregate.go     # Aggregation methods
│   │   └── to_sql.go        # ToSql interface implementation
│   ├── association/          # Association handling
│   │   ├── association.go   # Association interface
│   │   ├── belongs_to.go    # Belongs-to relationship
│   │   ├── has_many.go      # Has-many relationship
│   │   └── has_one.go       # Has-one relationship
│   ├── observer/             # Observer pattern
│   │   ├── observer.go      # Observer implementation
│   │   ├── event.go         # Event implementation
│   │   └── dispatcher.go    # Event dispatcher
│   ├── soft_delete/          # Soft delete implementation
│   │   └── soft_delete.go
│   ├── transaction/          # Transaction handling
│   │   ├── transaction.go   # Transaction implementation
│   │   ├── savepoint.go     # Savepoint handling
│   │   └── callback.go      # Transaction callbacks
│   ├── cursor/               # Cursor for streaming
│   │   └── cursor.go
│   ├── schema/               # Schema builder (reuse from eloquent)
│   │   ├── blueprint.go
│   │   ├── schema.go
│   │   ├── column.go
│   │   ├── index.go
│   │   ├── constants/
│   │   ├── grammars/
│   │   └── processors/
│   └── driver/               # Native database drivers
│       ├── driver.go        # Driver interface
│       ├── mysql.go         # MySQL driver
│       ├── postgres.go      # PostgreSQL driver
│       ├── sqlite.go        # SQLite driver
│       ├── sqlserver.go     # SQL Server driver
│       └── turso.go         # Turso driver
├── support/                  # Utility functions (copied from eloquent)
│   ├── collect/
│   ├── convert/
│   ├── database/
│   ├── env/
│   └── str/
├── errors/                   # Error definitions
│   └── errors.go
└── log/                      # Logger interface
    ├── log.go
    ├── std_logger.go
    └── noop_logger.go
```

### Key Design Decisions

#### 1. Native database/sql Usage
- Use `database/sql` directly for all database operations
- Implement custom connection pooling logic
- Use prepared statements for all queries
- Implement dialect-specific placeholder handling

#### 2. Custom SQL Builder
- Build SQL strings programmatically
- Support dialect-specific syntax
- Use parameterized queries to prevent SQL injection
- Implement clause-based query building (Where, Join, Group, etc.)

#### 3. Tag-based Model Mapping
- Use custom struct tags (e.g., `neat:"column:name;primaryKey"`)
- Support GORM tags for migration compatibility
- Implement tag parser for both neat and GORM tags
- Auto-detect table name from struct name

#### 4. Observer Pattern
- Implement observer pattern for model lifecycle events
- Support all event types: Creating, Created, Updating, Updated, etc.
- Allow multiple observers per model
- Execute events in correct order

#### 5. Association Loading
- Implement eager loading with `With()`
- Implement lazy loading with `Load()`
- Support belongs-to, has-many, has-one relationships
- Use separate queries for association loading (N+1 problem avoidance)

#### 6. Transaction Management
- Use `sql.Tx` for transactions
- Implement transaction callbacks
- Support nested transactions with savepoints
- Handle transaction rollback on errors

#### 7. Soft Deletes
- Use `deleted_at` timestamp column
- Exclude soft-deleted records by default
- Implement WithTrashed, OnlyTrashed, WithoutTrashed
- Support restore and force delete

## Implementation Phases

All 14 phases have been completed. This section provides a summary of what was implemented.

### Phase 1: Project Setup and Foundation ✅
Project structure, configuration system, database driver interface, native drivers (MySQL, PostgreSQL, SQLite, SQL Server, Turso), and connection management.

### Phase 2: ORM Abstraction Layer ✅
ORM implementation, model types with tag support, factory pattern, event system, and database entry point.

### Phase 3: Core Query Builder ✅
Query struct, CRUD operations, query building methods, advanced where clauses, JSON queries, SQL builder, and clause handling.

### Phase 4: Advanced Query Features ✅
Join operations, group by and having, aggregation methods, result retrieval, special query methods, and cursor implementation.

### Phase 5: Transaction Handling ✅
Basic transaction methods, transaction callbacks, and savepoint handling for nested transactions.

### Phase 6: Observer Pattern ✅
Observer implementation, event implementation, event dispatcher, and integration with query operations.

### Phase 7: Soft Deletes ✅
Soft delete implementation, soft delete query methods (WithTrashed, OnlyTrashed, WithoutTrashed), and integration with query builder.

### Phase 8: Associations ✅
Association interface, belongs-to, has-many, has-one relationships, eager/lazy loading, and association operations.

### Phase 9: ToSql Interface ✅
ToSql implementation for generating SQL without execution, ToRawSql method, and integration with query.

### Phase 10: Schema Builder Integration ✅
Schema builder integration from eloquent, schema methods, and database entry point integration.

### Phase 11: Testing Setup ✅
Unit tests, integration tests for MySQL/PostgreSQL/SQLite, and CI/CD setup with GitHub Actions.

### Phase 12: Documentation and Examples ✅
README, examples (basic-orm, advanced-queries, models, schema-builder, migrations, configuration), and documentation.

### Phase 13: Performance Optimization ✅
Benchmarking, optimization, and profiling.

### Phase 14: Final Polish and Release ✅
Code cleanup, dependency management, final testing, and release preparation.

---

## Timeline Summary

- **Phase 1**: Weeks 1-2 (Project Setup and Foundation)
- **Phase 2**: Weeks 3-4 (ORM Abstraction Layer)
- **Phase 3**: Weeks 5-8 (Core Query Builder)
- **Phase 4**: Weeks 9-11 (Advanced Query Features)
- **Phase 5**: Week 12 (Transaction Handling)
- **Phase 6**: Week 13 (Observer Pattern)
- **Phase 7**: Week 14 (Soft Deletes)
- **Phase 8**: Weeks 15-17 (Associations)
- **Phase 9**: Week 18 (ToSql Interface)
- **Phase 10**: Week 19 (Schema Builder Integration)
- **Phase 11**: Week 20 (Testing Setup)
- **Phase 12**: Week 21 (Documentation and Examples)
- **Phase 13**: Week 22 (Performance Optimization)
- **Phase 14**: Week 23 (Final Polish and Release)

**Total estimated timeline**: 23 weeks (5.75 months)

## Resource Requirements

### Team
- **Team size**: 1-2 developers
- **Skills required**:
  - Go development expertise
  - database/sql package knowledge
  - SQL dialect knowledge (MySQL, PostgreSQL, SQLite, SQL Server)
  - ORM design patterns
  - Testing and benchmarking

### Infrastructure
- **Testing**: Access to MySQL, PostgreSQL, SQLite, SQL Server for integration testing
- **CI/CD**: GitHub Actions for automated testing
- **Development**: Local development environments for all database types

### Tools
- Go toolchain (go test, go build, go mod)
- Database clients for all supported databases
- Benchmarking tools (go test -bench)
- Code coverage tools (go test -cover)
- Profiling tools (pprof)

## Success Criteria

1. **Feature Parity**: All features from eloquent are implemented
2. **No GORM Dependency**: Zero GORM dependencies in go.mod
3. **API Compatibility**: Same contract interface as eloquent
4. **Performance**: Performance improvement or parity with eloquent
5. **Test Coverage**: >80% code coverage
6. **Documentation**: Comprehensive documentation and examples
7. **Database Support**: All supported databases work correctly (SQLite, MySQL, PostgreSQL, SQL Server, Turso)

## Open Questions

1. **Tag Format**: Should we use `neat:` tags or a different format?
   - Option A: `neat:"column:name;primaryKey"`
   - Option B: `db:"column:name;primaryKey"`
   - Option C: `orm:"column:name;primaryKey"`
   - Recommendation: Option A for clear branding

2. **GORM Tag Support**: Should we support GORM tags for migration?
   - Option A: Yes, support GORM tags indefinitely
   - Option B: Yes, support with deprecation warning
   - Option C: No, only support neat tags
   - Recommendation: Option B for easier migration

3. **External Dependencies**: Should we use any external SQL builders?
   - Option A: Use sb as dependency
   - Option B: Use squirrel as dependency
   - Option C: Implement our own SQL builder
   - Recommendation: Option C for full control

4. **Version Strategy**: Should we version separately from eloquent?
   - Option A: Same version as eloquent
   - Option B: Independent versioning
   - Recommendation: Option B for independent development

## Current Status

All 14 implementation phases have been completed. The neat ORM is fully functional with:
- ✅ Native database/sql implementation (no GORM dependency)
- ✅ Full feature parity with eloquent
- ✅ Support for MySQL, PostgreSQL, SQLite
- ✅ Integration tests for supported databases
- ✅ Comprehensive documentation and examples

**Remaining work**: See `gaps.md` for test coverage gaps and missing integration tests (SQL Server, Turso).

---

**Created by:** Cascade AI Assistant
**Date:** May 21, 2026
**Last reviewed:** May 23, 2026
**Status**: ✅ Implementation Complete - Historical Reference
