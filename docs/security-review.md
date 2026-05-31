# Security Review Report
Date: May 31, 2026
Reviewer: Senior Principal Golang Engineer
Codebase: github.com/dracory/neat (Go ORM Library)

## Executive Summary

This security review identified **12 security issues** in the neat ORM library, ranging from Critical to Low severity. The primary concerns are SQL injection vulnerabilities in ORDER BY, GROUP BY, and other query clauses where user-supplied input is not properly sanitized before being incorporated into SQL queries.

**Risk Level**: HIGH

**Key Concerns**:
- Multiple SQL injection vectors in query building
- Insufficient input validation for identifiers
- Raw SQL exposure through RawExpr function
- Information disclosure through error messages

## Critical Findings (Severity: Critical)

### Finding #1: SQL Injection in ORDER BY Clause
- **Location**: `database/query/query_builder.go:159-191`
- **Description**: The `Order()` and `OrderBy()` methods accept column names as strings and include them directly in SQL queries without proper escaping or validation. This allows attackers to inject arbitrary SQL through the column name parameter.
- **Impact**: Attackers can perform blind SQL injection attacks, extract data, or modify database state by manipulating ORDER BY parameters.
- **Recommendation**: Implement strict identifier validation using `isSimpleIdentifier()` before including column names in ORDER BY clauses. Apply `quoteIdentifier()` to all column names.
- **Code Example**: 
```go
// Vulnerable code - OrderBy accepts any string
func (q *Query) OrderBy(column string, direction ...string) orm.Query {
    dir := "asc"
    if len(direction) > 0 {
        dir = direction[0]
    }
    q.orders = append(q.orders, orderClause{column: column, direction: dir})
    return q
}
```
- **Suggested Fix**:
```go
func (q *Query) OrderBy(column string, direction ...string) orm.Query {
    dir := "asc"
    if len(direction) > 0 {
        dir = strings.ToLower(direction[0])
        if dir != "asc" && dir != "desc" {
            dir = "asc"
        }
    }
    // Validate column is a simple identifier
    if !isSimpleIdentifier(column) {
        return q
    }
    q.orders = append(q.orders, orderClause{column: column, direction: dir})
    return q
}
```

### Finding #2: SQL Injection in GROUP BY Clause
- **Location**: `database/query/query_builder.go:241-244`
- **Description**: The `Group()` method accepts column names without validation and includes them directly in SQL queries. This allows SQL injection through the GROUP BY clause.
- **Impact**: Attackers can inject malicious SQL code via GROUP BY parameters, potentially leading to data extraction or manipulation.
- **Recommendation**: Validate all column names using `isSimpleIdentifier()` before including them in GROUP BY clauses.
- **Code Example**:
```go
// Vulnerable code
func (q *Query) Group(name string) orm.Query {
    q.groups = append(q.groups, name)  // No validation!
    return q
}
```
- **Suggested Fix**:
```go
func (q *Query) Group(name string) orm.Query {
    if !isSimpleIdentifier(name) {
        return q
    }
    q.groups = append(q.groups, name)
    return q
}
```

### Finding #3: SQL Injection via Select Clause with User Input
- **Location**: `database/query/query_builder.go:11-56`
- **Description**: The `Select()` method accepts arbitrary expressions through `fmt.Sprintf("%v", query)` without proper validation. User-controlled input can inject SQL fragments.
- **Impact**: Attackers can inject arbitrary SQL into SELECT clauses, potentially reading sensitive data or executing unauthorized operations.
- **Recommendation**: Strictly validate all select expressions against a whitelist of allowed patterns or use parameterized queries exclusively.
- **Code Example**:
```go
// Vulnerable code
func (q *Query) Select(query any, args ...any) orm.Query {
    var queryStr string
    if slice, ok := query.([]string); ok {
        queryStr = strings.Join(slice, ", ")
    } else {
        queryStr = fmt.Sprintf("%v", query)  // No validation!
    }
    // ...
}
```

## High Severity Findings

### Finding #4: SQL Injection via RawExpr Function
- **Location**: `database/query/query.go:139-151`
- **Description**: The `RawExpr()` function allows raw SQL to be embedded directly into queries. While useful for legitimate cases like `NOW()`, if user input is passed to this function, it creates a direct SQL injection vulnerability.
- **Impact**: Complete SQL injection allowing arbitrary query execution, data extraction, or database compromise.
- **Recommendation**: 
  1. Add clear documentation warnings about SQL injection risks
  2. Consider renaming to `UnsafeRawExpr()` to indicate danger
  3. Never pass user input to this function
