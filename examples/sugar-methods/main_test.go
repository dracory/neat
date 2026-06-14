package main_test

import (
	"testing"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/database/schema"
	mainpkg "github.com/dracory/neat/examples/sugar-methods"
)

func TestRunExample(t *testing.T) {
	// Use in-memory SQLite for testing
	err := mainpkg.RunExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunExample failed: %v", err)
	}
}

func TestDjangoFilter(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	err = db.Schema().Create("products", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.String("category")
		blueprint.Integer("price")
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Seed data
	products := []map[string]any{
		{"name": "Laptop", "category": "Electronics", "price": 1200},
		{"name": "Desk", "category": "Furniture", "price": 300},
		{"name": "Mouse", "category": "Electronics", "price": 25},
	}
	for _, p := range products {
		if err = db.Query().Table("products").Create(p); err != nil {
			t.Fatalf("failed to seed: %v", err)
		}
	}

	// Test Filter (Django-style) - should return only Electronics
	var results []map[string]any
	err = db.Query().Table("products").Filter("category = ?", "Electronics").All(&results)
	if err != nil {
		t.Fatalf("Filter failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 electronics, got %d", len(results))
	}
}

func TestDjangoExclude(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	err = db.Schema().Create("products", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.String("category")
		blueprint.Integer("price")
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Seed data
	products := []map[string]any{
		{"name": "Laptop", "category": "Electronics", "price": 1200},
		{"name": "Desk", "category": "Furniture", "price": 300},
		{"name": "Mouse", "category": "Electronics", "price": 25},
	}
	for _, p := range products {
		if err = db.Query().Table("products").Create(p); err != nil {
			t.Fatalf("failed to seed: %v", err)
		}
	}

	// Test Exclude (Django-style) - should exclude Furniture
	var results []map[string]any
	err = db.Query().Table("products").Exclude("category = ?", "Furniture").All(&results)
	if err != nil {
		t.Fatalf("Exclude failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 non-furniture products, got %d", len(results))
	}
}

func TestSequelizeFindAll(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	err = db.Schema().Create("products", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.String("status")
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Seed data
	products := []map[string]any{
		{"name": "Product1", "status": "active"},
		{"name": "Product2", "status": "inactive"},
		{"name": "Product3", "status": "active"},
	}
	for _, p := range products {
		if err = db.Query().Table("products").Create(p); err != nil {
			t.Fatalf("failed to seed: %v", err)
		}
	}

	// Test FindAll (Sequelize-style) with filter
	var results []map[string]any
	err = db.Query().Table("products").Filter("status = ?", "active").FindAll(&results)
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 active products, got %d", len(results))
	}
}

func TestSequelizeFindOne(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	err = db.Schema().Create("products", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Seed data
	products := []map[string]any{
		{"name": "First"},
		{"name": "Second"},
	}
	for _, p := range products {
		if err = db.Query().Table("products").Create(p); err != nil {
			t.Fatalf("failed to seed: %v", err)
		}
	}

	// Test FindOne (Sequelize-style)
	var result map[string]any
	err = db.Query().Table("products").OrderBy("id").FindOne(&result)
	if err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}
	if result["name"] != "First" {
		t.Errorf("expected 'First', got '%v'", result["name"])
	}
}

func TestSequelizeDestroy(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	err = db.Schema().Create("products", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Seed data
	products := []map[string]any{
		{"name": "Keep"},
		{"name": "Delete"},
	}
	for _, p := range products {
		if err = db.Query().Table("products").Create(p); err != nil {
			t.Fatalf("failed to seed: %v", err)
		}
	}

	// Test Destroy (Sequelize-style)
	result, err := db.Query().Table("products").Where("name = ?", "Delete").Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("expected 1 row affected, got %d", result.RowsAffected)
	}

	// Verify deletion
	var count int64
	err = db.Query().Table("products").Count(&count)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 product remaining, got %d", count)
	}
}

func TestMixedStyles(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	err = db.Schema().Create("products", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.String("category")
		blueprint.Integer("price")
		blueprint.String("status")
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Seed data
	products := []map[string]any{
		{"name": "Laptop", "category": "Electronics", "price": 1200, "status": "active"},
		{"name": "Mouse", "category": "Electronics", "price": 25, "status": "active"},
		{"name": "Desk", "category": "Furniture", "price": 300, "status": "active"},
		{"name": "OldLaptop", "category": "Electronics", "price": 500, "status": "inactive"},
	}
	for _, p := range products {
		if err = db.Query().Table("products").Create(p); err != nil {
			t.Fatalf("failed to seed: %v", err)
		}
	}

	// Mix Django Filter with Sequelize FindAll
	var results []map[string]any
	err = db.Query().
		Table("products").
		Filter("category = ?", "Electronics"). // Django
		Filter("price > ?", 100).              // Django
		Filter("status = ?", "active").        // Django
		FindAll(&results)                      // Sequelize
	if err != nil {
		t.Fatalf("mixed style query failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 expensive active electronic, got %d", len(results))
	}
	if results[0]["name"] != "Laptop" {
		t.Errorf("expected 'Laptop', got '%v'", results[0]["name"])
	}
}
