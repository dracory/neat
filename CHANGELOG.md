# Changelog

All notable changes to this project will be documented in this file.

## Breaking Changes

This section documents breaking changes and provides migration guides. Most Neat ORM development has been additive, but when breaking changes occur, they will be documented here with concrete migration examples.

### Versioning Policy

- **Major version (X.0.0)**: Breaking changes to the public API surface
- **Minor version (0.X.0)**: New features, backward-compatible additions
- **Patch version (0.0.X)**: Bug fixes, internal improvements

The primary public API surface is defined in `contracts/database/orm/`. Changes to these interfaces are considered breaking changes.

### Migration Guide Template

When a breaking change is introduced, it will be documented following this format:

#### [Change Description]

**Version**: X.Y.Z

**Impact**: Description of what changed

**Migration**: Step-by-step migration guide with code examples

```go
// Before (old code)
oldCode()

// After (new code)
newCode()
```

### Current Breaking Changes

*None at this time. Neat ORM has maintained backward compatibility since v0.1.0.*

## [Unreleased]

### Added
- Observer pattern for model lifecycle events
- Soft delete functionality with SoftDeletes struct
- Model associations (BelongsTo, HasMany, HasOne)
- Eager and lazy loading (With, Load, LoadMissing)
- ToSql interface for SQL generation without execution
- Schema builder integration from eloquent
- Unit tests for driver, db, orm, query, observer, soft_delete, and association components
- Integration test infrastructure
- CI/CD GitHub Actions workflow
- Comprehensive README.md documentation
- Example directories for basic-orm, advanced-queries, models, schema-builder, migrations, configuration, transactions, observers, associations, and soft-deletes
- Documentation files for driver-registration, query-builder, schema-builder, observers, associations, soft-deletes, transactions, and migrations
- Performance benchmarks for CRUD operations and query building

### Changed
- Updated all imports from eloquent to neat package
- Updated schema builder to use neat package structure
- Added schema field to Database struct
- Added Schema() method to Database entry point

### Fixed
- Fixed lint warnings in ToSql tests by using context.TODO() instead of nil
- Fixed import and type mismatches in schema integration
- Added missing dependency for github.com/spf13/cast

### Known Issues
- Schema builder requires config adapter to convert db.DBConfig to config.Config (deferred implementation)
- Full integration testing requires database connections to be set up

## [0.1.0] - Initial Release

### Added
- Initial project structure
- Database abstraction layer
- ORM abstraction layer
- Query builder
- Driver support for MySQL, PostgreSQL, SQLite, SQL Server, Turso
- Connection pooling
- Context support
- Configuration management
- Error handling
- Support utilities
