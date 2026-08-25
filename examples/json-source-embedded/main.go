package main

import (
	"embed"
	"fmt"
	"log"

	"github.com/dracory/neat"
	_ "modernc.org/sqlite"
)

//go:embed data/*.json data/*.jsonl
var jsonFS embed.FS

// User is a plain Go struct. The db tags map struct fields to JSON keys.
type User struct {
	ID     int    `db:"id"`
	Name   string `db:"name"`
	Email  string `db:"email"`
	Active bool   `db:"active"`
}

// Event maps to the events.jsonl file.
type Event struct {
	ID        int     `db:"id"`
	Type      string  `db:"type"`
	UserID    int     `db:"user_id"`
	Timestamp string  `db:"timestamp"`
	Amount    float64 `db:"amount"`
}

func main() {
	if err := RunExample(); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

// RunExample demonstrates querying embedded JSON and JSONL files via NewJsonFSSource.
func RunExample() error {
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

	// Query embedded JSON file — data/users.json → table "users"
	fmt.Println("=== NewJsonFSSource — query an embedded JSON file ===")

	var users []User
	err = database.Query().
		Model(neat.NewJsonFSSource(jsonFS, "data/users.json")).
		Where("active = ?", true).
		OrderBy("name", "asc").
		Get(&users)
	if err != nil {
		return fmt.Errorf("failed to query users: %w", err)
	}

	for _, u := range users {
		fmt.Printf("User #%d: %s <%s> (active: %v)\n", u.ID, u.Name, u.Email, u.Active)
	}

	// Query embedded JSONL file — data/events.jsonl → table "events"
	fmt.Println("\n=== NewJsonFSSource — query an embedded JSONL file ===")

	var events []Event
	err = database.Query().
		Model(neat.NewJsonFSSource(jsonFS, "data/events.jsonl")).
		Where("type = ?", "purchase").
		OrderBy("id", "asc").
		Get(&events)
	if err != nil {
		return fmt.Errorf("failed to query events: %w", err)
	}

	for _, e := range events {
		fmt.Printf("Event #%d: %s by user %d ($%.2f)\n", e.ID, e.Type, e.UserID, e.Amount)
	}

	return nil
}
