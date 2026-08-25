# CSV Source with Embedded Files

This example demonstrates `neat.NewCsvFSSource` — reading a CSV file from an embedded filesystem (`embed.FS`) and querying it with the array driver.

## How It Works

```go
//go:embed data/*.csv
var csvFS embed.FS

// NewCsvFSSource reads from embed.FS and returns an array-backed model
model := neat.NewCsvFSSource(csvFS, "data/users.csv")

database.Query().
    Model(model).
    Where("active = ?", true).
    Get(&users)
```

## Running

```bash
go run .
```

## Sample Output

```
=== NewCsvFSSource — query an embedded CSV file ===
User #1: Alice <alice@example.com> (active: true)
User #3: Charlie <charlie@example.com> (active: true)
User #4: Diana <diana@example.com> (active: true)

=== All users ===
User #1: Alice (active: true)
User #2: Bob (active: false)
User #3: Charlie (active: true)
User #4: Diana (active: true)
User #5: Eve (active: false)
```
