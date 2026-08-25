# JSONDB Driver with Embedded Files

This example demonstrates the `jsondb` driver using **embedded JSON/JSONL files** via `//go:embed`. The data is compiled into the Go binary — no external files needed at runtime.

## How It Works

```go
//go:embed data/*.json data/*.jsonl data/*.ndjson
var jsonFS embed.FS

config := neat.DBConfig{
    Default: "json_db",
    Connections: map[string]neat.ConnectionConfig{
        "json_db": {
            Driver:   "jsondb",
            Database: "data",
            FS:       jsonFS,
        },
    },
}
```

The driver reads the `data/` directory from the embedded filesystem. Each `.json`, `.jsonl`, or `.ndjson` file becomes a table.

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
