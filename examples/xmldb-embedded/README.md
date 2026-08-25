# XMLDB Driver with Embedded Files

This example demonstrates the `xmldb` driver using **embedded XML files** via `//go:embed`. The data is compiled into the Go binary — no external files needed at runtime.

## How It Works

```go
//go:embed data/*.xml
var xmlFS embed.FS

config := neat.DBConfig{
    Default: "xml_db",
    Connections: map[string]neat.ConnectionConfig{
        "xml_db": {
            Driver:   "xmldb",
            Database: "data",
            FS:       xmlFS,
        },
    },
}
```

The driver reads the `data/` directory from the embedded filesystem. Each `.xml` file becomes a table.

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