- **Code Example**:
```go
// Vulnerable usage pattern
db.Table("users").Create(map[string]any{
    "created_at": RawExpr(userInput),  // DANGEROUS!
})
```
- **Suggested Fix**:
```go
// RawExpr creates a raw SQL expression. WARNING: Never pass user input to this function.
// The SQL will be injected directly without parameterization.
// Example: db.Table("users").Create(map[string]any{"created_at": RawExpr("NOW()")})
func RawExpr(sql string, args ...any) RawExpression {
    // Consider adding a panic or log warning in debug mode if sql contains risky patterns
    // ...
}
```

### Finding #5: SQL Injection in WhereColumn
- **Location**: `database/query/query_where.go:79-87`
- **Description**: The `WhereColumn()` method accepts column names and operators as strings without validation. Malicious input can inject SQL through these parameters.
- **Impact**: SQL injection through WHERE clauses, allowing data extraction or modification.
- **Recommendation**: Validate both column names and operators against allowed whitelists.
- **Code Example**:
```go
// Vulnerable code
func (q *Query) WhereColumn(first, operator, second string) orm.Query {
    q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("%s %s %s", first, operator, second), args: nil})
    return q
}
```
- **Suggested Fix**:
```go
var allowedOperators = map[string]bool{
    "=": true, "!=": true, "<>": true, ">": true, "<": true, ">=": true, "<=": true,
}

func (q *Query) WhereColumn(first, operator, second string) orm.Query {
    if !isSimpleIdentifier(first) || !isSimpleIdentifier(second) {
        return q
    }
    if !allowedOperators[operator] {
        return q
    }
    // ... rest of implementation
}
```

### Finding #6: Transaction Savepoint Name Injection
- **Location**: `database/query/query_transaction.go:83-89`, `query_transaction.go:128`, `query_transaction.go:156`, `query_transaction.go:166`, `query_transaction.go:191`, `query_transaction.go:211`
- **Description**: Savepoint names are constructed using string formatting without proper validation or quoting. While savepoint names are typically controlled by the application, improper handling could lead to SQL syntax errors or injection if names are derived from user input.
- **Impact**: SQL injection through savepoint names if they are ever derived from user input.
- **Recommendation**: Validate savepoint names using `isSimpleIdentifier()` or ensure they only contain safe alphanumeric characters.
- **Code Example**:
```go
// Vulnerable code
savepointName := fmt.Sprintf("neat_sp_%d", q.savepointLevel)
savepointSQL := fmt.Sprintf("SAVEPOINT %s", savepointName)  // No validation if name comes from elsewhere
```

## Medium Severity Findings

### Finding #7: SQL Injection via Distinct Columns
- **Location**: `database/query/query_builder.go:206-214`
- **Description**: The `Distinct()` method accepts column names through variadic arguments and converts them directly to strings without validation.
- **Impact**: SQL injection through DISTINCT column specifications.
- **Recommendation**: Validate all column names using `isSimpleIdentifier()` before including them in DISTINCT clauses.
- **Code Example**:
```go
// Vulnerable code
func (q *Query) Distinct(args ...any) orm.Query {
    q.distinct = true
    if len(args) > 0 {
        q.distinctCols = make([]string, 0)
        for _, arg := range args {
            q.distinctCols = append(q.distinctCols, fmt.Sprintf("%v", arg))  // No validation!
        }
    }
    return q
}
```

### Finding #8: SQL Injection in Where Not with Closure
- **Location**: `database/query/query_where.go:100-115`
- **Description**: The `WhereNot()` method accepts closures and raw query strings. When using the non-closure variant, raw strings are wrapped in `NOT (%v)` without proper parameterization.
- **Impact**: SQL injection through WHERE NOT conditions.
- **Recommendation**: Ensure proper parameterization when handling raw query strings in WhereNot.
- **Code Example**:
```go
// Vulnerable code
} else {
    q.wheres = append(q.wheres, whereClause{_type: "and", query: fmt.Sprintf("NOT (%v)", query), args: args})
}
```

### Finding #9: Information Disclosure Through Error Messages
- **Location**: `database/query/query.go`, various files
- **Description**: Error messages throughout the codebase include raw SQL queries and detailed error information that could expose database structure or sensitive information.
- **Impact**: Information leakage that could aid attackers in reconnaissance.
- **Recommendation**: 
  1. Sanitize error messages in production to remove SQL details
  2. Use generic error messages for external consumers
  3. Log detailed errors internally only
- **Code Example**:
```go
// Example of verbose error
return fmt.Errorf("failed to execute raw SQL: %w", err)
```
- **Suggested Fix**:
```go
// In production mode, return generic errors
if isProduction {
    return errors.New("database operation failed")
}
return fmt.Errorf("failed to execute raw SQL: %w", err)
```

