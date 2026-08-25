//go:build integration

package jsondb

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dracory/neat"
	contractsdb "github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat/database"
	_ "modernc.org/sqlite"
)

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

// jsondbOrder maps to the orders.json fixture.
type jsondbOrder struct {
	ID        int     `db:"id"`
	UserID    int     `db:"user_id"`
	ProductID int     `db:"product_id"`
	Quantity  int     `db:"quantity"`
	Total     float64 `db:"total"`
}

func (jsondbOrder) TableName() string { return "orders" }

// jsondbOrderWithUser is a view model for JOIN queries across orders and users.
type jsondbOrderWithUser struct {
	ID       int     `db:"id"`
	UserName string  `db:"user_name"`
	Total    float64 `db:"total"`
}

// jsondbCountResult is a view model for COUNT aggregate queries.
type jsondbCountResult struct {
	Count int `db:"cnt"`
}

// jsondbSumResult is a view model for SUM aggregate queries.
type jsondbSumResult struct {
	Total float64 `db:"total_sum"`
}

// jsondbAvgResult is a view model for AVG aggregate queries.
type jsondbAvgResult struct {
	Avg float64 `db:"price_avg"`
}

// jsondbMinMaxResult is a view model for MIN/MAX aggregate queries.
type jsondbMinMaxResult struct {
	MinPrice float64 `db:"min_price"`
	MaxPrice float64 `db:"max_price"`
}

// jsondbTypedRow is a model for type inference tests.
type jsondbTypedRow struct {
	ID      int       `db:"id"`
	Count   int       `db:"count"`
	Price   float64   `db:"price"`
	Active  bool      `db:"active"`
	Title   string    `db:"title"`
	Created time.Time `db:"created"`
	Meta    string    `db:"meta"` // Nested object
	Tags    string    `db:"tags"` // Nested array
}

func (jsondbTypedRow) TableName() string { return "typed" }

// jsonFixtures holds the JSON/JSONL content for each table used in the integration tests.
var jsonFixtures = map[string]string{
	"users.json": `[
		{"id": 1, "name": "Alice", "email": "alice@example.com", "active": true},
		{"id": 2, "name": "Bob", "email": "bob@example.com", "active": false},
		{"id": 3, "name": "Charlie", "email": "charlie@example.com", "active": true}
	]`,
	"products.jsonl": `{"id": 1, "name": "Widget", "price": 19.99, "category": "Hardware"}
{"id": 2, "name": "Gadget", "price": 49.99, "category": "Electronics"}
{"id": 3, "name": "Gizmo", "price": 99.99, "category": "Electronics"}`,
	"orders.json": `[
		{"id": 1, "user_id": 1, "product_id": 2, "quantity": 3, "total": 149.97},
		{"id": 2, "user_id": 3, "product_id": 1, "quantity": 5, "total": 99.95},
		{"id": 3, "user_id": 1, "product_id": 3, "quantity": 1, "total": 99.99}
	]`,
	"typed.ndjson": `{"id": 1, "count": 100, "price": 19.99, "active": true, "title": "Hello", "created": "2024-01-15T10:30:00Z", "meta": {"city": "NYC", "zip": "10001"}, "tags": ["red", "blue"]}
{"id": 2, "count": 200, "price": 29.99, "active": false, "title": "World", "created": "2024-02-20T14:45:00Z", "meta": {"city": "SF"}, "tags": ["green"]}
{"id": 3, "count": 300, "price": 39.99, "active": true, "title": "Foo", "created": "2024-03-25T09:15:00Z", "meta": null, "tags": []}`,
}

// writeJSONFixtures writes the JSON fixture files into a temp directory and
// returns the directory path. The directory is cleaned up automatically when
// the test finishes.
func writeJSONFixtures(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "jsondb-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	for name, content := range jsonFixtures {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}
	return dir
}

// SetupJSONDBTest creates a jsondb-driver database connection backed by a temp
// directory of JSON/JSONL/NDJSON fixture files. The database and temp directory
// are cleaned up automatically when the test finishes.
func SetupJSONDBTest(t *testing.T) *database.Database {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dir := writeJSONFixtures(t)

	config := neat.DBConfig{
		Default: "json_db",
		Connections: map[string]neat.ConnectionConfig{
			"json_db": {
				Driver: contractsdb.DriverJSONDB,
				Database: dir,
			},
		},
	}

	db, err := neat.New(config)
	if err != nil {
		t.Fatalf("Failed to create jsondb database: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}
