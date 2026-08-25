//go:build integration

package jsondb_embedded

import (
	"embed"
	"testing"

	"github.com/dracory/neat"
	"github.com/dracory/neat/database"
	_ "modernc.org/sqlite"
)

//go:embed data/*.json data/*.jsonl
var jsonFS embed.FS

// jsondbUser maps to the users.json fixture.
type jsondbUser struct {
	ID     int    `db:"id"`
	Name   string `db:"name"`
	Email  string `db:"email"`
	Active bool   `db:"active"`
}

func (jsondbUser) TableName() string { return "users" }

// jsondbProduct maps to the products.jsonl fixture.
type jsondbProduct struct {
	ID       int     `db:"id"`
	Name     string  `db:"name"`
	Price    float64 `db:"price"`
	Category string  `db:"category"`
}

func (jsondbProduct) TableName() string { return "products" }

// jsondbOrderWithUser is a view model for JOIN queries across orders and users.
type jsondbOrderWithUser struct {
	ID       int     `db:"id"`
	UserName string  `db:"user_name"`
	Total    float64 `db:"total"`
}

// SetupJSONDBEmbeddedTest creates a jsondb-driver database connection backed by
// a real embed.FS (compiled into the test binary via //go:embed).
// The database is cleaned up automatically when the test finishes.
func SetupJSONDBEmbeddedTest(t *testing.T) *database.Database {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

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

	db, err := neat.New(config)
	if err != nil {
		t.Fatalf("Failed to create jsondb embedded database: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}