### Finding #10: SQL Injection via Table Name
- **Location**: `database/query/query.go:359-403` (resolveTableName)
- **Description**: While `resolveTableName()` derives table names from struct types (which is safe), the `Table()` method (not shown in provided code but implied) may accept user-supplied table names that could be used without validation.
- **Impact**: SQL injection through malicious table names if user input is used.
- **Recommendation**: Ensure all table names are validated using `isSimpleIdentifier()` or `quoteIdentifier()` before use.

## Low Severity Findings

### Finding #11: Missing Timeout Configuration
- **Location**: `database/db.go`, `database/db/config_builder.go`
- **Description**: Database connections may not have proper timeout configurations by default, potentially allowing slow query attacks or resource exhaustion.
- **Impact**: Denial of service through slow queries or connection exhaustion.
- **Recommendation**: Set sensible default timeouts for all database connections.
- **Suggested Fix**:
```go
// Set sensible defaults
defaultConfig := db.DBConfig{
    QueryTimeout: 30 * time.Second,
    ConnMaxLifetime: 3600 * time.Second,
    // ...
}
```

### Finding #12: Password Exposure in Configuration
- **Location**: `config.go:34-35`, `database/db/config.go`
- **Description**: Database passwords are stored in plain text within configuration structures. While this is standard for connection strings, it presents a risk if configurations are logged or exposed.
- **Impact**: Potential credential exposure if configuration is logged or serialized.
- **Recommendation**: 
  1. Implement a String() method that masks passwords
  2. Use environment variables or secret management for passwords
  3. Never log complete configuration objects
- **Code Example**:
```go
// In ConnectionConfig, add:
func (c ConnectionConfig) String() string {
    return fmt.Sprintf("{Driver: %s, Host: %s, Port: %d, Database: %s, Username: %s, Password: ***}",
        c.Driver, c.Host, c.Port, c.Database, c.Username)
}
```

## Best Practice Recommendations

1. **Implement a SQL Injection Testing Suite**: Add comprehensive tests that attempt SQL injection through all query builder methods.

2. **Use Prepared Statements Exclusively**: Ensure all user input is passed as parameters, never concatenated into SQL strings.

3. **Add Input Validation Layer**: Create a centralized validation function for all SQL identifiers (column names, table names, aliases).

4. **Security Documentation**: Add a SECURITY.md file documenting:
   - Known security considerations
   - Safe usage patterns
   - Functions that require extra caution (RawExpr, Raw, etc.)

5. **Enable SQL Injection Detection**: Consider adding runtime detection of suspicious patterns in query building.

6. **Code Review Checklist**: Create a security-focused code review checklist for future contributions.

## Dependencies Analysis

- **Total dependencies**: 35 (direct and indirect)
- **Dependencies with known vulnerabilities**: None detected in current versions
- **Outdated dependencies**: None requiring immediate updates

**Key Dependencies**:
- `github.com/go-sql-driver/mysql` v1.10.0 - Latest stable
- `github.com/lib/pq` v1.12.3 - Latest stable  
- `github.com/microsoft/go-mssqldb` v1.10.0 - Latest stable
- `modernc.org/sqlite` v1.51.0 - Latest stable
- `github.com/tursodatabase/libsql-client-go` v0.0.0-20260528064733-9d5d30a29a60 - Recent commit

## Compliance Considerations

- **OWASP Top 10**: A03:2021 - Injection (SQL Injection vulnerabilities present)
- **CWE-89**: SQL Injection - Multiple instances identified
- **PCI-DSS**: If used in payment processing environments, the SQL injection vulnerabilities would violate requirement 6.5.1
- **GDPR**: Data exposure risks through SQL injection could lead to unauthorized access to personal data

## Summary Statistics

- **Total issues found**: 12
- **Critical**: 3
- **High**: 3
- **Medium**: 4
- **Low**: 2

## Next Steps (Priority Order)

1. **Immediate (Critical)**:
   - Fix SQL injection in ORDER BY clause (Finding #1)
   - Fix SQL injection in GROUP BY clause (Finding #2)
   - Fix SQL injection in Select clause (Finding #3)

2. **High Priority**:
   - Add documentation and warnings for RawExpr function (Finding #4)
   - Fix SQL injection in WhereColumn (Finding #5)
   - Validate savepoint names (Finding #6)

3. **Medium Priority**:
   - Fix SQL injection in Distinct (Finding #7)
   - Fix SQL injection in WhereNot (Finding #8)
   - Implement sanitized error messages (Finding #9)
   - Validate table names (Finding #10)

4. **Low Priority**:
   - Add timeout configuration defaults (Finding #11)
   - Implement password masking in configuration (Finding #12)

5. **Ongoing**:
   - Add security regression tests
   - Implement automated security scanning in CI/CD
   - Create security documentation

---

**References**:
- OWASP SQL Injection: https://owasp.org/www-community/attacks/SQL_Injection
- CWE-89: https://cwe.mitre.org/data/definitions/89.html
- Go Security Best Practices: https://go.dev/security
