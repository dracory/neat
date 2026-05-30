# Integration Tests

This directory contains integration tests for the neat ORM.

## Setup

Integration tests require actual database connections to be set up. The tests can be run against:

- MySQL
- PostgreSQL
- SQL Server
- SQLite
- Turso

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

# SQL Server
SQLSERVER_HOST=127.0.0.1
SQLSERVER_PORT=1433
SQLSERVER_DATABASE=test
SQLSERVER_USER=sa
SQLSERVER_PASS=YourStrong@Passw0rd

# Turso (optional - if not set, tests use local SQLite)
TURSO_URL=your-database-url.turso.io
TURSO_AUTH_TOKEN=your-auth-token
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

# Run SQL Server integration tests
go test -tags=integration ./integration_tests/sqlserver/...

# Run SQLite integration tests
go test -tags=integration ./integration_tests/sqlite/...

# Run Turso integration tests
go test -tags=integration ./integration_tests/turso/...

# Run common integration tests
go test -tags=integration ./integration_tests/common/...
```

### Turso Integration Tests

Turso integration tests can run in two modes:

1. **Remote Turso Database**: Set `TURSO_URL` and `TURSO_AUTH_TOKEN` environment variables to test against a real Turso database
2. **Local SQLite**: If Turso environment variables are not set, tests will use a local SQLite database (since Turso is SQLite-based)

## CI/CD

Integration tests are automatically run in GitHub Actions using the tests.yml workflow. This workflow sets up MySQL, PostgreSQL, and SQL Server services and runs the tests with the integration build tag.

## Docker Compose

For local development, you can use Docker Compose to spin up the required databases:

```bash
docker compose up -d
```

This will start:
- MySQL on port 3306
- PostgreSQL on port 55432
- SQL Server on port 1433

Note: SQL Server requires manual database creation. Run the following command after starting the container:

```bash
docker exec neat-sqlserver-test /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P YourStrong@Passw0rd -No -Q "IF NOT EXISTS (SELECT name FROM sys.databases WHERE name = 'test') CREATE DATABASE test;"
```
