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

// User is a plain Go struct. The db tags map struct fields to CSV column names.
type User struct {
	ID     int    `db:"id"`
	Name   string `db:"name"`
	Email  string `db:"email"`
	Active bool   `db:"active"`
}

func main() {
	if err := RunExample(); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

// RunExample demonstrates querying a CSV file via NewCsvSource.
func RunExample() error {
	// Resolve the path to the sample CSV file (relative to this example directory)
	csvPath := filepath.Join("data", "users.csv")
	if _, err := os.Stat(csvPath); err != nil {
		// Fall back to path relative to the test working directory
		csvPath = filepath.Join("examples", "csv-source", "data", "users.csv")
	}

	// Create an array-driver database connection
	config := neat.DBConfig{
		Default: "array_db",
		Connections: map[string]neat.ConnectionConfig{
			"array_db": {
				Driver: contractsdb.DriverArray,
			},
		},
	}

	database, err := neat.New(config)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			fmt.Printf("failed to close database: %v", err)
		}
	}()

	// Query the CSV file directly — NewCsvSource parses it and wraps
	// it as an ArraySource. The table name is derived from the filename.
	// "users.csv" → table "users"
	fmt.Println("=== NewCsvSource — query a CSV file ===")

	var users []User
	err = database.Query().
		Model(neat.NewCsvFileSource(csvPath)).
		Where("active = ?", true).
		OrderBy("name", "asc").
		Get(&users)
	if err != nil {
		return fmt.Errorf("failed to query users: %w", err)
	}

	for _, u := range users {
		fmt.Printf("User #%d: %s <%s> (active: %v)\n", u.ID, u.Name, u.Email, u.Active)
	}

	// Query without filter — all rows
	fmt.Println("\n=== All users ===")

	var allUsers []User
	err = database.Query().
		Model(neat.NewCsvFileSource(csvPath)).
		OrderBy("id", "asc").
		Get(&allUsers)
	if err != nil {
		return fmt.Errorf("failed to query all users: %w", err)
	}

	for _, u := range allUsers {
		fmt.Printf("User #%d: %s (active: %v)\n", u.ID, u.Name, u.Active)
	}

	return nil
}
