# Integration Tests

This directory contains integration tests for the neat ORM.

## Setup

Integration tests require actual database connections to be set up. The tests can be run against:

- MySQL
- PostgreSQL
- SQLite

## Environment Variables

Before running integration tests, set up the following environment variables:

```bash
# MySQL
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_DATABASE=test
MYSQL_USER=root
MYSQL_PASS=root

# PostgreSQL
POSTGRES_HOST=127.0.0.1
POSTGRES_PORT=5432
POSTGRES_DATABASE=test
POSTGRES_USER=test
POSTGRES_PASS=test
POSTGRES_SSLMODE=disable
```

## Running Tests

To run integration tests with the integration build tag:

```bash
# Run all integration tests
go test -tags=integration ./integration_tests/...

# Run MySQL integration tests
go test -tags=integration ./integration_tests/mysql/...

# Run PostgreSQL integration tests
go test -tags=integration ./integration_tests/postgres/...

# Run SQLite integration tests
go test -tags=integration ./integration_tests/sqlite/...

# Run common integration tests
go test -tags=integration ./integration_tests/common/...
```

## CI/CD

Integration tests are automatically run in GitHub Actions using the integration-tests.yml workflow. This workflow sets up MySQL, PostgreSQL, and SQLite services and runs the tests with the integration build tag.
