package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/soft_delete"
)

// Product model using max-date sentinel soft delete strategy
// This allows NOT NULL constraints on the soft_deleted_at column
type Product struct {
	soft_delete.SoftDeletesMaxDate
	ID          uint      `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Price       float64   `json:"price" db:"price"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// This example demonstrates max-date sentinel soft delete strategy
func main() {
	if err := RunExample("sqlite://./max_date_example.db"); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

// RunExample demonstrates max-date sentinel soft delete usage
func RunExample(dsn string) error {
	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Create products table with NOT NULL soft_deleted_at column
	// Default value is set to the max-date sentinel (9999-12-31 23:59:59)
	err = db.Schema().Create("products", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.Float("price", 10, 2)
		blueprint.Text("description")
		blueprint.Timestamp("created_at")
		// For max-date strategy, soft_deleted_at should be NOT NULL with default sentinel value
		blueprint.Timestamp("soft_deleted_at").Default("9999-12-31 23:59:59")
	})
	if err != nil {
		return fmt.Errorf("failed to create products table: %w", err)
	}

	// Create some products
	fmt.Println("=== Creating Products ===")
	products := []map[string]any{
		{
			"name":        "Laptop",
			"price":       999.99,
			"description": "High-performance laptop",
			"created_at":  time.Now(),
		},
		{
			"name":        "Mouse",
			"price":       29.99,
			"description": "Wireless mouse",
			"created_at":  time.Now(),
		},
		{
			"name":        "Keyboard",
			"price":       79.99,
			"description": "Mechanical keyboard",
			"created_at":  time.Now(),
		},
	}

	for _, p := range products {
		err = db.Query().Table("products").Create(p)
		if err != nil {
			return fmt.Errorf("failed to create product: %w", err)
		}
	}
	fmt.Printf("Created %d products\n", len(products))

	// Count all active products (not soft deleted)
	fmt.Println("\n=== Active Products (Default Query) ===")
	var count int64
	err = db.Query().Model(&Product{}).Count(&count)
	if err != nil {
		return fmt.Errorf("failed to count products: %w", err)
	}
	fmt.Printf("Active products: %d\n", count)

	// List all active products
	var activeProducts []Product
	err = db.Query().Model(&Product{}).Get(&activeProducts)
	if err != nil {
		return fmt.Errorf("failed to get products: %w", err)
	}
	for _, p := range activeProducts {
		fmt.Printf("  - %s ($%.2f)\n", p.Name, p.Price)
	}

	// Soft delete a product (Laptop)
	fmt.Println("\n=== Soft Deleting 'Laptop' ===")
	result, err := db.Query().Model(&Product{}).Where("name = ?", "Laptop").Delete()
	if err != nil {
		return fmt.Errorf("failed to soft delete: %w", err)
	}
	fmt.Printf("Soft deleted %d product(s)\n", result.RowsAffected)

	// Count active products after soft delete
	fmt.Println("\n=== Active Products After Soft Delete ===")
	err = db.Query().Model(&Product{}).Count(&count)
	if err != nil {
		return fmt.Errorf("failed to count: %w", err)
	}
	fmt.Printf("Active products: %d (Laptop is now hidden)\n", count)

	// List active products (should not include Laptop)
	var remainingProducts []Product
	err = db.Query().Model(&Product{}).Get(&remainingProducts)
	if err != nil {
		return fmt.Errorf("failed to get products: %w", err)
	}
	for _, p := range remainingProducts {
		fmt.Printf("  - %s ($%.2f)\n", p.Name, p.Price)
	}

	// Include soft deleted products with WithTrashed()
	fmt.Println("\n=== All Products Including Soft Deleted (WithTrashed) ===")
	var allProducts []Product
	err = db.Query().Model(&Product{}).WithSoftDeleted().Get(&allProducts)
	if err != nil {
		return fmt.Errorf("failed to get all products: %w", err)
	}
	fmt.Printf("Total products (including soft deleted): %d\n", len(allProducts))
	for _, p := range allProducts {
		status := "active"
		if p.IsSoftDeleted() {
			status = "soft deleted"
		}
		fmt.Printf("  - %s ($%.2f) [%s]\n", p.Name, p.Price, status)
	}

	// Query only soft deleted products
	fmt.Println("\n=== Only Soft Deleted Products (OnlySoftDeleted) ===")
	var deletedProducts []Product
	err = db.Query().Model(&Product{}).OnlySoftDeleted().Get(&deletedProducts)
	if err != nil {
		return fmt.Errorf("failed to get deleted products: %w", err)
	}
	fmt.Printf("Soft deleted products: %d\n", len(deletedProducts))
	for _, p := range deletedProducts {
		fmt.Printf("  - %s was soft deleted at %s\n", p.Name, p.SoftDeletedAt.Format(time.RFC3339))
	}

	// Restore the soft deleted product
	fmt.Println("\n=== Restoring Soft Deleted Product ===")
	var laptop Product
	err = db.Query().Model(&Product{}).OnlySoftDeleted().Where("name = ?", "Laptop").First(&laptop)
	if err != nil {
		return fmt.Errorf("failed to find soft deleted laptop: %w", err)
	}

	restoreResult, err := db.Query().Model(&Product{}).Where("id = ?", laptop.ID).RestoreSoftDeleted()
	if err != nil {
		return fmt.Errorf("failed to restore: %w", err)
	}
	fmt.Printf("Restored %d product(s)\n", restoreResult.RowsAffected)

	// Verify restoration
	fmt.Println("\n=== Active Products After Restore ===")
	err = db.Query().Model(&Product{}).Count(&count)
	if err != nil {
		return fmt.Errorf("failed to count: %w", err)
	}
	fmt.Printf("Active products: %d (Laptop is back!)\n", count)

	// Show the restored product
	var restoredProduct Product
	err = db.Query().Model(&Product{}).Where("name = ?", "Laptop").First(&restoredProduct)
	if err != nil {
		return fmt.Errorf("failed to find restored laptop: %w", err)
	}
	fmt.Printf("Restored: %s ($%.2f) - SoftDeletedAt: %s\n",
		restoredProduct.Name,
		restoredProduct.Price,
		restoredProduct.SoftDeletedAt.Format(time.RFC3339))

	// Demonstrate force delete (permanent deletion)
	fmt.Println("\n=== Force Delete (Permanent Deletion) ===")
	forceResult, err := db.Query().Model(&Product{}).Where("name = ?", "Mouse").ForceDelete()
	if err != nil {
		return fmt.Errorf("failed to force delete: %w", err)
	}
	fmt.Printf("Force deleted %d product(s) permanently\n", forceResult.RowsAffected)

	// Final count
	fmt.Println("\n=== Final Active Product Count ===")
	err = db.Query().Model(&Product{}).WithSoftDeleted().Count(&count)
	if err != nil {
		return fmt.Errorf("failed to count: %w", err)
	}
	fmt.Printf("Total products in database: %d (Mouse was permanently deleted)\n", count)

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("Max-date sentinel soft delete strategy benefits:")
	fmt.Println("  - Compatible with NOT NULL column constraints")
	fmt.Println("  - Better index performance (range scans vs IS NULL)")
	fmt.Println("  - No NULL handling complexity in queries")

	return nil
}
