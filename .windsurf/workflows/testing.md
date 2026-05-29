---
description: Comprehensive testing workflow for the Neat ORM project including unit tests and integration tests with Docker Compose
---

# Testing Workflow

This workflow describes the comprehensive testing approach for the Neat ORM project.

## Overview

The project uses two types of tests:
- **Unit tests**: Fast, isolated tests that don't require external dependencies
- **Integration tests**: Tests that require actual database connections (MySQL, PostgreSQL)

## Prerequisites

- Go development environment
- Docker and Docker Compose (for integration tests)

## Running Tests

### Unit Tests

Run all unit tests (excludes integration tests):

```bash
go test ./...
```

Run unit tests with verbose output:

```bash
go test -v ./...
```

### Integration Tests

Integration tests require database containers to be running.

#### Start Database Containers

```bash
docker-compose up -d
```

This starts:
- MySQL 8.0 on port 3306 (user: root, password: root, database: test)
- PostgreSQL 15 on port 55432 (user: test, password: test, database: test)

#### Run MySQL Integration Tests

```bash
go test -v -tags=integration ./integration_tests/mysql/...
```

#### Run PostgreSQL Integration Tests

```bash
go test -v -tags=integration ./integration_tests/postgres/...
```

#### Run Specific Integration Test

```bash
# MySQL
go test -v -tags=integration ./integration_tests/mysql/... -run TestName

# PostgreSQL
go test -v -tags=integration ./integration_tests/postgres/... -run TestName
```

#### Stop Database Containers

```bash
docker-compose down
```

### Run All Tests

To run both unit and integration tests:

```bash
# Start databases first
docker-compose up -d

# Run all tests
go test -v ./...

# Stop databases
docker-compose down
```

## Test Structure

### Unit Tests

Located in package directories alongside the code they test. Example:
- `database/query/query_test.go`
- `database/query/builder_test.go`

### Integration Tests

Located in `integration_tests/` directory with database-specific subdirectories:
- `integration_tests/mysql/` - MySQL-specific integration tests
- `integration_tests/postgres/` - PostgreSQL-specific integration tests
- `integration_tests/models/` - Shared test models
- `integration_tests/common/` - Shared test utilities

Integration test files must include the build tag:

```go
//go:build integration
```

## Writing New Tests

### Unit Test Example

```go
package query

import (
    "testing"
)

func TestNewFeature(t *testing.T) {
    // Setup
    q := NewQuery(context.Background(), nil, nil, "", nil, nil)
    
    // Test
    result := q.SomeMethod()
    
    // Assert
    if result == nil {
        t.Error("Expected non-nil result")
    }
}
```

### Integration Test Example

```go
//go:build integration

package mysql

import (
    "testing"
    "github.com/dracory/neat/integration_tests/models"
)

func TestNewFeature(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    db := SetupMySQLTest(t)
    query := db.Query()

    // Create test data
    user := &models.User{Name: "test"}
    if err := query.Model(&models.User{}).Create(user); err != nil {
        t.Fatalf("Failed to create user: %v", err)
    }

    // Test your feature
    var result models.User
    if err := query.Model(&models.User{}).Where("id = ?", user.ID).First(&result); err != nil {
        t.Errorf("Failed to find user: %v", err)
    }
}
```

## Database-Specific Considerations

### PostgreSQL

- **Placeholder Syntax**: Uses `$1, $2, ...` instead of `?`
- **ID Retrieval**: Uses `RETURNING id` instead of `LastInsertId()`
- **Data Types**: `BIGSERIAL`, `JSONB`, `TIMESTAMP`

### MySQL

- **Placeholder Syntax**: Uses `?`
- **ID Retrieval**: Uses `LastInsertId()`
- **Data Types**: `BIGINT UNSIGNED AUTO_INCREMENT`, `JSON`, `DATETIME`

## Common Issues

### Issue: Integration tests fail with connection refused

**Solution**: Ensure Docker containers are running with `docker-compose up -d`

### Issue: PostgreSQL syntax errors

**Solution**: Check that placeholder syntax uses `$1, $2` not `?`

### Issue: ID not set after INSERT (PostgreSQL)

**Solution**: Ensure `RETURNING id` is used in INSERT queries

## CI/CD

The GitHub Actions workflow (`.github/workflows/integration-tests.yml`) automatically:
1. Starts Docker Compose services
2. Runs integration tests for both MySQL and PostgreSQL
3. Stops containers after tests complete

## Environment Variables

Override default database settings:

**MySQL:**
- `MYSQL_HOST` (default: 127.0.0.1)
- `MYSQL_PORT` (default: 3306)
- `MYSQL_DATABASE` (default: test)
- `MYSQL_USER` (default: root)
- `MYSQL_PASS` (default: root)

**PostgreSQL:**
- `POSTGRES_HOST` (default: 127.0.0.1)
- `POSTGRES_PORT` (default: 55432)
- `POSTGRES_DATABASE` (default: test)
- `POSTGRES_USER` (default: test)
- `POSTGRES_PASS` (default: test)
- `POSTGRES_SSLMODE` (default: disable)
