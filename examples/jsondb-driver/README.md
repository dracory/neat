# JSONDB Driver Example

This example demonstrates how to use Neat's `jsondb` driver to point at a directory of JSON, JSONL, or NDJSON files and query them as database tables.

The directory serves as the database, and each `.json`, `.jsonl`, or `.ndjson` file is represented as a table named after its filename (without the extension). Object keys across the rows define the column schema, with type inference and widening done automatically.

## Files

- `data/users.json`: A standard JSON array of objects (`"users"` table).
- `data/products.json`: A standard JSON array of objects (`"products"` table).
- `data/orders.jsonl`: A JSON Lines file (`"orders"` table).

## How to Run

You can run this example using the standard Go toolchain:

```bash
go run examples/jsondb-driver/main.go
```

## Running the Tests

To run the unit tests for this example:

```bash
go test -v ./examples/jsondb-driver/...
```
