package main

import (
	"os"
	"testing"

	"github.com/dracory/neat"
	_ "modernc.org/sqlite"
)

func TestJSONFile(t *testing.T) {
	jsonPath := resolvePath("data", "users.json")
	if _, err := os.Stat(jsonPath); err != nil {
		t.Fatalf("sample JSON file not found: %s", jsonPath)
	}

	config := neat.DBConfig{
		Default: "array_db",
		Connections: map[string]neat.ConnectionConfig{
			"array_db": {Driver: "array"},
		},
	}

	database, err := neat.New(config)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer database.Close()

	// Query active users
	var users []User
	err = database.Query().
		Model(neat.NewJsonFileSource(jsonPath)).
		Where("active = ?", true).
		OrderBy("name", "asc").
		Get(&users)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}

	if len(users) != 3 {
		t.Fatalf("expected 3 active users, got %d", len(users))
	}
	if users[0].Name != "Alice" {
		t.Errorf("expected first user 'Alice', got '%s'", users[0].Name)
	}
	if users[1].Name != "Charlie" {
		t.Errorf("expected second user 'Charlie', got '%s'", users[1].Name)
	}
	if users[2].Name != "Diana" {
		t.Errorf("expected third user 'Diana', got '%s'", users[2].Name)
	}

	// Query all users
	var allUsers []User
	err = database.Query().
		Model(neat.NewJsonFileSource(jsonPath)).
		OrderBy("id", "asc").
		Get(&allUsers)
	if err != nil {
		t.Fatalf("failed to query all: %v", err)
	}
	if len(allUsers) != 5 {
		t.Fatalf("expected 5 total users, got %d", len(allUsers))
	}
}

func TestJSONLFile(t *testing.T) {
	jsonlPath := resolvePath("data", "events.jsonl")
	if _, err := os.Stat(jsonlPath); err != nil {
		t.Fatalf("sample JSONL file not found: %s", jsonlPath)
	}

	config := neat.DBConfig{
		Default: "array_db",
		Connections: map[string]neat.ConnectionConfig{
			"array_db": {Driver: "array"},
		},
	}

	database, err := neat.New(config)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer database.Close()

	// Query purchase events
	var events []Event
	err = database.Query().
		Model(neat.NewJsonFileSource(jsonlPath)).
		Where("type = ?", "purchase").
		OrderBy("id", "asc").
		Get(&events)
	if err != nil {
		t.Fatalf("failed to query events: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 purchase events, got %d", len(events))
	}
	if events[0].Amount != 49.99 {
		t.Errorf("expected first event amount=49.99, got %v", events[0].Amount)
	}
	if events[1].Amount != 12.50 {
		t.Errorf("expected second event amount=12.50, got %v", events[1].Amount)
	}

	// Query all events
	var allEvents []Event
	err = database.Query().
		Model(neat.NewJsonFileSource(jsonlPath)).
		OrderBy("id", "asc").
		Get(&allEvents)
	if err != nil {
		t.Fatalf("failed to query all events: %v", err)
	}
	if len(allEvents) != 5 {
		t.Fatalf("expected 5 total events, got %d", len(allEvents))
	}
}
