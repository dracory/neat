//go:build integration

package csvdb_embedded

import (
	"testing"
)

// TestCSVDBEmbeddedBasicQuery verifies that the csvdb driver can read CSV
// files from an embedded filesystem and query them as database tables.
func TestCSVDBEmbeddedBasicQuery(t *testing.T) {
	db := SetupCSVDBEmbeddedTest(t)

	var users []csvdbUser
	err := db.Query().Model(&csvdbUser{}).OrderBy("id", "asc").Get(&users)
	if err != nil {
		t.Fatalf("failed to query users: %v", err)
	}
	if len(users) != 3 {
		t.Fatalf("expected 3 users, got %d", len(users))
	}
	if users[0].Name != "Alice" {
		t.Errorf("expected first user 'Alice', got '%s'", users[0].Name)
	}
	if users[1].Name != "Bob" {
		t.Errorf("expected second user 'Bob', got '%s'", users[1].Name)
	}
	if users[2].Name != "Charlie" {
		t.Errorf("expected third user 'Charlie', got '%s'", users[2].Name)
	}
}

// TestCSVDBEmbeddedWhereBool verifies WHERE clause filtering with boolean
// values on embedded CSV data.
func TestCSVDBEmbeddedWhereBool(t *testing.T) {
	db := SetupCSVDBEmbeddedTest(t)

	var users []csvdbUser
	err := db.Query().
		Model(&csvdbUser{}).
		Where("active = ?", true).
		OrderBy("name", "asc").
		Get(&users)
	if err != nil {
		t.Fatalf("failed to query active users: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 active users, got %d", len(users))
	}
	if users[0].Name != "Alice" {
		t.Errorf("expected first user 'Alice', got '%s'", users[0].Name)
	}
	if users[1].Name != "Charlie" {
		t.Errorf("expected second user 'Charlie', got '%s'", users[1].Name)
	}
}

// TestCSVDBEmbeddedProductsQuery verifies querying the products table from
// embedded CSV data with a WHERE filter and ORDER BY.
func TestCSVDBEmbeddedProductsQuery(t *testing.T) {
	db := SetupCSVDBEmbeddedTest(t)

	var products []csvdbProduct
	err := db.Query().
		Model(&csvdbProduct{}).
		Where("price > ?", 50).
		OrderBy("price", "desc").
		Get(&products)
	if err != nil {
		t.Fatalf("failed to query products: %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("expected 1 expensive product, got %d", len(products))
	}
	if products[0].Name != "Gizmo" {
		t.Errorf("expected 'Gizmo', got '%s'", products[0].Name)
	}
}

// TestCSVDBEmbeddedJoin verifies JOIN queries across two embedded CSV tables.
func TestCSVDBEmbeddedJoin(t *testing.T) {
	db := SetupCSVDBEmbeddedTest(t)

	var results []csvdbOrderWithUser
	err := db.Query().
		Table("orders AS o").
		LeftJoin("users AS u ON o.user_id = u.id").
		Select("o.id, u.name AS user_name, o.total").
		OrderBy("o.id", "asc").
		Get(&results)
	if err != nil {
		t.Fatalf("failed to query join: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 joined rows, got %d", len(results))
	}
	if results[0].UserName != "Alice" {
		t.Errorf("expected first order's user 'Alice', got '%s'", results[0].UserName)
	}
}

// TestCSVDBEmbeddedConnectionInit verifies that Connection() returns a working
// query that uses the cached connection (the embedded FS data is still available).
func TestCSVDBEmbeddedConnectionInit(t *testing.T) {
	db := SetupCSVDBEmbeddedTest(t)

	conn, err := db.Connection("csv_db")
	if err != nil {
		t.Fatalf("Connection() failed: %v", err)
	}

	var users []csvdbUser
	err = conn.Query().
		Model(&csvdbUser{}).
		OrderBy("id", "asc").
		Get(&users)
	if err != nil {
		t.Fatalf("failed to query via Connection(): %v", err)
	}
	if len(users) != 3 {
		t.Fatalf("expected 3 users, got %d", len(users))
	}
	if users[0].Name != "Alice" {
		t.Errorf("expected 'Alice', got '%s'", users[0].Name)
	}
}
