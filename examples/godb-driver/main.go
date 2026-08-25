package main

import (
	"fmt"
	"log"

	contractsdb "github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat"
	"github.com/dracory/neat/database/driver"
	"github.com/dracory/neat/examples/godb-driver/data"
	_ "modernc.org/sqlite"
)

// BlogWithCategory is a view model for our JOIN query
type BlogWithCategory struct {
	ID           int64  `db:"id"`
	Title        string `db:"title"`
	CategoryName string `db:"category_name"`
}

func main() {
	if err := RunExample(); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

// RunExample demonstrates the GODB driver: treat native Go variables compiled in the
// binary as database tables, all stored in an in-memory SQLite database at Open() time.
func RunExample() error {
	// Create a GODB-driver database connection.
	// Style A (map) config is used here.
	config := neat.DBConfig{
		Default: "go_db",
		Connections: map[string]neat.ConnectionConfig{
			"go_db": {
				Driver: contractsdb.DriverGODB,
				Tables: driver.Tables{
					"blogs":      data.Blogs,
					"categories": data.Categories,
				},
			},
		},
	}

	database, err := neat.New(config)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			fmt.Printf("failed to close database: %v\n", err)
		}
	}()

	// Query blogs in Category #2 — "blogs" table
	fmt.Println("=== Blogs in Category #2 (WHERE category_id = 2) ===")

	var blogs []data.Blog
	err = database.Query().
		Model(&data.Blog{}).
		Where("category_id = ?", 2).
		OrderBy("id", "asc").
		Get(&blogs)
	if err != nil {
		return fmt.Errorf("failed to query blogs: %w", err)
	}

	for _, b := range blogs {
		fmt.Printf("  Blog #%d: %s (Category #%d)\n", b.ID, b.Title, b.CategoryID)
	}

	// JOIN across both tables — blogs and categories
	fmt.Println("\n=== Blogs with Category Name (LEFT JOIN) ===")

	var results []BlogWithCategory
	err = database.Query().
		Table("blogs").
		LeftJoin("categories ON blogs.category_id = categories.id").
		Select("blogs.id, blogs.title, categories.name AS category_name").
		OrderBy("blogs.id", "asc").
		Get(&results)
	if err != nil {
		return fmt.Errorf("failed to query join: %w", err)
	}

	for _, r := range results {
		fmt.Printf("  Blog #%d: %s [%s]\n", r.ID, r.Title, r.CategoryName)
	}

	// Count aggregate
	var count int64
	err = database.Query().Table("blogs").Count(&count)
	if err != nil {
		return fmt.Errorf("failed to query count: %w", err)
	}
	fmt.Printf("\nTotal blogs: %d\n", count)

	return nil
}
