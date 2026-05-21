# Integration Tests

This directory contains integration tests for the neat ORM.

## Setup

Integration tests require actual database connections to be set up. The tests can be run against:

- MySQL
- PostgreSQL
- SQLite
- SQL Server
- Turso

## Environment Variables

Before running integration tests, set up the following environment variables:

```bash
# MySQL
MYSQL_DSN="mysql://user:password@localhost:3306/testdb"

# PostgreSQL
POSTGRES_DSN="postgres://user:password@localhost:5432/testdb?sslmode=disable"

# SQLite
SQLITE_DSN="sqlite:///path/to/database.db"

# SQL Server
SQLSERVER_DSN="sqlserver://user:password@localhost:1433?database=testdb"

# Turso
TURSO_DSN="libsql://test-url?authToken=test-token"
```

## Running Tests

To run integration tests for a specific database:

```bash
# MySQL
go test ./integration_tests/mysql/...

# PostgreSQL
go test ./integration_tests/postgres/...

# SQLite
go test ./integration_tests/sqlite/...

# SQL Server
go test ./integration_tests/sqlserver/...
```

## Note

The integration test infrastructure is currently being set up. This directory structure is a placeholder for future integration tests.
