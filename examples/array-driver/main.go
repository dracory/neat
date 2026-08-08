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

	// Querying the array-backed model using the legacy interface-based approach
	fmt.Println("=== Legacy Interface-Backed Query ===")
	var legacyStatuses []Status
	err = database.Query().Model(&StatusSource{}).OrderBy("id", "asc").Get(&legacyStatuses)
	if err != nil {
		return fmt.Errorf("failed to query legacy statuses: %w", err)
	}
	for _, s := range legacyStatuses {
		fmt.Printf("Legacy Status #%d: %s (Color: %s)\n", s.ID, s.Name, s.Color)
	}

	// Querying using the new NewArraySourceFrom slice helper (primary, day-to-day entry point)
	fmt.Println("\n=== New Slice-Backed Query (NewArraySourceFrom) ===")
	staticData := []Status{
		{ID: 1, Name: "Pending", Color: "yellow"},
		{ID: 2, Name: "Active", Color: "green"},
		{ID: 3, Name: "Inactive", Color: "red"},
	}

	var statuses []Status
	err = database.Query().Model(neat.NewArraySourceFrom(staticData)).OrderBy("id", "asc").Get(&statuses)
	if err != nil {
		return fmt.Errorf("failed to query statuses: %w", err)
	}

	for _, s := range statuses {
		fmt.Printf("Status #%d: %s (Color: %s)\n", s.ID, s.Name, s.Color)
	}

	// You can use all standard query builder methods with the helper
	fmt.Println("\n=== Filtering Array Data with Slice Helper ===")
	var activeStatus Status
	err = database.Query().Model(neat.NewArraySourceFrom(staticData)).Where("name = ?", "Active").First(&activeStatus)
	if err != nil {
		return fmt.Errorf("failed to find active status: %w", err)
	}
	fmt.Printf("Active status color: %s\n", activeStatus.Color)

	return nil
}
