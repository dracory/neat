# Migrations

This document describes the migration system in Neat ORM.

## What are Migrations?

Migrations allow you to version-control your database schema. They provide a way to define and apply schema changes in a consistent and reproducible manner.

## Migration Configuration

Configure migrations in your DBConfig:

```go
config := neat.DBConfig{
    Migrations: neat.MigrationConfig{
        Driver: "sql", // or "orm"
        Table:  "migrations",
    },
}
```

## Creating Migrations

Create a migration file with up and down methods:

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

## Running Migrations

### Run All Pending Migrations

```go
err := db.Migrate()
```

### Run Specific Migration

```go
err := db.Migrate("20240101000000_create_users_table")
```

## Rolling Back Migrations

### Rollback Last Migration

```go
err := db.Rollback()
```

### Rollback Multiple Migrations

```go
err := db.Rollback(3) // Rollback last 3 migrations
```

### Rollback All Migrations

```go
err := db.RollbackAll()
```

## Migration Status

### Check Migration Status

```go
status, err := db.MigrationStatus()
```

## Migration Best Practices

1. **Always provide a Down method**: Ensure you can rollback any migration
2. **Use descriptive names**: Make migration names clear and descriptive
3. **Test migrations**: Always test migrations on a copy of your database
4. **Keep migrations reversible**: Avoid destructive operations that can't be undone
5. **Use transactions**: Wrap migration operations in transactions for atomicity

## Migration Drivers

### SQL Driver

Uses raw SQL files for migrations. Suitable for complex database-specific operations.

### ORM Driver

Uses the schema builder for migrations. Provides database-agnostic migrations.

## Note

This documentation is a placeholder and will be expanded as the migration system is fully implemented.
