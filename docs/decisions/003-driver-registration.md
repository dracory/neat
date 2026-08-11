# ADR-003: Driver Registration System

## Status

Accepted

## Context

Neat ORM needed to support multiple database drivers (MySQL, PostgreSQL, SQLite, SQL Server, Oracle, Turso, plus embedded array/CSV/JSON/XML/Go-data sources) while:
- Avoiding tight coupling to specific driver implementations
- Ensuring driver-specific code is isolated
- Providing a consistent interface across all drivers

## Decision

Neat ORM uses a switch-based driver factory. Drivers implement a common `Driver` interface and are instantiated by a `createDriver` function in `database/orm/orm.go` that maps a driver name string to the corresponding constructor:

```go
// Driver interface defines the contract all drivers must implement
type Driver interface {
    Open(dsn string) (*sql.DB, error)
    Close(db *sql.DB) error
    Ping(ctx context.Context, db *sql.DB) error
    BeginTx(ctx context.Context, db *sql.DB, opts *sql.TxOptions) (*sql.Tx, error)
    Placeholder(n int) string
    Dialect() string
}

// createDriver returns the driver implementation for the given name
func createDriver(driverName string) driver.Driver {
    switch driverName {
    case "mysql":
        return driver.NewMySQL()
    case "postgres":
        return driver.NewPostgreSQL()
    case "sqlite":
        return driver.NewSQLite()
    case "sqlserver":
        return driver.NewSQLServer()
    case "turso":
        return driver.NewTurso()
    case "oracle":
        return driver.NewOracle()
    case "array":
        return driver.NewArray()
    case "csvdb":
        return driver.NewCSVDB()
    case "jsondb":
        return driver.NewJSONDB()
    case "xmldb":
        return driver.NewXMLDB()
    case "godb":
        return driver.NewGODB()
    default:
        return driver.NewMySQL() // Default to MySQL
    }
}
```

The driver name comes from the `Driver` field of `ConnectionConfig`. The `database/driver/` package contains all driver implementations and constructors.

## Rationale

The switch-based factory provides several benefits:

1. **Simplicity**: A single function maps names to drivers — no registry state to manage
2. **Isolation**: Each driver implementation is self-contained in its own file
3. **Consistency**: All drivers implement the same `Driver` interface
4. **Compile-time safety**: Missing drivers are caught at build time (unlike a runtime registry where a missing import would only fail at runtime)
5. **Testability**: Easy to mock drivers for testing by passing a custom driver name

## Consequences

**Positive:**
- Simple to understand and maintain
- Clean separation between driver implementations
- Consistent API across all supported databases
- No initialization order issues (no `init()` functions needed)
- All drivers are always available — no "driver not registered" runtime errors

**Negative:**
- Adding a new driver requires modifying the `createDriver` switch statement
- No runtime pluggability — users cannot register custom drivers without modifying core code
- Unknown driver names silently fall back to MySQL instead of failing loudly

**Mitigations:**
- The switch is small and changes are infrequent
- Users who need a custom driver can implement the `Driver` interface and add a case
- The MySQL fallback is a deliberate default for backward compatibility
