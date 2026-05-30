# Seeders Example

This example demonstrates how to use seeders to populate your database with initial data.

## What are Seeders?

Seeders are used to populate your database with test or initial data. They provide a convenient way to fill your database with sample records for development and testing purposes.

## Running the Example

```bash
go run main.go
```

## What This Example Demonstrates

1. **Creating seeders**: Implements the `Seeder` interface with `Signature()` and `Run()` methods
2. **Using Seed()**: Runs seeders every time called
3. **Using SeedOnce()**: Runs seeders only once (subsequent calls skip them)
4. **Using Seeder facade**: Advanced operations like getting specific seeders
5. **Verifying seeded data**: Checking that data was successfully seeded

## Key Concepts

### Seeder Interface

A seeder must implement:

```go
type Seeder interface {
    Signature() string
    Run() error
}
```

### Running Seeders

```go
// Run seeders (executes every time)
db.Seed([]contractsseeder.Seeder{userSeeder, roleSeeder})

// Run seeders once (subsequent calls skip)
db.SeedOnce([]contractsseeder.Seeder{userSeeder, roleSeeder})

// Use facade for advanced operations
facade := db.Seeder()
facade.Register(seeders)
facade.Call(seeders)
```

## Best Practices

- Use descriptive signatures for your seeders
- Handle errors gracefully in your seeders
- Use `SeedOnce` for idempotent operations
- Keep seeders simple and focused
