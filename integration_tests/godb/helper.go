//go:build integration

package godb

import (
	"testing"

	"github.com/dracory/neat"
	contractsdb "github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat/database"
	"github.com/dracory/neat/database/driver"
	_ "modernc.org/sqlite"
)

type godbUser struct {
	ID     int    `db:"id"`
	Name   string `db:"name"`
	Email  string `db:"email"`
	Active bool   `db:"active"`
}

type godbProduct struct {
	ID       int     `db:"id"`
	Name     string  `db:"name"`
	Price    float64 `db:"price"`
	Category string  `db:"category"`
}

type godbOrder struct {
	ID        int     `db:"id"`
	UserID    int     `db:"user_id"`
	ProductID int     `db:"product_id"`
	Quantity  int     `db:"quantity"`
	Total     float64 `db:"total"`
}

// godbOrderWithUser is a view model for JOIN queries across orders and users.
type godbOrderWithUser struct {
	ID       int     `db:"id"`
	UserName string  `db:"user_name"`
	Total    float64 `db:"total"`
}

var usersData = []godbUser{
	{ID: 1, Name: "Alice", Email: "alice@example.com", Active: true},
	{ID: 2, Name: "Bob", Email: "bob@example.com", Active: false},
	{ID: 3, Name: "Charlie", Email: "charlie@example.com", Active: true},
}

var productsData = []godbProduct{
	{ID: 1, Name: "Widget", Price: 19.99, Category: "Hardware"},
	{ID: 2, Name: "Gadget", Price: 49.99, Category: "Electronics"},
	{ID: 3, Name: "Gizmo", Price: 99.99, Category: "Electronics"},
}

var ordersData = []godbOrder{
	{ID: 1, UserID: 1, ProductID: 2, Quantity: 3, Total: 149.97},
	{ID: 2, UserID: 3, ProductID: 1, Quantity: 5, Total: 99.95},
	{ID: 3, UserID: 1, ProductID: 3, Quantity: 1, Total: 99.99},
}

// SetupGODBTest creates a godb-driver database connection backed by Go in-memory slices.
func SetupGODBTest(t *testing.T) *database.Database {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := neat.DBConfig{
		Default: "go_db",
		Connections: map[string]neat.ConnectionConfig{
			"go_db": {
				Driver: contractsdb.DriverGODB,
				Tables: driver.Tables{
					"users":    usersData,
					"products": productsData,
					"orders":   ordersData,
				},
			},
		},
	}

	db, err := neat.New(config)
	if err != nil {
		t.Fatalf("Failed to create godb database: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}
