package main

import (
	"fmt"
	"log"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/database/schema"
)

// This example demonstrates sugar methods for Django and Sequelize compatibility
func main() {
	if err := RunExample("sqlite://./sugar_methods.db"); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

// RunExample demonstrates Django-style and Sequelize-style sugar methods
func RunExample(dsn string) error {
	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Create products table for the example
	err = db.Schema().Create("products", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.String("category")
		blueprint.Integer("price")
		blueprint.String("status")
		blueprint.Timestamp("created_at")
	})
	if err != nil {
		return fmt.Errorf("failed to create products table: %w", err)
	}

	// Seed data
	products := []map[string]any{
		{"name": "Laptop", "category": "Electronics", "price": 1200, "status": "active", "created_at": "2026-01-01 10:00:00"},
		{"name": "Mouse", "category": "Electronics", "price": 25, "status": "active", "created_at": "2026-01-02 10:00:00"},
		{"name": "Desk", "category": "Furniture", "price": 300, "status": "active", "created_at": "2026-01-03 10:00:00"},
		{"name": "Chair", "category": "Furniture", "price": 150, "status": "inactive", "created_at": "2026-01-04 10:00:00"},
		{"name": "Monitor", "category": "Electronics", "price": 400, "status": "active", "created_at": "2026-01-05 10:00:00"},
	}

	for _, product := range products {
		if err := db.Query().Table("products").Create(product); err != nil {
			return fmt.Errorf("failed to seed product: %w", err)
		}
	}
	fmt.Println("=== Seeded 5 products ===")

	// ============================
	// Django-style Sugar Methods
	// ============================

	fmt.Println("\n=== Django-style: Filter() ===")
	// Filter is an alias for Where
	var electronics []map[string]any
	err = db.Query().Table("products").Filter("category = ?", "Electronics").All(&electronics)
	if err != nil {
		return fmt.Errorf("Filter failed: %w", err)
	}
	fmt.Printf("Found %d electronics using Filter():\n", len(electronics))
	for _, p := range electronics {
		fmt.Printf("  - %s ($%d)\n", p["name"], p["price"])
	}

	fmt.Println("\n=== Django-style: Exclude() ===")
	// Exclude is an alias for WhereNot
	var nonFurniture []map[string]any
	err = db.Query().Table("products").Exclude("category = ?", "Furniture").All(&nonFurniture)
	if err != nil {
		return fmt.Errorf("Exclude failed: %w", err)
	}
	fmt.Printf("Found %d non-furniture products using Exclude():\n", len(nonFurniture))
	for _, p := range nonFurniture {
		fmt.Printf("  - %s (%s)\n", p["name"], p["category"])
	}

	fmt.Println("\n=== Django-style: All() ===")
	// All is an alias for Get
	var allProducts []map[string]any
	err = db.Query().Table("products").All(&allProducts)
	if err != nil {
		return fmt.Errorf("All failed: %w", err)
	}
	fmt.Printf("Found %d total products using All()\n", len(allProducts))

	// ============================
	// Sequelize-style Sugar Methods
	// ============================

	fmt.Println("\n=== Sequelize-style: FindAll() ===")
	// FindAll is an alias for All/Get
	var activeProducts []map[string]any
	err = db.Query().Table("products").Filter("status = ?", "active").FindAll(&activeProducts)
	if err != nil {
		return fmt.Errorf("FindAll failed: %w", err)
	}
	fmt.Printf("Found %d active products using FindAll():\n", len(activeProducts))
	for _, p := range activeProducts {
		fmt.Printf("  - %s\n", p["name"])
	}

	fmt.Println("\n=== Sequelize-style: FindOne() ===")
	// FindOne is an alias for First
	var firstProduct map[string]any
	err = db.Query().Table("products").OrderBy("id").FindOne(&firstProduct)
	if err != nil {
		return fmt.Errorf("FindOne failed: %w", err)
	}
	fmt.Printf("First product using FindOne(): %s\n", firstProduct["name"])

	fmt.Println("\n=== Sequelize-style: Destroy() ===")
	// Destroy is an alias for Delete
	// First, let's add a product to destroy
	err = db.Query().Table("products").Create(map[string]any{
		"name": "Temporary", "category": "Test", "price": 1, "status": "inactive", "created_at": "2026-01-06 10:00:00",
	})
	if err != nil {
		return fmt.Errorf("failed to create temporary product: %w", err)
	}

	result, err := db.Query().Table("products").Where("name = ?", "Temporary").Destroy()
	if err != nil {
		return fmt.Errorf("Destroy failed: %w", err)
	}
	fmt.Printf("Destroyed %d product(s) using Destroy()\n", result.RowsAffected)

	// ============================
	// Chaining Example
	// ============================

	fmt.Println("\n=== Chaining Django + Sequelize Methods ===")
	var expensiveElectronics []map[string]any
	err = db.Query().
		Table("products").
		Filter("category = ?", "Electronics").      // Django-style
		Filter("price > ?", 100).                    // Django-style
		Filter("status = ?", "active").               // Django-style
		OrderBy("price", "desc").
		FindAll(&expensiveElectronics)               // Sequelize-style
	if err != nil {
		return fmt.Errorf("chained query failed: %w", err)
	}
	fmt.Printf("Found %d expensive active electronics:\n", len(expensiveElectronics))
	for _, p := range expensiveElectronics {
		fmt.Printf("  - %s ($%d)\n", p["name"], p["price"])
	}

	fmt.Println("\n=== Sugar Methods Example Complete ===")
	return nil
}
