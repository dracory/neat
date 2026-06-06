# Dependencies

This document lists all direct dependencies of Neat ORM and explains their purpose.

## Database Drivers

### github.com/go-sql-driver/mysql v1.10.0
**Purpose**: MySQL database driver  
**Usage**: Required for MySQL database connections  
**Rationale**: The standard Go MySQL driver, actively maintained and widely used

### github.com/lib/pq v1.12.3
**Purpose**: PostgreSQL database driver  
**Usage**: Required for PostgreSQL database connections  
**Rationale**: The standard Go PostgreSQL driver, stable and well-tested

### github.com/microsoft/go-mssqldb v1.10.0
**Purpose**: SQL Server database driver  
**Usage**: Required for Microsoft SQL Server database connections  
**Rationale**: Official Microsoft driver for SQL Server

### github.com/sijms/go-ora/v2 v2.9.0
**Purpose**: Oracle database driver  
**Usage**: Required for Oracle database connections  
**Rationale**: Third-party Oracle driver with good Oracle support

### github.com/tursodatabase/libsql-client-go v0.0.0-20260528064733-9d5d30a29a60
**Purpose**: Turso (SQLite-compatible) database driver  
**Usage**: Required for Turso database connections  
**Rationale**: Official Turso client for SQLite-compatible edge database

### modernc.org/sqlite v1.51.0
**Purpose**: SQLite database driver  
**Usage**: Required for SQLite database connections  
**Rationale**: Pure Go SQLite implementation (no CGO), cross-platform compatible

## Utility Libraries

### github.com/spf13/cast v1.10.0
**Purpose**: Type casting utilities  
**Usage**: Used in `database/schema/` for type conversions in schema operations  
**Rationale**: Provides safe type casting between different Go types, essential for schema migrations and type handling

### github.com/samber/lo v1.53.0
**Purpose**: Functional programming utilities  
**Usage**: Used in `support/collect/` for collection operations  
**Rationale**: Provides functional programming helpers like Map, Filter, Reduce for data manipulation

## Test Dependencies

### github.com/dromara/carbon/v2 v2.6.16
**Purpose**: Date/time handling library  
**Usage**: Used in `support/database/` tests for timestamp operations  
**Rationale**: Comprehensive datetime library for testing date/time functionality

### github.com/google/uuid v1.6.0
**Purpose**: UUID generation and parsing  
**Usage**: Used in `support/database/` tests for UUID operations  
**Rationale**: Standard UUID library for testing UUID functionality

## Go Standard Library Extensions

### golang.org/x/exp v0.0.0-20260508232706-74f9aab9d74a
**Purpose**: Experimental Go features  
**Usage**: Used for modern Go language features not yet in stdlib  
**Rationale**: Access to experimental features that may become standard in future Go versions

### golang.org/x/text v0.37.0
**Purpose**: Text processing extensions  
**Usage**: Used for text encoding/decoding operations  
**Rationale**: Extended text processing capabilities beyond standard library

## Dependency Management

### Audit Process

Dependencies are audited periodically using:
```bash
go mod tidy          # Remove unused dependencies
go mod why <package> # Trace why a package is needed
```

### Adding New Dependencies

Before adding a new dependency:
1. Evaluate if the feature can be implemented without it
2. Check if an existing dependency can provide the functionality
3. Prefer standard library solutions when possible
4. Document the rationale in this file

### Removing Dependencies

Dependencies are removed when:
- They are no longer used (detected via `go mod tidy`)
- A better alternative becomes available
- The functionality is moved to the standard library

## Indirect Dependencies

The following are indirect dependencies (dependencies of our direct dependencies):
- filippo.io/edwards25519
- github.com/antlr4-go/antlr/v4
- github.com/coder/websocket
- github.com/dustin/go-humanize
- github.com/golang-sql/civil
- github.com/golang-sql/sqlexp
- github.com/google/go-cmp
- github.com/mattn/go-isatty
- github.com/ncruces/go-strftime
- github.com/remyoudompheng/bigfft
- github.com/shopspring/decimal
- golang.org/x/crypto
- golang.org/x/sys
- modernc.org/libc
- modernc.org/mathutil
- modernc.org/memory

These are managed automatically by Go modules and are not directly controlled by Neat ORM.
