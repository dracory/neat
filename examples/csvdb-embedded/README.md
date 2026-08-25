# CSVDB Driver with Embedded Files

This example demonstrates the `csvdb` driver using **embedded CSV files** via `//go:embed`. The CSV data is compiled into the Go binary — no external files needed at runtime.

## How It Works

```go
//go:embed data/*.csv
var csvFS embed.FS

config := neat.DBConfig{
    Default: "csv_db",
    Connections: map[string]neat.ConnectionConfig{
        "csv_db": {
            Driver:   "csvdb",
            Database: "data",    // root directory inside embed.FS
            FS:       csvFS,     // pass the embedded filesystem
        },
    },
}
```

The driver reads the `data/` directory from the embedded filesystem. Each `.csv` file becomes a table. The full query builder works: WHERE, JOIN, ORDER BY, aggregates, etc.

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

## vs. Disk-Based CSVDB

| Feature | `csvdb` (disk) | `csvdb` (embedded) |
|---------|----------------|---------------------|
| Data source | `os.ReadDir` | `fs.ReadDir` via `embed.FS` |
| Runtime files | Required | Not needed |
| Binary size | Smaller | Larger (data embedded) |
| Configuration | `Database: "data/"` | `Database: "data", FS: csvFS` |
