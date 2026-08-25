//go:build integration

package csvdb_embedded

import (
	"embed"
	"testing"

	"github.com/dracory/neat"
	"github.com/dracory/neat/database"
	_ "modernc.org/sqlite"
)

//go:embed data/*.csv
var csvFS embed.FS

// csvdbUser maps to the users.csv fixture.
type csvdbUser struct {
	ID     int    `db:"id"`
	Name   string `db:"name"`
	Email  string `db:"email"`
	Active bool   `db:"active"`
}

func (csvdbUser) TableName() string { return "users" }

// csvdbProduct maps to the products.csv fixture.
type csvdbProduct struct {
	ID       int     `db:"id"`
	Name     string  `db:"name"`
	Price    float64 `db:"price"`
	Category string  `db:"category"`
}

func (csvdbProduct) TableName() string { return "products" }

// csvdbOrderWithUser is a view model for JOIN queries across orders and users.
type csvdbOrderWithUser struct {
	ID       int     `db:"id"`
	UserName string  `db:"user_name"`
	Total    float64 `db:"total"`
}

// SetupCSVDBEmbeddedTest creates a csvdb-driver database connection backed by
// a real embed.FS (compiled into the test binary via //go:embed).
// The database is cleaned up automatically when the test finishes.
func SetupCSVDBEmbeddedTest(t *testing.T) *database.Database {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := neat.DBConfig{
		Default: "csv_db",
		Connections: map[string]neat.ConnectionConfig{
			"csv_db": {
				Driver:   "csvdb",
				Database: "data",
				FS:       csvFS,
			},
		},
	}

	db, err := neat.New(config)
	if err != nil {
		t.Fatalf("Failed to create csvdb embedded database: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}
