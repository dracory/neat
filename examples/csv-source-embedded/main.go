package main

import (
	"embed"
	"fmt"
	"log"

	"github.com/dracory/neat"
	_ "modernc.org/sqlite"
)

//go:embed data/*.csv
var csvFS embed.FS

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

// RunExample demonstrates querying an embedded CSV file via NewCsvFSSource.
func RunExample() error {
	// Create an array-driver database connection
	config := neat.DBConfig{
		Default: "array_db",
		Connections: map[string]neat.ConnectionConfig{
			"array_db": {
				Driver: "array",
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

	// Query the embedded CSV file — NewCsvFSSource reads from embed.FS
	// and wraps it as an ArraySource. The table name is derived from the filename.
	// "data/users.csv" → table "users"
	fmt.Println("=== NewCsvFSSource — query an embedded CSV file ===")

	var users []User
	err = database.Query().
		Model(neat.NewCsvFSSource(csvFS, "data/users.csv")).
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
		Model(neat.NewCsvFSSource(csvFS, "data/users.csv")).
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
