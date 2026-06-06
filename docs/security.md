# Security Guide

This document outlines security best practices and considerations when using Neat ORM.

## Table of Contents

- [Parameterized Queries](#parameterized-queries)
- [DSN Redaction](#dsn-redaction)
- [Raw SQL Usage](#raw-sql-usage)
- [User-Supplied Identifiers](#user-supplied-identifiers)
- [Connection Security](#connection-security)

## Parameterized Queries

Neat ORM uses parameterized queries by default, which protects against SQL injection attacks.

### Safe Usage

```go
// ✅ SAFE: Parameterized query
db.Table("users").Where("email = ?", userInput).Find(&users)

// ✅ SAFE: Multiple parameters
db.Table("users").Where("name = ? AND age > ?", name, age).Find(&users)

// ✅ SAFE: IN clause with parameters
db.Table("users").Where("id IN ?", ids).Find(&users)
```

### Why This Is Safe

Parameterized queries separate the SQL statement from the data values. The database driver automatically escapes and quotes the parameters, preventing SQL injection regardless of the input content.

## DSN Redaction

Neat ORM automatically redacts sensitive credentials from Data Source Names (DSN) in log messages and error reports to prevent accidental credential leakage.

### Implementation

The `redactDSN` function in [`database/db.go`](../database/db.go#L322) removes credentials from DSN strings before they appear in logs:

```go
// redactDSN removes credentials from a DSN string for safe logging/error messages.
func redactDSN(dsn string) string {
    // Handles URL-style DSNs (e.g., postgres://user:pass@host/db)
    // Handles mysql:// DSNs with user:pass@tcp(host:port)/db format
    // Returns: scheme://[REDACTED]@host/db
}
```

### Example

```go
// Original DSN: postgres://admin:secret123@localhost:5432/mydb
// Redacted DSN: postgres://[REDACTED]@localhost:5432/mydb
```

All error messages and log output that include DSN information use this function automatically.

## Raw SQL Usage

The `Raw()` method allows you to execute raw SQL, but this comes with security risks if not used carefully.

### Dangerous Usage

```go
// ❌ DANGEROUS: SQL injection vulnerability
userInput := "users; DROP TABLE users; --"
db.Raw("SELECT * FROM " + userInput).Find(&results)

// ❌ DANGEROUS: Direct string concatenation
db.Raw("SELECT * FROM " + tableName).Find(&results)
```

### Safe Usage

```go
// ✅ SAFE: Using parameterized values
db.Raw("SELECT * FROM users WHERE name = ?", userInput).Find(&results)

// ✅ SAFE: Using parameterized values with multiple placeholders
db.Raw("SELECT * FROM users WHERE name = ? AND age > ?", name, age).Find(&results)
```

### When to Use Raw SQL

Use `Raw()` only when:
- You need database-specific features not supported by Neat's query builder
- You need complex queries that are difficult to express with the ORM
- You need to execute stored procedures or database functions

Always use parameterized values with `Raw()` - never concatenate user input into the SQL string.

## User-Supplied Identifiers

When user input is used as table names, column names, or other SQL identifiers, special care is required because identifiers cannot be parameterized in most databases.

### Dangerous Usage

```go
// ❌ DANGEROUS: User-supplied table name
tableName := userInput // Could be "users; DROP TABLE users; --"
db.Table(tableName).Find(&results)

// ❌ DANGEROUS: User-supplied column name
columnName := userInput // Could be malicious
db.Select(columnName).Find(&results)
```

### Safe Approaches

#### 1. Whitelist Validation

```go
// ✅ SAFE: Whitelist allowed tables
allowedTables := map[string]bool{
    "users":    true,
    "products": true,
    "orders":   true,
}

if !allowedTables[userInput] {
    return errors.New("invalid table name")
}
db.Table(userInput).Find(&results)
```

#### 2. Identifier Quoting

```go
// ✅ SAFER: Use identifier quoting (driver-specific)
// Neat's quote functions handle this internally when using the query builder
db.Table("users").Where("name = ?", value).Find(&results)
```

#### 3. Avoid User-Supplied Identifiers

The safest approach is to avoid allowing users to specify identifiers altogether. Use fixed table/column names in your code and use parameters for values only.

## Connection Security

### SSL/TLS Connections

Neat ORM supports SSL/TLS connections for databases that require encrypted connections. Configure SSL settings in your DSN or connection configuration:

```go
// PostgreSQL with SSL
db, err := neat.New("postgres://user:pass@host:5432/db?sslmode=require")

// MySQL with SSL
db, err := neat.New("mysql://user:pass@host:3306/db?tls=true")
```

### Connection Pool Security

- Use connection pooling to limit the number of open connections
- Set appropriate timeouts to prevent connection exhaustion
- Configure read replicas to distribute load when needed

See the [database package documentation](../database/) for connection pool configuration options.

## Best Practices Summary

1. **Always use parameterized queries** for user-supplied data
2. **Never concatenate user input** into SQL strings
3. **Validate user-supplied identifiers** against a whitelist if absolutely necessary
4. **Avoid `Raw()` unless necessary** - prefer the query builder
5. **Review logs** to ensure no sensitive data is being logged
6. **Use SSL/TLS** for database connections in production
7. **Keep dependencies updated** to get security patches
8. **Run security audits** on your code regularly

## Reporting Security Issues

If you discover a security vulnerability in Neat ORM, please report it responsibly:

1. Do not create public issues for security vulnerabilities
2. Send details to the maintainers privately
3. Include steps to reproduce the vulnerability
4. Allow time for the issue to be fixed before disclosure

## Additional Resources

- [OWASP SQL Injection Guide](https://owasp.org/www-community/attacks/SQL_Injection)
- [Go Database SQL Injection Prevention](https://go.dev/doc/database/sql-injection)
- [PostgreSQL SSL Configuration](https://www.postgresql.org/docs/current/ssl-tcp.html)
- [MySQL SSL Configuration](https://dev.mysql.com/doc/refman/8.0/en/encrypted-connections.html)
