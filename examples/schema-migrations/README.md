# Schema Migrations (Interface-Based)

This example demonstrates the interface-based migration system, which provides a cleaner, more structured approach to managing database schema changes in Neat ORM using the `schema.Migration` interface.

## Features Demonstrated

### Interface-Based Migration System
- **Structured Migration Objects**: Each migration is a self-contained struct implementing the `Migration` interface
- **BaseMigration Pattern**: Embed `BaseMigration` to get automatic schema access via `SchemaSetter` interface
- **Automatic Schema Injection**: Schema is automatically set during registration via `SchemaSetter` interface
- **Clean Signatures**: Migration IDs are intrinsic to the migration object via `Signature()` method
- **Type-Safe Operations**: Schema access through `GetSchema()` method instead of manual field management
- **Better Testability**: Interface-based design enables easy testing and mocking
- **No Global Registration**: Migrations are created and registered explicitly

### Migration Operations
- Creating tables with various column types
- Adding indexes to existing tables
- Adding columns to existing tables
- Rolling back migrations
- Foreign key relationships (commented for SQLite compatibility)

## Advantages Over Function-Based System

1. **No Global Registration**: Migrations are created explicitly, not registered globally
2. **Self-Contained**: Each migration object knows its own signature
3. **Better Organization**: Migrations can be organized in separate packages
4. **Type Safety**: Compile-time checking of migration structure
5. **Extensibility**: Easy to add custom behavior through struct embedding
6. **Testability**: Can test migrations in isolation
7. **IDE Support**: Better autocomplete and navigation
8. **Automatic Schema Injection**: No manual schema setting required

## Running the Example

```bash
cd examples/schema-migrations
go run main.go
```

This will:
1. Create a SQLite database (`example_schema_migrations.db`)
2. Run all migrations in order
3. Demonstrate rolling back the last migration

## Migration Structure

Each migration follows this pattern:

```go
type CreateUsersTable struct {
    schema.BaseMigration
}

func (m *CreateUsersTable) Signature() string {
    return "create_users_table"
}

func (m *CreateUsersTable) Up() error {
    return m.GetSchema().Create("users", func(blueprint contractsschema.Blueprint) {
        blueprint.ID()
        blueprint.String("name")
        blueprint.String("email")
        blueprint.Unique("email")
        blueprint.Timestamps()
    })
}

func (m *CreateUsersTable) Down() error {
    return m.GetSchema().DropIfExists("users")
}
```

## Usage Pattern

```go
// Create migration instances
migrations := []contractsschema.Migration{
    &CreateUsersTable{},
    &CreatePostsTable{},
    &AddPostsIndexes{},
}

// Register migrations with schema (automatic schema injection via SchemaSetter)
schema := db.Schema()
schema.Register(migrations)

// Run migrations
for _, migration := range migrations {
    if err := migration.Up(); err != nil {
        log.Fatal(err)
    }
}
```

## Implementation Details

### SchemaSetter Interface
The `SchemaSetter` interface is defined in `contracts/database/schema/schema.go`:

```go
type SchemaSetter interface {
    SetSchema(schema Schema)
    GetSchema() Schema
}
```

### BaseMigration Struct
The `BaseMigration` struct is implemented in `database/schema/base_migration.go`:

```go
type BaseMigration struct {
    schema Schema
}

func (b *BaseMigration) SetSchema(schema Schema) {
    b.schema = schema
}

func (b *BaseMigration) GetSchema() Schema {
    return b.schema
}
```

### Automatic Schema Injection
The `Schema.Register()` method automatically detects migrations that implement `SchemaSetter` and injects the schema:

```go
func (r *Schema) Register(migrations []Migration) {
    for _, migration := range migrations {
        if setter, ok := migration.(SchemaSetter); ok {
            setter.SetSchema(r)
        }
    }
    r.migrations = migrations
}
```

## When to Use Interface-Based Migrations

- You prefer structured, object-oriented design
- You want better testability and type safety
- You need to organize migrations in separate packages
- You want to avoid global registration patterns
- You're building a large-scale application with many migrations
- You need custom migration behavior through struct embedding

## Comparison with Function-Based System

The function-based system (see `examples/migrations`) uses global registration and closures, while this interface-based approach uses structured objects and method calls. This example demonstrates the interface-based approach using the `schema.Migration` interface directly, without the `database/migration` package.

## Prerequisites

- SQLite database (or modify the DSN to use your preferred database)
- Neat ORM with schema migration support
