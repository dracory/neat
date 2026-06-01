package main

import (
	"fmt"
	"log"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/database/schema"
)

// This example demonstrates JSON query functionality with SQLite
func main() {
	if err := RunExample("sqlite://./json_example.db"); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

// RunExample demonstrates JSON query functionality
func RunExample(dsn string) error {
	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Create products table with JSON column
	err = db.Schema().Create("products", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.Json("attributes") // JSON column for storing product attributes
		blueprint.Timestamp("created_at").Nullable()
	})
	if err != nil {
		return fmt.Errorf("failed to create products table: %w", err)
	}

	// Seed sample data
	fmt.Println("=== Seeding Sample Data ===")
	products := []map[string]any{
		{
			"name":       "Laptop",
			"attributes": `{"color":"silver","price":999.99,"specs":{"cpu":"i7","ram":16},"tags":["electronics","computer"]}`,
		},
		{
			"name":       "Mouse",
			"attributes": `{"color":"black","price":29.99,"specs":{"wireless":true},"tags":["electronics","accessory"]}`,
		},
		{
			"name":       "Keyboard",
			"attributes": `{"color":"white","price":79.99,"specs":{"wireless":true,"backlit":true},"tags":["electronics","accessory"]}`,
		},
		{
			"name":       "Monitor",
			"attributes": `{"color":"black","price":299.99,"specs":{"size":27,"resolution":"4K"},"tags":["electronics","display"]}`,
		},
	}

	for _, product := range products {
		err = db.Query().Table("products").Create(product)
		if err != nil {
			return fmt.Errorf("failed to create product: %w", err)
		}
	}
	fmt.Printf("Created %d products\n", len(products))

	// Query JSON field for specific value
	fmt.Println("\n=== Query JSON Field for Value ===")
	var silverProducts []map[string]any
	err = db.Query().Table("products").WhereJsonContains("attributes->color", "silver").Get(&silverProducts)
	if err != nil {
		return fmt.Errorf("failed to query by color: %w", err)
	}
	fmt.Printf("Found %d silver products\n", len(silverProducts))
	for _, p := range silverProducts {
		fmt.Printf("  - %s\n", p["name"])
	}

	// Query JSON array for value
	fmt.Println("\n=== Query JSON Array for Value ===")
	var accessories []map[string]any
	err = db.Query().Table("products").WhereJsonContains("attributes->tags", "accessory").Get(&accessories)
	if err != nil {
		return fmt.Errorf("failed to query by tag: %w", err)
	}
	fmt.Printf("Found %d accessories\n", len(accessories))
	for _, p := range accessories {
		fmt.Printf("  - %s\n", p["name"])
	}

	// Check if JSON key exists
	fmt.Println("\n=== Check if JSON Key Exists ===")
	var wirelessProducts []map[string]any
	err = db.Query().Table("products").WhereJsonContainsKey("attributes->specs->wireless").Get(&wirelessProducts)
	if err != nil {
		return fmt.Errorf("failed to query by key existence: %w", err)
	}
	fmt.Printf("Found %d wireless products\n", len(wirelessProducts))
	for _, p := range wirelessProducts {
		fmt.Printf("  - %s\n", p["name"])
	}

	// Check JSON array length
	fmt.Println("\n=== Check JSON Array Length ===")
	var productsWithTags []map[string]any
	err = db.Query().Table("products").WhereJsonLength("attributes->tags", ">=", 2).Get(&productsWithTags)
	if err != nil {
		return fmt.Errorf("failed to query by array length: %w", err)
	}
	fmt.Printf("Found %d products with 2+ tags\n", len(productsWithTags))
	for _, p := range productsWithTags {
		fmt.Printf("  - %s\n", p["name"])
	}

	// Array indexing
	fmt.Println("\n=== Array Indexing ===")
	var firstTagProducts []map[string]any
	err = db.Query().Table("products").WhereJsonContains("attributes->tags->0", "electronics").Get(&firstTagProducts)
	if err != nil {
		return fmt.Errorf("failed to query by array index: %w", err)
	}
	fmt.Printf("Found %d products with 'electronics' as first tag\n", len(firstTagProducts))
	for _, p := range firstTagProducts {
		fmt.Printf("  - %s\n", p["name"])
	}

	// Update JSON field with path
	fmt.Println("\n=== Update JSON Field with Path ===")
	_, err = db.Query().Table("products").Where("name = ?", "Laptop").Update("attributes->color", "gray")
	if err != nil {
		return fmt.Errorf("failed to update JSON field: %w", err)
	}
	fmt.Println("Updated laptop color from silver to gray")

	// Verify the update
	var updatedProduct map[string]any
	err = db.Query().Table("products").Where("name = ?", "Laptop").Get(&updatedProduct)
	if err != nil {
		return fmt.Errorf("failed to get updated product: %w", err)
	}
	fmt.Printf("Updated product attributes: %s\n", updatedProduct["attributes"])

	// Combined queries
	fmt.Println("\n=== Combined Query with Multiple Conditions ===")
	var combinedResults []map[string]any
	err = db.Query().Table("products").
		WhereJsonContains("attributes->tags", "electronics").
		WhereJsonContainsKey("attributes->specs->wireless").
		Get(&combinedResults)
	if err != nil {
		return fmt.Errorf("failed to run combined query: %w", err)
	}
	fmt.Printf("Found %d wireless electronics products\n", len(combinedResults))
	for _, p := range combinedResults {
		fmt.Printf("  - %s\n", p["name"])
	}

	// Clean up
	fmt.Println("\n=== Cleaning Up ===")
	_, err = db.Query().Table("products").Delete()
	if err != nil {
		return fmt.Errorf("failed to clean up: %w", err)
	}
	fmt.Println("Deleted all products")

	return nil
}
