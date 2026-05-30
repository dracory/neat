# Seeders

Seeders are used to populate your database with test or initial data. They provide a convenient way to fill your database with sample records for development and testing purposes.

## Overview

The seeder system in Neat ORM consists of:

- **Seeder Interface**: Defines the contract for seeders with `Signature()` and `Run()` methods
- **Runner**: Executes seeders and tracks which ones have been run
- **Registry**: Global registry for registering seeders

## Creating a Seeder

A seeder must implement the `Seeder` interface:

```go
type Seeder interface {
    // Signature returns the unique signature of the seeder
    Signature() string
    // Run executes the seeder logic
    Run() error
}
```

### Example Seeder

```go
package seeders

import (
    "github.com/dracory/neat/database"
)

type UserSeeder struct {
    db *database.Database
}

func (s *UserSeeder) Signature() string {
    return "user_seeder"
}

func (s *UserSeeder) Run() error {
    users := []map[string]interface{}{
        {"name": "John Doe", "email": "john@example.com"},
        {"name": "Jane Smith", "email": "jane@example.com"},
    }

    for _, user := range users {
        err := s.db.Query().Table("users").Insert(user)
        if err != nil {
            return err
        }
    }

    return nil
}
```

## Running Seeders

### Using the Database Methods

The `Database` struct provides three methods for running seeders:

#### Seed

Runs the specified seeders. Each seeder will run every time `Seed` is called.

```go
db, _ := database.New(config)
seeders := []contractsseeder.Seeder{
    &UserSeeder{db: db},
    &RoleSeeder{db: db},
}

err := db.Seed(seeders)
if err != nil {
    log.Fatal(err)
}
```

#### SeedOnce

Runs the specified seeders only once. Subsequent calls to `SeedOnce` with the same seeders will skip them.

```go
db, _ := database.New(config)
seeders := []contractsseeder.Seeder{
    &UserSeeder{db: db},
}

// First call - runs the seeder
err := db.SeedOnce(seeders)
if err != nil {
    log.Fatal(err)
}

// Second call - skips the seeder (already run)
err = db.SeedOnce(seeders)
if err != nil {
    log.Fatal(err)
}
```

#### Seeder

Returns a seeder facade for advanced operations:

```go
db, _ := database.New(config)
facade := db.Seeder()

// Register seeders
facade.Register([]contractsseeder.Seeder{
    &UserSeeder{db: db},
    &RoleSeeder{db: db},
})

// Run specific seeders
err := facade.Call([]contractsseeder.Seeder{&UserSeeder{db: db}})
if err != nil {
    log.Fatal(err)
}

// Run seeders only once
err = facade.CallOnce([]contractsseeder.Seeder{&UserSeeder{db: db}})
if err != nil {
    log.Fatal(err)
}
```

## Using the Global Registry

You can register seeders globally using the registry:

```go
import (
    "github.com/dracory/neat/database/seeder"
)

// Register a seeder
seeder.RegisterSeeder("user_seeder", &UserSeeder{db: db})

// Retrieve a seeder
s := seeder.GetSeeder("user_seeder")

// Get all registered seeders
allSeeders := seeder.GetSeeders()
```

**Important Note**: The global registry (`database/seeder/registry.go`) is independent from the per-instance Runner returned by `db.Seeder()`. Seeders registered globally via `RegisterSeeder()` will not automatically be available to `db.Seed()` or `db.SeedOnce()`. If you want to use the global registry, you must manually retrieve seeders from it and pass them to the seeding methods.

## Seeder Dependencies

Seeders do not have built-in dependency ordering. If you need to run seeders in a specific order, you should:

1. Call them in the desired order manually
2. Or create a master seeder that calls other seeders in sequence

### Example: Master Seeder

```go
type MasterSeeder struct {
    db *database.Database
}

func (s *MasterSeeder) Signature() string {
    return "master_seeder"
}

func (s *MasterSeeder) Run() error {
    // Run seeders in order
    seeders := []contractsseeder.Seeder{
        &RoleSeeder{db: s.db},    // Run first
        &PermissionSeeder{db: s.db}, // Run second
        &UserSeeder{db: s.db},    // Run third (depends on roles)
    }

    return s.db.Seed(seeders)
}
```

## Best Practices

1. **Use descriptive signatures**: Make your seeder signatures unique and descriptive
2. **Handle errors gracefully**: Always check and return errors from your seeders
3. **Use transactions**: Wrap seeder operations in transactions for data consistency
4. **Use SeedOnce for idempotent operations**: Use `SeedOnce` when you want to ensure data is only seeded once
5. **Keep seeders simple**: Avoid complex logic in seeders
6. **Use factories for test data**: For complex test data, consider using the factory pattern instead

## Testing Seeders

You can test seeders using the `ResetCallOnce` method:

```go
func TestUserSeeder(t *testing.T) {
    db := setupTestDatabase()
    defer db.Close()

    seeder := &UserSeeder{db: db}
    facade := db.Seeder()
    facade.Register([]contractsseeder.Seeder{seeder})

    // Run the seeder
    err := facade.CallOnce([]contractsseeder.Seeder{seeder})
    if err != nil {
        t.Fatal(err)
    }

    // Verify data was seeded
    count, _ := db.Query().Table("users").Count()
    if count != 2 {
        t.Errorf("Expected 2 users, got %d", count)
    }

    // Reset for next test
    facade.ResetCallOnce()
}
```

## API Reference

### Database Methods

- `Seed(seeders []contractsseeder.Seeder) error` - Runs the specified seeders
- `SeedOnce(seeders []contractsseeder.Seeder) error` - Runs the specified seeders only once
- `Seeder() contractsseeder.Facade` - Returns a seeder facade for advanced operations

### Facade Methods

- `Register(seeders []contractsseeder.Seeder)` - Registers seeders
- `GetSeeder(name string) contractsseeder.Seeder` - Gets a seeder by signature
- `GetSeeders() []contractsseeder.Seeder` - Gets all registered seeders
- `Call(seeders []contractsseeder.Seeder) error` - Executes the specified seeders
- `CallOnce(seeders []contractsseeder.Seeder) error` - Executes the specified seeders only once

### Registry Functions

- `RegisterSeeder(name string, s contractsseeder.Seeder)` - Registers a seeder globally
- `GetSeeder(name string) contractsseeder.Seeder` - Retrieves a seeder from the global registry
- `GetSeeders() []contractsseeder.Seeder` - Retrieves all seeders from the global registry
- `ClearRegistry()` - Clears the global registry (useful for testing)
