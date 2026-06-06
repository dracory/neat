# ADR-002: Interface-in-Contracts Pattern

## Status

Accepted

## Context

Neat ORM needed to separate interface definitions from implementations to:
- Enable dependency injection and testing
- Allow multiple implementations (e.g., different database drivers)
- Provide a clear public API surface that users can program against
- Prevent tight coupling between application code and specific implementations

## Decision

Neat ORM places all public interfaces in the `contracts/` package, separate from their implementations in packages like `database/`, `errors/`, etc.

**Structure:**
```
contracts/
  database/
    orm/          # ORM interface definitions
  errors/         # Error interface definitions
  log/            # Logging interface definitions

database/
  db.go           # Database implementation
  query/          # Query implementation
  driver/         # Driver implementations
```

Users import interfaces from `contracts/` and implementations from `database/`:

```go
import (
    "github.com/dracory/neat/contracts/database/orm"
    "github.com/dracory/neat/database"
)

func main() {
    var db orm.Database = database.New("dsn")
}
```

## Rationale

The interface-in-contracts pattern provides several benefits:

1. **Clear API surface**: `contracts/` defines what users can rely on as stable
2. **Testing**: Easy to mock interfaces for unit tests without real database
3. **Dependency injection**: Applications can inject different implementations
4. **Versioning stability**: Interfaces change less frequently than implementations
5. **Documentation**: Users only need to read `contracts/` to understand the API
6. **Decoupling**: Implementation changes don't affect interface consumers

## Consequences

**Positive:**
- Clear separation between public API and internal implementation
- Excellent testability with mock implementations
- Easy to swap implementations (e.g., for testing or different backends)
- Stable API surface for users

**Negative:**
- Additional import statements in user code
- More complex package structure
- Potential confusion about which package to import for what
- Interface drift risk if implementations outpace interface updates

**Mitigations:**
- Documentation clearly explains the pattern
- Examples show proper import patterns
- Interface updates are synchronized with implementation changes
- README provides import guidance
