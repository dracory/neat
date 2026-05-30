# Migrations

This document describes the migration system in Neat ORM.

## What are Migrations?

Migrations allow you to version-control your database schema. They provide a way to define and apply schema changes in a consistent and reproducible manner.

## Migration Configuration

Configure migrations in your DBConfig:

```go
config := neat.DBConfig{
    Migrations: neat.MigrationConfig{
        Driver: "orm", // Currently only "orm" driver is supported
        Table:  "migrations",
    },
}
```

## Creating Migrations

### Using the Migrator

Create a new migration file using the Migrator:

```go
db, err := neat.NewFromDSN("sqlite://./mydb.db")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Create a new migration file
err = db.Migrate("create users table", "./migrations")
if err != nil {
    log.Fatal(err)
}
```

This will create a file like `1717080000_create_users_table.go` in the `./migrations` directory with the following template:

```go
package migrations

import (
	"github.com/dracory/neat/database/schema"
)

// Up applies the migration
func Up(schema schema.Schema) error {
	// Add your migration logic here
	return nil
}

// Down rolls back the migration
func Down(schema schema.Schema) error {
	// Add your rollback logic here
	return nil
}
```

### Manual Migration Creation

You can also create migration files manually. Here's an example:

```go
package migrations

import "github.com/dracory/neat/database/schema"

func Up(schema schema.Schema) error {
    schema.Create("users", func(table schema.Blueprint) {
        table.ID()
        table.String("name")
        table.String("email").Unique()
        table.Timestamps()
    })
    return nil
}

func Down(schema schema.Schema) error {
    schema.DropTableIfExists("users")
    return nil
}
```

## Registering Migrations

To use migrations, you need to register them in the global registry:

```go
import (
    "github.com/dracory/neat/database/migration"
    "github.com/dracory/neat/database/schema"
    contractsschema "github.com/dracory/neat/contracts/database/schema"
)

func init() {
    migration.RegisterMigration("1717080000_create_users_table", migration.Migration{
        Up: func(schema contractsschema.Schema) error {
            // Your up logic
            return nil
        },
        Down: func(schema contractsschema.Schema) error {
            // Your down logic
            return nil
        },
    })
}
```

## Running Migrations

### Run All Pending Migrations

```go
err := db.Migrate("./migrations")
```

### Run Migrations from Multiple Paths

```go
err := db.Migrate("./migrations", "./custom/migrations")
```

## Rolling Back Migrations

### Rollback Last Migration

```go
err := db.MigrateDown(1, "./migrations")
```

### Rollback Multiple Migrations

```go
err := db.MigrateDown(3, "./migrations") // Rollback last 3 migrations
```

### Fresh Migration (Drop All Tables and Re-run)

```go
err := db.MigrateFresh("./migrations")
```

### Reset Migration (Rollback All and Re-run)

```go
err := db.MigrateReset("./migrations")
```

## Migration Status

### Check Migration Status

```go
status, err := db.MigrationStatus("./migrations")
if err != nil {
    log.Fatal(err)
}

for _, s := range status {
    fmt.Printf("%s: Batch %d, Ran: %v\n", s.Name, s.Batch, s.Ran)
}
```

## Migration Best Practices

1. **Always provide a Down method**: Ensure you can rollback any migration
2. **Use descriptive names**: Make migration names clear and descriptive
3. **Test migrations**: Always test migrations on a copy of your database
4. **Keep migrations reversible**: Avoid destructive operations that can't be undone
5. **Use transactions**: The migration system automatically wraps operations in transactions for atomicity
6. **Register migrations in init()**: Ensure migrations are registered before they are run

## Migration Repository

The migration system uses a repository table to track which migrations have been applied. By default, this table is named `migrations` but can be configured via the `MigrationConfig.Table` field.

The repository table contains:
- `id`: Auto-incrementing primary key
- `migration`: The migration name (e.g., "1717080000_create_users_table")
- `batch`: The batch number (migrations run together share the same batch)
- `created_at`: Timestamp when the migration was applied

## Current Limitations

- Only the "orm" driver is currently supported (uses schema builder)
- Migrations must be manually registered in the global registry
- Migration file generation creates Go files that need to be manually edited to register the migration

## Migration Loading

The migration system supports two ways to load migrations:

1. **Global Registry** - Migrations registered via `migration.RegisterMigration()` are automatically loaded, even without migration files on disk. This is useful for testing and simple setups.

2. **File-based Loading** - If migration paths are provided and directories exist, the system will also scan for `.go` files in those directories and load corresponding migrations from the registry.

The system first loads all migrations from the global registry, then optionally supplements with file-based loading if paths are provided. This allows for flexible migration management - you can register migrations programmatically or use file-based discovery, or both.
