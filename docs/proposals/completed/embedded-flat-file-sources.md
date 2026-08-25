# Embedded Flat-File Sources (CSVDB, JSONDB, XMLDB with embed.FS)

**Date**: October 2026
**Status**: Completed
**Priority**: Medium

## Problem Statement

Neat ORM provided directory-backed database drivers for CSV, JSON, and XML files (`csvdb`, `jsondb`, `xmldb`), as well as the `godb` driver for embedded Go slice structures. However, for file-based flat sources (CSV, JSON, XML), application developers who wanted to compile files into their Go binary using `//go:embed` previously had to load files into memory manually or use filesystem paths that exist on disk.

### Desired Experience

```go
//go:embed data/*.csv
var csvFS embed.FS

config := neat.DBConfig{
    Default: "csv_db",
    Connections: map[string]neat.ConnectionConfig{
        "csv_db": {
            Driver:   "csvdb",
            Database: "data", // relative path inside csvFS
            FS:       csvFS,
        },
    },
}

database, _ := neat.New(config)
defer database.Close()

// Query tables populated from embedded CSV files
var users []User
err := database.Query().Model(&User{}).Where("active = ?", true).Get(&users)
```

Additionally, developers querying single embedded files using `database.Query().Model(...)` can pass an `embed.FS` instance:

```go
//go:embed data/users.json
var dataFS embed.FS

database.Query().
    Model(neat.NewJsonFSSource(dataFS, "data/users.json")).
    Where("active = ?", true).
    Get(&users)
```

## Solution

1. Added `FS fs.FS` field to `neat.ConnectionConfig` and `db.ConnectionConfig`.
2. Updated `BuildOrm` in `database/orm/orm.go` to invoke `SetFS(connConfig.FS)` on drivers supporting embedded filesystems (`CSVDB`, `JSONDB`, `XMLDB`).
3. Updated `CSVDB`, `JSONDB`, and `XMLDB` driver `Open()` methods to scan `fs.FS` directory entries when `SetFS` has been configured, falling back to disk (`os.ReadDir`, `os.Open`) when `FS` is `nil`.
4. Added FS-aware model source constructors:
   - `neat.NewCsvFSSource(sys fs.FS, filePath string)`
   - `neat.NewCsvFSSourceWithDelimiter(sys fs.FS, filePath string, delimiter rune)`
   - `neat.NewJsonFSSource(sys fs.FS, filePath string)`
   - `neat.NewXmlFSSource(sys fs.FS, filePath string)`
