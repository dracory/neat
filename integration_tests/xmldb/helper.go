//go:build integration

package xmldb

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dracory/neat"
	"github.com/dracory/neat/database"
	_ "modernc.org/sqlite"
)

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

// xmldbOrder maps to the orders.xml fixture.
type xmldbOrder struct {
	ID        int     `db:"id"`
	UserID    int     `db:"user_id"`
	ProductID int     `db:"product_id"`
	Quantity  int     `db:"quantity"`
	Total     float64 `db:"total"`
}

func (xmldbOrder) TableName() string { return "orders" }

// xmldbOrderWithUser is a view model for JOIN queries across orders and users.
type xmldbOrderWithUser struct {
	ID       int     `db:"id"`
	UserName string  `db:"user_name"`
	Total    float64 `db:"total"`
}

// xmldbCountResult is a view model for COUNT aggregate queries.
type xmldbCountResult struct {
	Count int `db:"cnt"`
}

// xmldbSumResult is a view model for SUM aggregate queries.
type xmldbSumResult struct {
	Total float64 `db:"total_sum"`
}

// xmldbAvgResult is a view model for AVG aggregate queries.
type xmldbAvgResult struct {
	Avg float64 `db:"price_avg"`
}

// xmldbMinMaxResult is a view model for MIN/MAX aggregate queries.
type xmldbMinMaxResult struct {
	MinPrice float64 `db:"min_price"`
	MaxPrice float64 `db:"max_price"`
}

// xmldbTypedRow is a model for type inference tests.
type xmldbTypedRow struct {
	ID      int       `db:"id"`
	Count   int       `db:"count"`
	Price   float64   `db:"price"`
	Active  bool      `db:"active"`
	Title   string    `db:"title"`
	Created time.Time `db:"created"`
	Meta    string    `db:"meta"` // Nested object
	Tags    string    `db:"tags"` // Nested array
}

func (xmldbTypedRow) TableName() string { return "typed" }

// xmlFixtures holds the XML content for each table used in the integration tests.
var xmlFixtures = map[string]string{
	"users.xml": `<users>
		<user id="1">
			<name>Alice</name>
			<email>alice@example.com</email>
			<active>true</active>
		</user>
		<user id="2">
			<name>Bob</name>
			<email>bob@example.com</email>
			<active>false</active>
		</user>
		<user id="3">
			<name>Charlie</name>
			<email>charlie@example.com</email>
			<active>true</active>
		</user>
	</users>`,
	"products.xml": `<products>
		<product id="1">
			<name>Widget</name>
			<price>19.99</price>
			<category>Hardware</category>
		</product>
		<product id="2">
			<name>Gadget</name>
			<price>49.99</price>
			<category>Electronics</category>
		</product>
		<product id="3">
			<name>Gizmo</name>
			<price>99.99</price>
			<category>Electronics</category>
		</product>
	</products>`,
	"orders.xml": `<orders>
		<order id="1">
			<user_id>1</user_id>
			<product_id>2</product_id>
			<quantity>3</quantity>
			<total>149.97</total>
		</order>
		<order id="2">
			<user_id>3</user_id>
			<product_id>1</product_id>
			<quantity>5</quantity>
			<total>99.95</total>
		</order>
		<order id="3">
			<user_id>1</user_id>
			<product_id>3</product_id>
			<quantity>1</quantity>
			<total>99.99</total>
		</order>
	</orders>`,
	"typed.xml": `<typeds>
		<typed id="1">
			<count>100</count>
			<price>19.99</price>
			<active>true</active>
			<title>Hello</title>
			<created>2024-01-15T10:30:00Z</created>
			<meta>
				<city>NYC</city>
				<zip>10001</zip>
			</meta>
			<tags>
				<tag>red</tag>
				<tag>blue</tag>
			</tags>
		</typed>
		<typed id="2">
			<count>200</count>
			<price>29.99</price>
			<active>false</active>
			<title>World</title>
			<created>2024-02-20T14:45:00Z</created>
			<meta>
				<city>SF</city>
			</meta>
			<tags>
				<tag>green</tag>
			</tags>
		</typed>
		<typed id="3">
			<count>300</count>
			<price>39.99</price>
			<active>true</active>
			<title>Foo</title>
			<created>2024-03-25T09:15:00Z</created>
			<tags></tags>
		</typed>
	</typeds>`,
}

// writeXMLFixtures writes the XML fixture files into a temp directory and
// returns the directory path. The directory is cleaned up automatically when
// the test finishes.
func writeXMLFixtures(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "xmldb-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	for name, content := range xmlFixtures {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}
	return dir
}

// SetupXMLDBTest creates an xmldb-driver database connection backed by a temp
// directory of XML fixture files. The database and temp directory
// are cleaned up automatically when the test finishes.
func SetupXMLDBTest(t *testing.T) *database.Database {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dir := writeXMLFixtures(t)

	config := neat.DBConfig{
		Default: "xml_db",
		Connections: map[string]neat.ConnectionConfig{
			"xml_db": {
				Driver:   "xmldb",
				Database: dir,
			},
		},
	}

	db, err := neat.New(config)
	if err != nil {
		t.Fatalf("Failed to create xmldb database: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}
