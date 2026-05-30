# Changelog

All notable changes to this project will be documented in this file.

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
