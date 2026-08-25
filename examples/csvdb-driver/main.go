package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	contractsdb "github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat"
	_ "modernc.org/sqlite"
)

// User maps to the users.csv table.
type User struct {
	ID     int    `db:"id"`
	Name   string `db:"name"`
	Email  string `db:"email"`
	Active bool   `db:"active"`
}

func (u *User) TableName() string { return "users" }

// Product maps to the products.csv table.
type Product struct {
	ID       int     `db:"id"`
	Name     string  `db:"name"`
	Price    float64 `db:"price"`
	Category string  `db:"category"`
}

func (p *Product) TableName() string { return "products" }

// OrderWithUser is a view model for JOIN queries across orders and users.
type OrderWithUser struct {
	ID       int     `db:"id"`
	UserName string  `db:"user_name"`
	Total    float64 `db:"total"`
}

func main() {
	if err := RunExample(); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

// RunExample demonstrates the CSVDB driver: point at a directory of CSV
// files and query them as if they were database tables.
func RunExample() error {
	// Resolve the data directory (relative to this example directory)
	dataDir := filepath.Join("data")
	if _, err := os.Stat(dataDir); err != nil {
		// Fall back to path relative to the test working directory
		dataDir = filepath.Join("examples", "csvdb-driver", "data")
	}

	// Create a CSVDB-driver database connection.
	// The directory is the database; each .csv file becomes a table.
	config := neat.DBConfig{
		Default: "csv_db",
		Connections: map[string]neat.ConnectionConfig{
			"csv_db": {
				Driver: contractsdb.DriverCSVDB,
				Database: dataDir,
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

	// Query active users — data/users.csv → "users" table
	fmt.Println("=== Active users (WHERE active = true) ===")

	var users []User
	err = database.Query().
		Model(&User{}).
		Where("active = ?", true).
		OrderBy("name", "asc").
		Get(&users)
	if err != nil {
		return fmt.Errorf("failed to query users: %w", err)
	}

	for _, u := range users {
		fmt.Printf("  User #%d: %s <%s>\n", u.ID, u.Name, u.Email)
	}

	// Query expensive products — data/products.csv → "products" table
	fmt.Println("\n=== Expensive products (price > 50) ===")

	var products []Product
	err = database.Query().
		Model(&Product{}).
		Where("price > ?", 50).
		OrderBy("price", "desc").
		Get(&products)
	if err != nil {
		return fmt.Errorf("failed to query products: %w", err)
	}

	for _, p := range products {
		fmt.Printf("  Product #%d: %s ($%.2f, %s)\n", p.ID, p.Name, p.Price, p.Category)
	}

	// JOIN across two CSV-backed tables — orders + users
	fmt.Println("\n=== Orders with user names (JOIN) ===")

	var results []OrderWithUser
	err = database.Query().
		Table("orders AS o").
		LeftJoin("users AS u ON o.user_id = u.id").
		Select("o.id, u.name AS user_name, o.total").
		OrderBy("o.id", "asc").
		Get(&results)
	if err != nil {
		return fmt.Errorf("failed to query join: %w", err)
	}

	for _, r := range results {
		fmt.Printf("  Order #%d: %s ordered $%.2f\n", r.ID, r.UserName, r.Total)
	}

	return nil
}
