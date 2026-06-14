package main_test

import (
	"testing"
	"time"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/soft_delete"
	mainpkg "github.com/dracory/neat/examples/soft-delete-max-date"
)

// Product model using max-date sentinel soft delete strategy
type Product struct {
	soft_delete.SoftDeletesMaxDate
	ID          uint      `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Price       float64   `json:"price" db:"price"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

func TestRunExample(t *testing.T) {
	err := mainpkg.RunExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunExample failed: %v", err)
	}
}

func TestMaxDateSoftDelete_CreateAndSoftDelete(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	err = db.Schema().Create("products", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.Float("price", 10, 2)
		blueprint.Text("description")
		blueprint.Timestamp("created_at")
		blueprint.Timestamp("soft_deleted_at").Default("9999-12-31 23:59:59")
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Create a product
	err = db.Query().Table("products").Create(map[string]any{
		"name":        "Test Product",
		"price":       99.99,
		"description": "A test product",
		"created_at":  time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	// Count should be 1
	var count int64
	err = db.Query().Model(&Product{}).Count(&count)
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 product, got %d", count)
	}

	// Soft delete the product
	_, err = db.Query().Model(&Product{}).Where("name = ?", "Test Product").Delete()
	if err != nil {
		t.Fatalf("failed to soft delete: %v", err)
	}

	// Count should be 0 (soft deleted products are hidden by default)
	err = db.Query().Model(&Product{}).Count(&count)
	if err != nil {
		t.Fatalf("failed to count after soft delete: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 active products after soft delete, got %d", count)
	}

	// WithTrashed should show 1
	err = db.Query().Model(&Product{}).WithSoftDeleted().Count(&count)
	if err != nil {
		t.Fatalf("failed to count with trashed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 product with trashed, got %d", count)
	}
}

func TestMaxDateSoftDelete_Restore(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	err = db.Schema().Create("products", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.Float("price", 10, 2)
		blueprint.Timestamp("soft_deleted_at").Default("9999-12-31 23:59:59")
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Create and soft delete a product
	err = db.Query().Table("products").Create(map[string]any{
		"name":  "Restorable Product",
		"price": 49.99,
	})
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	_, err = db.Query().Model(&Product{}).Where("name = ?", "Restorable Product").Delete()
	if err != nil {
		t.Fatalf("failed to soft delete: %v", err)
	}

	// Verify it's soft deleted
	var product Product
	err = db.Query().Model(&Product{}).OnlySoftDeleted().Where("name = ?", "Restorable Product").First(&product)
	if err != nil {
		t.Fatalf("failed to find soft deleted product: %v", err)
	}
	if !product.IsSoftDeleted() {
		t.Error("expected product to be soft deleted")
	}

	// Restore the product
	_, err = db.Query().Model(&Product{}).Where("name = ?", "Restorable Product").RestoreSoftDeleted()
	if err != nil {
		t.Fatalf("failed to restore: %v", err)
	}

	// Verify it's restored
	var restored Product
	err = db.Query().Model(&Product{}).Where("name = ?", "Restorable Product").First(&restored)
	if err != nil {
		t.Fatalf("failed to find restored product: %v", err)
	}
	if restored.IsSoftDeleted() {
		t.Error("expected product to be restored (not soft deleted)")
	}

	// Verify SoftDeletedAt is set to MaxSoftDeletedAt
	if !restored.SoftDeletedAt.Equal(soft_delete.MaxSoftDeletedAt) {
		t.Errorf("expected SoftDeletedAt to be MaxSoftDeletedAt, got %v", restored.SoftDeletedAt)
	}
}

func TestMaxDateSoftDelete_ForceDelete(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	err = db.Schema().Create("products", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.Float("price", 10, 2)
		blueprint.Timestamp("soft_deleted_at").Default("9999-12-31 23:59:59")
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Create a product
	err = db.Query().Table("products").Create(map[string]any{
		"name":  "Deletable Product",
		"price": 29.99,
	})
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	// Force delete (permanent)
	_, err = db.Query().Model(&Product{}).Where("name = ?", "Deletable Product").ForceDelete()
	if err != nil {
		t.Fatalf("failed to force delete: %v", err)
	}

	// Count with trashed should be 0 (permanently deleted)
	var count int64
	err = db.Query().Model(&Product{}).WithSoftDeleted().Count(&count)
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 products after force delete, got %d", count)
	}
}

func TestMaxDateSoftDelete_OnlySoftDeleted(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	err = db.Schema().Create("products", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.Float("price", 10, 2)
		blueprint.Timestamp("soft_deleted_at").Default("9999-12-31 23:59:59")
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Create active and soft deleted products
	products := []map[string]any{
		{"name": "Active Product", "price": 10.00},
		{"name": "Deleted Product 1", "price": 20.00},
		{"name": "Deleted Product 2", "price": 30.00},
	}
	for _, p := range products {
		err = db.Query().Table("products").Create(p)
		if err != nil {
			t.Fatalf("failed to create: %v", err)
		}
	}

	// Soft delete two products
	_, err = db.Query().Model(&Product{}).Where("name LIKE ?", "Deleted%").Delete()
	if err != nil {
		t.Fatalf("failed to soft delete: %v", err)
	}

	// OnlySoftDeleted should return 2
	var deletedProducts []Product
	err = db.Query().Model(&Product{}).OnlySoftDeleted().Get(&deletedProducts)
	if err != nil {
		t.Fatalf("failed to get deleted products: %v", err)
	}
	if len(deletedProducts) != 2 {
		t.Errorf("expected 2 soft deleted products, got %d", len(deletedProducts))
	}

	// Default query should return 1
	var activeProducts []Product
	err = db.Query().Model(&Product{}).Get(&activeProducts)
	if err != nil {
		t.Fatalf("failed to get active products: %v", err)
	}
	if len(activeProducts) != 1 {
		t.Errorf("expected 1 active product, got %d", len(activeProducts))
	}
	if activeProducts[0].Name != "Active Product" {
		t.Errorf("expected 'Active Product', got '%s'", activeProducts[0].Name)
	}
}

func TestMaxDateSoftDelete_MaxSoftDeletedAtValue(t *testing.T) {
	// Verify the sentinel value
	expected := time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	if !soft_delete.MaxSoftDeletedAt.Equal(expected) {
		t.Errorf("MaxSoftDeletedAt should be %v, got %v", expected, soft_delete.MaxSoftDeletedAt)
	}

	// Verify a product with SoftDeletedAt = MaxSoftDeletedAt is NOT soft deleted
	product := Product{}
	product.SoftDeletedAt = soft_delete.MaxSoftDeletedAt
	if product.IsSoftDeleted() {
		t.Error("product with SoftDeletedAt = MaxSoftDeletedAt should NOT be soft deleted")
	}

	// Verify a product with SoftDeletedAt = now IS soft deleted
	product.SoftDeletedAt = time.Now()
	if !product.IsSoftDeleted() {
		t.Error("product with SoftDeletedAt = now should be soft deleted")
	}
}
