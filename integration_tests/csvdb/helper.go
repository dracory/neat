//go:build integration

package csvdb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dracory/neat"
	contractsdb "github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat/database"
	_ "modernc.org/sqlite"
)

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

// csvdbCountResult is a view model for COUNT aggregate queries.
type csvdbCountResult struct {
	Count int `db:"cnt"`
}

// csvdbSumResult is a view model for SUM aggregate queries.
type csvdbSumResult struct {
	Total float64 `db:"total_sum"`
}

// csvdbAvgResult is a view model for AVG aggregate queries.
type csvdbAvgResult struct {
	Avg float64 `db:"price_avg"`
}

// csvdbMinMaxResult is a view model for MIN/MAX aggregate queries.
type csvdbMinMaxResult struct {
	MinPrice float64 `db:"min_price"`
	MaxPrice float64 `db:"max_price"`
}

// csvdbTypedRow is a model for type inference tests.
type csvdbTypedRow struct {
	ID     int     `db:"id"`
	Count  int     `db:"count"`
	Price  float64 `db:"price"`
	Active bool    `db:"active"`
	Title  string  `db:"title"`
}

func (csvdbTypedRow) TableName() string { return "typed" }

// csvFixtures holds the CSV content for each table used in the integration tests.
var csvFixtures = map[string]string{
	"users.csv":    "id,name,email,active\n1,Alice,alice@example.com,true\n2,Bob,bob@example.com,false\n3,Charlie,charlie@example.com,true\n",
	"products.csv": "id,name,price,category\n1,Widget,19.99,Hardware\n2,Gadget,49.99,Electronics\n3,Gizmo,99.99,Electronics\n",
	"orders.csv":   "id,user_id,product_id,quantity,total\n1,1,2,3,149.97\n2,3,1,5,99.95\n3,1,3,1,99.99\n",
	"typed.csv":    "id,count,price,active,title\n1,100,19.99,true,Hello\n2,200,29.99,false,World\n3,300,39.99,true,Foo\n",
}

// writeCSVFixtures writes the CSV fixture files into a temp directory and
// returns the directory path. The directory is cleaned up automatically when
// the test finishes.
func writeCSVFixtures(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "csvdb-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	for name, content := range csvFixtures {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}
	return dir
}

// SetupCSVDBTest creates a csvdb-driver database connection backed by a temp
// directory of CSV fixture files. The database and temp directory are cleaned
// up automatically when the test finishes.
//
// Unlike the SQLite integration tests, there is no createTestTables step —
// the tables are created at Open time from the CSV file headers. There is also
// no cleanupTestData step — each test gets a fresh temp directory with fresh
// CSV files, so there is no cross-test data leakage.
func SetupCSVDBTest(t *testing.T) *database.Database {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dir := writeCSVFixtures(t)

	config := neat.DBConfig{
		Default: "csv_db",
		Connections: map[string]neat.ConnectionConfig{
			"csv_db": {
				Driver: contractsdb.DriverCSVDB,
				Database: dir,
			},
		},
	}

	db, err := neat.New(config)
	if err != nil {
		t.Fatalf("Failed to create csvdb database: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}
