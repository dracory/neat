# Driver Registration

This document describes how to register and use database drivers in Neat ORM.

## Supported Drivers

Neat ORM supports the following database drivers:

- MySQL
- PostgreSQL
- SQLite
- SQL Server
- Turso (SQLite edge)

## Driver Configuration

Drivers are configured in the DBConfig struct:

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
```

## Driver-Specific Options

### MySQL

- `Charset`: Character set (default: utf8mb4)
- `Loc`: Locale setting

### PostgreSQL

- `Schema`: Database schema (default: public)
- `SSLMode`: SSL mode (disable, require, verify-ca, verify-full)
- `Timezone`: Timezone setting

### SQL Server

- No specific options currently supported

### SQLite

- Database file path

### Turso

- Database URL
- Auth token

## Custom Drivers

Custom drivers can be registered by implementing the Driver interface:

```go
type Driver interface {
    Open(dsn string) (*sql.DB, error)
    Close(db *sql.DB) error
    Ping(ctx context.Context, db *sql.DB) error
    BeginTx(ctx context.Context, db *sql.DB, opts *sql.TxOptions) (*sql.Tx, error)
    Placeholder(n int) string
    Dialect() string
}
```

## Note

Driver registration is handled internally by the ORM. This documentation is a placeholder and will be expanded as the driver system is fully implemented.
