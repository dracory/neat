package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	contractsdb "github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat"
	_ "modernc.org/sqlite"
)

// User is a plain Go struct. The db tags map struct fields to JSON keys.
type User struct {
	ID      int       `db:"id"`
	Name    string    `db:"name"`
	Email   string    `db:"email"`
	Active  bool      `db:"active"`
	Created time.Time `db:"created"`
}

// Event is a plain Go struct for the JSONL file.
type Event struct {
	ID        int       `db:"id"`
	Type      string    `db:"type"`
	UserID    int       `db:"user_id"`
	Timestamp time.Time `db:"timestamp"`
	Amount    float64   `db:"amount"`
}

func main() {
	if err := RunExample(); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

// RunExample demonstrates querying JSON and JSONL files via NewJsonSource.
func RunExample() error {
	// Resolve the path to the sample data files (relative to this example directory)
	jsonPath := resolvePath("data", "users.json")
	jsonlPath := resolvePath("data", "events.jsonl")

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

	// 1. Query a JSON file — NewJsonSource parses the array of objects.
	//    The table name is derived from the filename: "users.json" → "users"
	fmt.Println("=== NewJsonSource — query a JSON file ===")

	var users []User
	err = database.Query().
		Model(neat.NewJsonFileSource(jsonPath)).
		Where("active = ?", true).
		OrderBy("name", "asc").
		Get(&users)
	if err != nil {
		return fmt.Errorf("failed to query users: %w", err)
	}

	for _, u := range users {
		fmt.Printf("User #%d: %s <%s> (created: %s)\n", u.ID, u.Name, u.Email, u.Created.Format("2006-01-02"))
	}

	// 2. Query a JSONL file — one JSON object per line.
	//    "events.jsonl" → table "events"
	fmt.Println("\n=== NewJsonSource — query a JSONL file ===")

	var events []Event
	err = database.Query().
		Model(neat.NewJsonFileSource(jsonlPath)).
		Where("type = ?", "purchase").
		OrderBy("id", "asc").
		Get(&events)
	if err != nil {
		return fmt.Errorf("failed to query events: %w", err)
	}

	for _, e := range events {
		if e.Amount > 0 {
			fmt.Printf("Event #%d: %s (user: %d, amount: $%.2f)\n", e.ID, e.Type, e.UserID, e.Amount)
		} else {
			fmt.Printf("Event #%d: %s (user: %d)\n", e.ID, e.Type, e.UserID)
		}
	}

	// 3. Query all events
	fmt.Println("\n=== All events ===")

	var allEvents []Event
	err = database.Query().
		Model(neat.NewJsonFileSource(jsonlPath)).
		OrderBy("id", "asc").
		Get(&allEvents)
	if err != nil {
		return fmt.Errorf("failed to query all events: %w", err)
	}

	for _, e := range allEvents {
		fmt.Printf("Event #%d: %s by user %d at %s\n", e.ID, e.Type, e.UserID, e.Timestamp.Format("2006-01-02 15:04"))
	}

	return nil
}

// resolvePath tries a path relative to the working directory, then falls
// back to a path relative to the repo root.
func resolvePath(parts ...string) string {
	p := filepath.Join(parts...)
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return filepath.Join(append([]string{"examples", "json-source"}, parts...)...)
}
