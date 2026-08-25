package main

import (
	"os"
	"path/filepath"
	"testing"

	contractsdb "github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat"
	_ "modernc.org/sqlite"
)

func TestRunExample(t *testing.T) {
	if err := RunExample(); err != nil {
		t.Fatalf("RunExample failed: %v", err)
	}
}

func TestJSONDBExampleQueries(t *testing.T) {
	// Resolve the data directory
	dataDir := filepath.Join("data")
	if _, err := os.Stat(dataDir); err != nil {
		dataDir = filepath.Join("examples", "jsondb-driver", "data")
	}

	config := neat.DBConfig{
		Default: "json_db",
		Connections: map[string]neat.ConnectionConfig{
			"json_db": {
				Driver: contractsdb.DriverJSONDB,
				Database: dataDir,
			},
		},
	}

	database, err := neat.New(config)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer func() { _ = database.Close() }()

	// Query active users
	var users []User
	err = database.Query().
		Model(&User{}).
		Where("active = ?", true).
		OrderBy("name", "asc").
		Get(&users)
	if err != nil {
		t.Fatalf("failed to query users: %v", err)
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

	// Query all users
	var allUsers []User
	err = database.Query().
		Model(&User{}).
		OrderBy("id", "asc").
		Get(&allUsers)
	if err != nil {
		t.Fatalf("failed to query all users: %v", err)
	}
	if len(allUsers) != 3 {
		t.Fatalf("expected 3 total users, got %d", len(allUsers))
	}

	// Query expensive products
	var products []Product
	err = database.Query().
		Model(&Product{}).
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

	// JOIN across orders and users
	var results []OrderWithUser
	err = database.Query().
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
