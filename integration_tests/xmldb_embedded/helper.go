//go:build integration

package xmldb_embedded

import (
	"embed"
	"testing"

	"github.com/dracory/neat"
	"github.com/dracory/neat/database"
	_ "modernc.org/sqlite"
)

//go:embed data/*.xml
var xmlFS embed.FS

// xmldbUser maps to the users.xml fixture.
type xmldbUser struct {
	ID     int    `db:"id"`
	Name   string `db:"name"`
	Email  string `db:"email"`
	Active bool   `db:"active"`
}

func (xmldbUser) TableName() string { return "users" }

// xmldbProduct maps to the products.xml fixture.
type xmldbProduct struct {
	ID       int     `db:"id"`
	Name     string  `db:"name"`
	Price    float64 `db:"price"`
	Category string  `db:"category"`
}

func (xmldbProduct) TableName() string { return "products" }

// xmldbOrderWithUser is a view model for JOIN queries across orders and users.
type xmldbOrderWithUser struct {
	ID       int     `db:"id"`
	UserName string  `db:"user_name"`
	Total    float64 `db:"total"`
}

// SetupXMLDBEmbeddedTest creates an xmldb-driver database connection backed by
// a real embed.FS (compiled into the test binary via //go:embed).
// The database is cleaned up automatically when the test finishes.
func SetupXMLDBEmbeddedTest(t *testing.T) *database.Database {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

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

	db, err := neat.New(config)
	if err != nil {
		t.Fatalf("Failed to create xmldb embedded database: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}
