package main

import (
	"fmt"
	"log"

	"github.com/dracory/neat"
	_ "modernc.org/sqlite"
)

// StatusSource implements ArraySource to provide static status data.
type StatusSource struct {
}

func (s *StatusSource) TableName() string {
	return "statuses"
}

func (s *StatusSource) Rows() ([]map[string]any, error) {
	return []map[string]any{
		{"id": 1, "name": "Pending", "color": "yellow"},
		{"id": 2, "name": "Active", "color": "green"},
		{"id": 3, "name": "Inactive", "color": "red"},
	}, nil
}

// Status represents the model for querying the statuses table.
type Status struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Color string `db:"color"`
}

func main() {
	if err := RunExample(); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

// RunExample demonstrates usage of the array driver for static data.
func RunExample() error {
	// Configure connection with the 'array' driver
	config := neat.DBConfig{
		Default: "array_db",
		Connections: map[string]neat.ConnectionConfig{
			"array_db": {
				Driver:   "array",
				Database: ":memory:",
			},
		},
	}

	database, err := neat.New(config)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	defer database.Close()

	// Querying the array-backed model
	fmt.Println("=== Querying Array-Backed Data ===")
	var statuses []Status

	// Just by passing the ArraySource to Model(), Neat automatically populates the table
	err = database.Query().Model(&StatusSource{}).OrderBy("id", "asc").Get(&statuses)
	if err != nil {
		return fmt.Errorf("failed to query statuses: %w", err)
	}

	for _, s := range statuses {
		fmt.Printf("Status #%d: %s (Color: %s)\n", s.ID, s.Name, s.Color)
	}

	// You can use all standard query builder methods
	fmt.Println("\n=== Filtering Array Data ===")
	var activeStatus Status
	err = database.Query().Model(&StatusSource{}).Where("name = ?", "Active").First(&activeStatus)
	if err != nil {
		return fmt.Errorf("failed to find active status: %w", err)
	}
	fmt.Printf("Active status color: %s\n", activeStatus.Color)

	return nil
}
