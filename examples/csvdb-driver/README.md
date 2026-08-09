# CSVDB Driver Example

This example demonstrates the `csvdb` driver, which lets you point at a directory of CSV files and query them as if they were database tables.

## How It Works

The directory is the database. Each `.csv` file is a table. The filename (without `.csv`) is the table name. The CSV header row defines the columns.

```
data/
  users.csv      → "users" table
  products.csv   → "products" table
  orders.csv     → "orders" table
```

All CSV files are parsed and loaded into an in-memory SQLite database when the connection is opened. From there, the full query builder works: WHERE, JOIN, ORDER BY, aggregates, etc.

## Running

```bash
go run .
```

## Sample Output

```
=== Active users (WHERE active = true) ===
  User #1: Alice <alice@example.com>
  User #3: Charlie <charlie@example.com>

=== Expensive products (price > 50) ===
  Product #3: Gizmo ($99.99, Electronics)

=== Orders with user names (JOIN) ===
  Order #1: Alice ordered $149.97
  Order #2: Charlie ordered $99.95
  Order #3: Alice ordered $99.99
```

## Configuration

```go
config := neat.DBConfig{
    Default: "csv_db",
    Connections: map[string]neat.ConnectionConfig{
        "csv_db": {
            Driver:   "csvdb",
            Database: "data/",   // directory path
        },
    },
}

database, _ := neat.New(config)
defer database.Close()
```

## Type Inference

Column types are inferred from the CSV data:

| CSV content              | SQLite type |
|--------------------------|-------------|
| `123`, `-456`            | INTEGER     |
| `19.99`, `-0.5`          | REAL        |
| `true`, `false`          | INTEGER (0/1) |
| `2024-01-15T10:30:00Z`   | DATETIME    |
| `2024-01-15`             | DATETIME    |
| `2024-01-15 10:30:00`    | DATETIME    |
| `01/02/2006`             | DATETIME    |
| `hello`, `abc123`        | TEXT        |
| `Inf`, `NaN`             | TEXT (not REAL) |

Mixed types in a column widen: INTEGER → REAL → TEXT. DATETIME mixed with INTEGER or REAL widens to TEXT.

### Robustness Features

- **UTF-8 BOM stripping**: CSV files exported from Excel with a BOM prefix are handled correctly
- **Duplicate column detection**: Headers with duplicate column names produce a clear error
- **Ragged row validation**: Data rows with more fields than the header are rejected; rows with fewer fields become NULL
- **Table name collision detection**: Files that produce colliding table names (e.g., `Users.csv` and `users.csv` on case-sensitive filesystems) are detected
- **Row limit**: Maximum 100,000 data rows per CSV file to prevent unbounded memory consumption
- **Transaction-wrapped loading**: All rows for a table are inserted in a single transaction for atomicity and performance
