# Schema Migration Interface Example

This example demonstrates the **schema Migration interface** approach, which is an alternative design pattern that exists in the codebase but is not currently used.

## Running the Example

```bash
go run main.go
```

## What This Example Demonstrates

This example shows how to use the `schema.Migration` interface defined in `contracts/database/schema/schema.go`:

```go
type Migration interface {
    Signature() string
    Up() error
    Down() error
}
```

## Schema Migration Interface Approach

### Migration Implementation

```go
type CreateUsersTable struct {
    schema contractsschema.Schema
}

func NewCreateUsersTable(schema contractsschema.Schema) *CreateUsersTable {
    return &CreateUsersTable{schema: schema}
}

func (m *CreateUsersTable) Signature() string {
    return "create_users_table"
}

func (m *CreateUsersTable) Up() error {
    return m.schema.Create("users", func(blueprint contractsschema.Blueprint) {
        blueprint.ID()
        blueprint.String("name")
        blueprint.String("email")
        blueprint.Unique("email")
        blueprint.Timestamps()
    })
}

func (m *CreateUsersTable) Down() error {
    return m.schema.DropIfExists("users")
}
```

### Usage

```go
// Create migrations with schema reference
migrations := []contractsschema.Migration{
    NewCreateUsersTable(db.Schema()),
    NewCreatePostsTable(db.Schema()),
}

// Register with schema
db.Schema().Register(migrations)

// Run migrations
for _, migration := range db.Schema().Migrations() {
    if err := migration.Up(); err != nil {
        return err
    }
}
```

## Comparison: Three Migration Approaches

### 1. Schema Migration Interface (This Example)

**Pros:**
- Interface-based design (better testability)
- Self-contained migration objects
- Similar to Seeder pattern (consistent design)

**Cons:**
- **Not actually used** in the current codebase
- Requires manual schema injection
- No built-in tracking/registry
- No transaction support
- No context support
- No batch management
- Manual execution order management

### 2. Current Migrator System (Function-Based)

**Location:** `examples/migrations/main.go`

**Pros:**
- **Currently active** and maintained
- Global registration with `migrator.RegisterMigration()`
- Built-in migration tracking in database
- Batch management
- Transaction support (configurable)
- Metadata tracking (timestamps, duration)
- Multiple ID formats (datetime, date, unix, custom)
- Rollback by step or batch
- Status checking

**Cons:**
- Function-based design (less testable)
- Global registration pattern (coupling)
- File-based discovery
- No context support
- Limited to schema builder operations
- Migration IDs derived from registration, not intrinsic

### 3. Proposed Interface-Based System

**Location:** `docs/proposals/interface-based-migration-system.md`

**Pros:**
- Interface-based design (best testability)
- Self-contained migration objects with intrinsic IDs
- Context support (cancellation, timeouts)
- Direct transaction access (raw SQL support)
- No global state
- Explicit migration management
- Metadata tracking
- Inspired by successful patterns (github.com/dracory/migrate)

**Cons:**
- **Not implemented yet** (proposal stage)
- Requires breaking changes
- Migration from current system needed

## Key Differences Summary

| Feature | Schema Interface | Current Migrator | Proposed System |
|---------|----------------|-----------------|-----------------|
| Status | Unused/Legacy | Active | Proposal |
| Design | Interface | Function-based | Interface |
| Registration | Manual | Global registry | Explicit |
| Tracking | None | Database | Database |
| Context | No | No | Yes |
| Transactions | No | Yes | Yes |
| Raw SQL | No | No | Yes |
| Batches | No | Yes | Yes |
| Metadata | None | Yes | Yes |
| Schema Access | Manual injection | Built-in | Transaction parameter |

## Why Schema Interface Isn't Used

The schema Migration interface appears to be:
1. **Legacy code** from an earlier design iteration
2. **Unused infrastructure** that was built but never adopted
3. **Abandoned approach** in favor of the current function-based migrator system

The current system chose the function-based approach with global registration, likely for:
- Simplicity of use
- Built-in tracking and management
- Batch operations
- Transaction support

However, the proposal suggests moving back to an interface-based approach (but improved) to address limitations of the current system.

## Recommendation

- **For learning**: This example shows the interface pattern that exists but isn't used
- **For current projects**: Use the current migrator system (`examples/migrations/main.go`)
- **For future**: Consider the proposed interface-based system when implemented
