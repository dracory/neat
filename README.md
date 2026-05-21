# Neat ORM

A powerful and elegant ORM (Object-Relational Mapping) library for Go, designed to provide a clean and intuitive interface for database operations. Neat aims to feature parity with Laravel's Eloquent ORM while being built from scratch without GORM dependencies.

## Features

- **Query Builder**: Fluent and intuitive query building interface
- **ORM**: Full ORM support with models and relationships
- **Schema Builder**: Database schema creation and migration
- **Multiple Database Support**: MySQL, PostgreSQL, SQLite, SQL Server, Turso
- **Transactions**: Robust transaction support
- **Observers**: Model lifecycle event system
- **Soft Deletes**: Soft delete functionality
- **Associations**: BelongsTo, HasMany, HasOne relationships
- **Connection Pooling**: Efficient connection management
- **Context Support**: Full context.Context support throughout

## Installation

```bash
go get github.com/dracory/neat
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    
    "github.com/dracory/neat"
)

type User struct {
    ID       uint   `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
}

func main() {
    // Create database connection
    db, err := neat.NewFromDSN("mysql://user:password@localhost:3306/mydb")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Query users
    var users []User
    err = db.Query().Where("name", "John").Get(&users)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Found %d users", len(users))
}
```

## Configuration

### Using DSN String

```go
db, err := neat.NewFromDSN("mysql://user:password@localhost:3306/mydb")
```

### Using DBConfig

```go
config := neat.DBConfig{
    Default: "default",
    Connections: map[string]neat.ConnectionConfig{
        "default": {
            Driver:   "mysql",
            Host:     "localhost",
            Port:     3306,
            Database: "mydb",
            Username: "user",
            Password: "password",
        },
    },
}

db, err := neat.New(config)
```

### Supported DSN Formats

- **MySQL**: `mysql://user:password@localhost:3306/database`
- **PostgreSQL**: `postgres://user:password@localhost:5432/database?sslmode=disable`
- **SQLite**: `sqlite:///path/to/database.db`
- **SQL Server**: `sqlserver://user:password@localhost:1433?database=database`

## ORM Usage

### Creating Records

```go
user := User{
    Name:  "John Doe",
    Email: "john@example.com",
}
err := db.Query().Create(&user)
```

### Querying Records

```go
var user User
err := db.Query().Where("id", 1).First(&user)

var users []User
err := db.Query().Where("name", "John").Get(&users)
```

### Updating Records

```go
err := db.Query().Where("id", 1).Update("name", "Jane")
```

### Deleting Records

```go
result, err := db.Query().Where("id", 1).Delete()
```

### Transactions

```go
err := db.Transaction(func(tx neat.Query) error {
    err := tx.Create(&user1)
    if err != nil {
        return err
    }
    
    err = tx.Create(&user2)
    if err != nil {
        return err
    }
    
    return nil
})
```

## Schema Builder

```go
err := db.Schema().Create("users", func(table neat.Blueprint) {
    table.ID()
    table.String("name")
    table.String("email").Unique()
    table.Timestamps()
})
```

## Observers

```go
type UserObserver struct{}

func (o *UserObserver) Creating(event neat.Event) error {
    log.Println("Creating user")
    return nil
}

func (o *UserObserver) Created(event neat.Event) error {
    log.Println("User created")
    return nil
}

// Register observer
db.Orm().Observe([]neat.ModelToObserver{
    {Model: User{}, Observer: UserObserver{}},
})
```

## Soft Deletes

```go
type User struct {
    neat.SoftDeletes
    ID   uint
    Name string
}

// Soft delete
db.Query().Where("id", 1).Delete()

// Include soft-deleted records
db.Query().WithTrashed().Where("id", 1).First(&user)

// Only soft-deleted records
db.Query().OnlyTrashed().Where("id", 1).First(&user)

// Restore soft-deleted record
db.Query().Restore(&user)

// Force delete (permanent)
db.Query().ForceDelete(&user)
```

## Associations

```go
type Post struct {
    ID     uint
    Title  string
    UserID uint
}

type User struct {
    ID    uint
    Name  string
    Posts []Post
}

// Eager loading
db.Query().With("posts").Where("id", 1).First(&user)

// Lazy loading
db.Query().Load(&user, "posts")

// Association operations
db.Query().Association("posts").Append(&user, &post)
```

## Supported Databases

- MySQL 5.7+
- PostgreSQL 12+
- SQLite 3+
- SQL Server 2017+
- Turso (SQLite edge)

## API Documentation

For detailed API documentation, see the [docs](./docs) directory.

## Examples

For more examples, see the [examples](./examples) directory.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## Roadmap

See [docs/implementation/plan.md](./docs/implementation/plan.md) for the implementation roadmap.
