# JSON Source with Embedded Files

This example demonstrates `neat.NewJsonFSSource` — reading JSON and JSONL files from an embedded filesystem (`embed.FS`) and querying them with the array driver.

## How It Works

```go
//go:embed data/*.json data/*.jsonl
var jsonFS embed.FS

// NewJsonFSSource reads from embed.FS and returns an array-backed model
model := neat.NewJsonFSSource(jsonFS, "data/users.json")

database.Query().
    Model(model).
    Where("active = ?", true).
    Get(&users)
```

Both `.json` (array of objects) and `.jsonl`/`.ndjson` (one object per line) formats are supported.

## Running

```bash
go run .
```

## Sample Output

```
=== NewJsonFSSource — query an embedded JSON file ===
User #1: Alice <alice@example.com> (active: true)
User #3: Charlie <charlie@example.com> (active: true)
User #4: Diana <diana@example.com> (active: true)

=== NewJsonFSSource — query an embedded JSONL file ===
Event #3: purchase by user 17 ($49.99)
Event #5: purchase by user 42 ($12.50)
```
