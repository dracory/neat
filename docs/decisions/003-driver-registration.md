# ADR-003: Driver Registration System

## Status

Accepted

## Context

Neat ORM needed to support multiple database drivers (MySQL, PostgreSQL, SQLite, SQL Server, Oracle, Turso) while:
- Avoiding tight coupling to specific driver implementations
- Allowing users to easily add custom drivers
- Ensuring driver-specific code is isolated
- Providing a consistent interface across all drivers

## Decision

Neat ORM uses a registration-based driver system where drivers implement a common `Driver` interface and register themselves with a central registry:

```go
// Driver interface defines the contract all drivers must implement
type Driver interface {
    Dialect() string
    Placeholder(i int) string
    Quote(value string) string
    // ... other driver-specific methods
}

// Drivers register themselves during package initialization
func init() {
    RegisterDriver("mysql", &MySQLDriver{})
}
```

The `database/driver/` package maintains a registry of available drivers and provides lookup functionality:

```go
func RegisterDriver(name string, driver Driver)
func GetDriver(name string) (Driver, error)
```

## Rationale

The registration system provides several benefits:

1. **Extensibility**: Users can register custom drivers without modifying core code
2. **Isolation**: Each driver implementation is self-contained
3. **Consistency**: All drivers implement the same interface
4. **Lazy loading**: Drivers are only loaded when needed
5. **Pluggability**: New drivers can be added as separate packages
6. **Testability**: Easy to mock drivers for testing

## Consequences

**Positive:**
- Easy to add new database drivers
- Clean separation between driver implementations
- Consistent API across all supported databases
- Custom drivers can be added by users
- Driver-specific code is isolated

**Negative:**
- Slightly more complex initialization (drivers must register themselves)
- Potential for driver name collisions
- Runtime errors if driver not registered (vs compile-time)
- Additional indirection when using drivers

**Mitigations:**
- Built-in drivers register automatically on import
- Clear naming conventions for driver names
- Error messages guide users when drivers are missing
- Documentation explains driver registration process
