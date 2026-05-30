---
name: integration-testing
description: Guide for setting up and running integration tests with MySQL and PostgreSQL using Docker Compose
---

# Integration Testing Skill

This skill provides guidance for setting up and running integration tests for the Neat ORM project with MySQL and PostgreSQL databases.

## Prerequisites

- Docker and Docker Compose installed
- Go development environment

## Database Setup with Docker Compose

### Starting Database Containers

```bash
docker-compose up -d
```

This starts:
- **MySQL 8.0** on port `3306` (user: `root`, password: `root`, database: `test`)
- **PostgreSQL 15** on port `55432` (user: `test`, password: `test`, database: `test`)

### Stopping Database Containers

```bash
docker-compose down
```

## Running Integration Tests

### MySQL Integration Tests

```bash
go test -v ./integration_tests/mysql/...
```

### PostgreSQL Integration Tests

```bash
go test -v ./integration_tests/postgres/...
```

### Specific Test

```bash
go test -v ./integration_tests/mysql/... -run TestName
```

## Database-Specific Considerations

### PostgreSQL Specifics

1. **Placeholder Syntax**: PostgreSQL uses `$1, $2, ...` instead of `?` for parameterized queries
   - Implemented in `database/query/builder_insert.go` and `database/query/builder_where.go`
   - Uses the driver's `Placeholder` function for dialect-specific placeholders

2. **ID Retrieval**: PostgreSQL doesn't support `LastInsertId()`, use `RETURNING id` instead
   - Implemented in `database/query/query_create.go`
   - Uses `Query` instead of `Exec` when `RETURNING id` is appended

3. **Data Types**: 
   - Use `BIGSERIAL` for auto-incrementing primary keys
   - Use `JSONB` for JSON columns
   - Use `TIMESTAMP` for datetime fields

### MySQL Specifics

1. **Placeholder Syntax**: MySQL uses `?` for parameterized queries
2. **ID Retrieval**: MySQL supports `LastInsertId()` after INSERT
3. **Data Types**:
   - Use `BIGINT UNSIGNED AUTO_INCREMENT` for primary keys
   - Use `JSON` for JSON columns
   - Use `DATETIME` for datetime fields

## Adding New Integration Tests

### Test Structure

Integration tests should follow this structure:

```go
package mysql // or postgres

import (
    "testing"
    "github.com/dracory/neat/integration_tests/models"
)

func TestNewFeature(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    db := SetupMySQLTest(t) // or SetupPostgresTest(t)
    query := db.Query()

    // Your test code here
}
```

### Helper Functions

- `SetupMySQLTest(t)`: Sets up MySQL connection and creates test tables
- `SetupPostgresTest(t)`: Sets up PostgreSQL connection and creates test tables
- `SetupMySQLConnection(t)`: Sets up MySQL connection without table creation
- `SetupPostgresConnection(t)`: Sets up PostgreSQL connection without table creation

### Test Models

Use models from `integration_tests/models/models.go`:
- `User`: Basic user model with soft deletes
- `Address`: BelongsTo relationship with User
- `Book`: BelongsTo relationship with User
- `People`: Simple model
- `JsonData`: For JSON query testing

## Common Issues and Solutions

### Issue: "relation does not exist"

**Cause**: Tables not created in the database

**Solution**: Ensure you're using `SetupMySQLTest` or `SetupPostgresTest` which creates tables. If using `SetupMySQLConnection` or `SetupPostgresConnection`, you need to create tables manually.

### Issue: "syntax error at or near" (PostgreSQL)

**Cause**: Incorrect placeholder syntax (`?` instead of `$1, $2`)

**Solution**: Ensure the query builder uses dialect-specific placeholders via `b.query.driver.Placeholder`

### Issue: ID not set after INSERT (PostgreSQL)

**Cause**: PostgreSQL doesn't support `LastInsertId()`

**Solution**: Use `RETURNING id` clause and scan the result

### Issue: Foreign key is 0 in relation loading

**Cause**: ID not properly retrieved after INSERT

**Solution**: Check that the `Create` method properly sets the ID using the appropriate method for the database

## Configuration Files

### docker-compose.yml

Located at project root. Defines MySQL and PostgreSQL services with:
- Health checks to ensure databases are ready
- Port mappings for local access
- Environment variables for credentials

### integration_tests/mysql/helper.go

MySQL-specific helper functions and table creation SQL.

### integration_tests/postgres/helper.go

PostgreSQL-specific helper functions and table creation SQL.

## Environment Variables

You can override default database settings using environment variables:

**MySQL:**
- `MYSQL_HOST` (default: `127.0.0.1`)
- `MYSQL_PORT` (default: `3306`)
- `MYSQL_DATABASE` (default: `test`)
- `MYSQL_USER` (default: `root`)
- `MYSQL_PASS` (default: `root`)

**PostgreSQL:**
- `POSTGRES_HOST` (default: `127.0.0.1`)
- `POSTGRES_PORT` (default: `55432`)
- `POSTGRES_DATABASE` (default: `test`)
- `POSTGRES_USER` (default: `test`)
- `POSTGRES_PASS` (default: `test`)
- `POSTGRES_SSLMODE` (default: `disable`)

## CI/CD Integration

The GitHub Actions workflow (`.github/workflows/tests.yml`) has two jobs:
- **`unit-tests`**: Runs unit tests without any database services, using `grep -v '/integration_tests/'` to exclude integration packages
- **`integration-tests`**: Spins up MySQL and PostgreSQL service containers then runs all `integration_tests/` suites without any build tags
