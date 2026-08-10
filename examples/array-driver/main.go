package main

import (
	"fmt"
	"log"

	"github.com/dracory/neat"
	_ "modernc.org/sqlite"
)

// Status is a plain Go struct — no ArraySource interface, no TableName(), no Rows().
// Just struct tags for column mapping (db > neat > gorm > snake_case).
type Status struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Color string `db:"color"`
}

// LegacyStatusSource implements ArraySource the old way — shown in
// ExampleLegacyInterface for comparison. You don't need this anymore.
type LegacyStatusSource struct{}

func (s *LegacyStatusSource) TableName() string { return "statuses" }

func (s *LegacyStatusSource) Rows() ([]map[string]any, error) {
	return []map[string]any{
		{"id": 1, "name": "Pending", "color": "yellow"},
		{"id": 2, "name": "Active", "color": "green"},
		{"id": 3, "name": "Inactive", "color": "red"},
	}, nil
}

func main() {
	if err := RunExample(); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

// RunExample runs all array driver examples in sequence.
func RunExample() error {
	database, err := newDatabase()
	if err != nil {
		return err
	}
	defer func() {
		if err := database.Close(); err != nil {
			fmt.Printf("failed to close database: %v", err)
		}
	}()

	if err := ExampleStructSlice(database); err != nil {
		return err
	}
	if err := ExampleWhereAndFirst(database); err != nil {
		return err
	}
	if err := ExampleMapSlice(database); err != nil {
		return err
	}
	if err := ExampleCustomTableName(database); err != nil {
		return err
	}
	if err := ExampleLegacyInterface(database); err != nil {
		return err
	}

	return nil
}

// newDatabase creates an in-memory array-driver database connection.
func newDatabase() (*neat.Database, error) {
	return neat.NewMemoryDB()
}

// staticData returns the shared slice of Status structs used across examples.
func staticData() []Status {
	return []Status{
		{ID: 1, Name: "Pending", Color: "yellow"},
		{ID: 2, Name: "Active", Color: "green"},
		{ID: 3, Name: "Inactive", Color: "red"},
	}
}

// ExampleStructSlice demonstrates NewArraySourceFrom with a slice of structs —
// the primary, zero-boilerplate entry point for the array driver.
//
// Pass a slice of structs directly: no custom ArraySource struct, no
// TableName() method, no Rows() method. The table name is auto-generated
// from the struct type, and the column schema is inferred from the data.
//
//	database.Query().
//	    Model(neat.NewArraySourceFrom(statuses)).
//	    OrderBy("id", "asc").
//	    Get(&results)
func ExampleStructSlice(database *neat.Database) error {
	fmt.Println("=== NewArraySourceFrom (struct slice) ===")

	var statuses []Status
	err := database.Query().
		Model(neat.NewArraySourceFrom(staticData())).
		OrderBy("id", "asc").
		Get(&statuses)
	if err != nil {
		return fmt.Errorf("failed to query statuses: %w", err)
	}

	for _, s := range statuses {
		fmt.Printf("Status #%d: %s (Color: %s)\n", s.ID, s.Name, s.Color)
	}
	return nil
}

// ExampleWhereAndFirst demonstrates that all standard query builder methods
// (Where, First, OrWhere, Limit, etc.) work with NewArraySourceFrom just as
// they do with regular database-backed models.
//
//	database.Query().
//	    Model(neat.NewArraySourceFrom(statuses)).
//	    Where("name = ?", "Active").
//	    First(&result)
func ExampleWhereAndFirst(database *neat.Database) error {
	fmt.Println("\n=== Filtering with Where + First ===")

	var activeStatus Status
	err := database.Query().
		Model(neat.NewArraySourceFrom(staticData())).
		Where("name = ?", "Active").
		First(&activeStatus)
	if err != nil {
		return fmt.Errorf("failed to find active status: %w", err)
	}

	fmt.Printf("Active status color: %s\n", activeStatus.Color)
	return nil
}

// ExampleMapSlice demonstrates NewArraySourceFrom with a []map[string]any
// slice instead of a struct slice. This is useful when you build rows
// dynamically at runtime rather than from a fixed struct type.
//
//	database.Query().
//	    Model(neat.NewArraySourceFrom(mapData)).
//	    OrderBy("id", "asc").
//	    Get(&results)
func ExampleMapSlice(database *neat.Database) error {
	fmt.Println("\n=== NewArraySourceFrom (map slice) ===")

	mapData := []map[string]any{
		{"id": 1, "name": "Pending", "color": "yellow"},
		{"id": 2, "name": "Active", "color": "green"},
	}

	var results []Status
	err := database.Query().
		Model(neat.NewArraySourceFrom(mapData)).
		OrderBy("id", "asc").
		Get(&results)
	if err != nil {
		return fmt.Errorf("failed to query map data: %w", err)
	}

	for _, s := range results {
		fmt.Printf("Status #%d: %s (Color: %s)\n", s.ID, s.Name, s.Color)
	}
	return nil
}

// ExampleCustomTableName demonstrates NewArraySource with the .Table() setter
// to override the auto-generated table name. This is useful when you need a
// specific table name for JOINs, raw SQL, or more readable error messages.
//
//	database.Query().
//	    Model(neat.NewArraySource(rows).Table("statuses")).
//	    OrderBy("id", "asc").
//	    Get(&results)
func ExampleCustomTableName(database *neat.Database) error {
	fmt.Println("\n=== NewArraySource with .Table() ===")

	mapData := []map[string]any{
		{"id": 1, "name": "Pending", "color": "yellow"},
		{"id": 2, "name": "Active", "color": "green"},
	}

	var results []Status
	err := database.Query().
		Model(neat.NewArraySource(mapData).Table("statuses")).
		OrderBy("id", "asc").
		Get(&results)
	if err != nil {
		return fmt.Errorf("failed to query with custom table name: %w", err)
	}

	for _, s := range results {
		fmt.Printf("Status #%d: %s (Color: %s)\n", s.ID, s.Name, s.Color)
	}
	return nil
}

// ExampleLegacyInterface demonstrates the original interface-based approach
// where you define a custom struct implementing ArraySource (TableName() +
// Rows()). This still works and is fully supported, but NewArraySourceFrom
// is preferred for new code since it requires no boilerplate.
//
//	type StatusSource struct{}
//	func (s *StatusSource) TableName() string { return "statuses" }
//	func (s *StatusSource) Rows() ([]map[string]any, error) { ... }
//
//	database.Query().
//	    Model(&StatusSource{}).
//	    OrderBy("id", "asc").
//	    Get(&results)
func ExampleLegacyInterface(database *neat.Database) error {
	fmt.Println("\n=== Legacy ArraySource interface (for comparison) ===")

	var statuses []Status
	err := database.Query().
		Model(&LegacyStatusSource{}).
		OrderBy("id", "asc").
		Get(&statuses)
	if err != nil {
		return fmt.Errorf("failed to query legacy statuses: %w", err)
	}

	for _, s := range statuses {
		fmt.Printf("Legacy Status #%d: %s (Color: %s)\n", s.ID, s.Name, s.Color)
	}
	return nil
}
