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

// User is a second struct used to demonstrate JOINs across sources
// in the same in-memory database.
type User struct {
	ID     int    `db:"id"`
	Name   string `db:"name"`
	Status string `db:"status"`
}

func main() {
	if err := RunExample(); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

// RunExample runs all NewMemoryDB examples in sequence.
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
	if err := ExampleMultiTableJoin(database); err != nil {
		return err
	}

	return nil
}

// newDatabase creates an in-memory database using the zero-config NewMemoryDB
// constructor. No DBConfig, no connection map, no driver string — just call
// neat.NewMemoryDB() and start querying.
func newDatabase() (*neat.Database, error) {
	return neat.NewMemoryDB()
}

// staticStatuses returns the shared slice of Status structs.
func staticStatuses() []Status {
	return []Status{
		{ID: 1, Name: "Pending", Color: "yellow"},
		{ID: 2, Name: "Active", Color: "green"},
		{ID: 3, Name: "Inactive", Color: "red"},
	}
}

// staticUsers returns the shared slice of User structs.
func staticUsers() []User {
	return []User{
		{ID: 1, Name: "Alice", Status: "Active"},
		{ID: 2, Name: "Bob", Status: "Pending"},
		{ID: 3, Name: "Charlie", Status: "Active"},
	}
}

// ExampleStructSlice demonstrates the simplest usage: create an in-memory
// database with NewMemoryDB, load a struct slice as a table, and query it.
//
//	database, _ := neat.NewMemoryDB()
//	defer database.Close()
//
//	database.Query().
//	    Model(neat.NewArraySourceFrom(statuses)).
//	    OrderBy("id", "asc").
//	    Get(&results)
func ExampleStructSlice(database *neat.Database) error {
	fmt.Println("=== NewMemoryDB: struct slice query ===")

	var statuses []Status
	err := database.Query().
		Model(neat.NewArraySourceFrom(staticStatuses())).
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

// ExampleWhereAndFirst demonstrates filtering with Where and First on an
// in-memory database — same query builder API as real databases.
//
//	database.Query().
//	    Model(neat.NewArraySourceFrom(statuses)).
//	    Where("name = ?", "Active").
//	    First(&result)
func ExampleWhereAndFirst(database *neat.Database) error {
	fmt.Println("\n=== NewMemoryDB: Where + First ===")

	var activeStatus Status
	err := database.Query().
		Model(neat.NewArraySourceFrom(staticStatuses())).
		Where("name = ?", "Active").
		First(&activeStatus)
	if err != nil {
		return fmt.Errorf("failed to find active status: %w", err)
	}

	fmt.Printf("Active status color: %s\n", activeStatus.Color)
	return nil
}

// ExampleMultiTableJoin demonstrates loading multiple sources into the same
// in-memory database and JOINing across them. Each Model() call populates a
// new table in the same SQLite database, enabling cross-source queries.
//
// Use NewArraySource().Table("name") to set explicit table names when you
// need to reference them in JOINs — NewArraySourceFrom auto-generates
// unique table names (e.g., "array_user_1") that are hard to predict.
//
//	database.Query().Model(neat.NewArraySource(statusRows).Table("statuses")).Get(&_)
//	database.Query().Model(neat.NewArraySource(userRows).Table("users")).Get(&_)
//
//	database.Query().
//	    Table("users").
//	    LeftJoin("statuses", "users.status = statuses.name").
//	    Select("users.id", "users.name", "statuses.color").
//	    Get(&results)
func ExampleMultiTableJoin(database *neat.Database) error {
	fmt.Println("\n=== NewMemoryDB: multi-table JOIN ===")

	// Populate both tables with explicit names in the same in-memory SQLite database
	statusRows := []map[string]any{
		{"id": 1, "name": "Pending", "color": "yellow"},
		{"id": 2, "name": "Active", "color": "green"},
		{"id": 3, "name": "Inactive", "color": "red"},
	}
	var dummyStatuses []Status
	if err := database.Query().
		Model(neat.NewArraySource(statusRows).Table("statuses")).
		Get(&dummyStatuses); err != nil {
		return fmt.Errorf("failed to populate statuses: %w", err)
	}

	userRows := []map[string]any{
		{"id": 1, "name": "Alice", "status": "Active"},
		{"id": 2, "name": "Bob", "status": "Pending"},
		{"id": 3, "name": "Charlie", "status": "Active"},
	}
	var dummyUsers []User
	if err := database.Query().
		Model(neat.NewArraySource(userRows).Table("users")).
		Get(&dummyUsers); err != nil {
		return fmt.Errorf("failed to populate users: %w", err)
	}

	// JOIN across both tables
	type UserWithColor struct {
		ID    int    `db:"id"`
		Name  string `db:"name"`
		Color string `db:"color"`
	}

	var results []UserWithColor
	q := database.Query().
		Table("users").
		LeftJoin("statuses ON users.status = statuses.name").
		Select("users.id", "users.name", "statuses.color").
		OrderBy("users.id", "asc")

	err := q.Get(&results)
	if err != nil {
		return fmt.Errorf("failed to join users and statuses: %w", err)
	}

	for _, r := range results {
		fmt.Printf("User #%d: %s (Status color: %s)\n", r.ID, r.Name, r.Color)
	}
	return nil
}
