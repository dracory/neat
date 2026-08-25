package main

import (
	"testing"

	"github.com/dracory/neat"
	_ "modernc.org/sqlite"
)

func TestRunExample(t *testing.T) {
	if err := RunExample(); err != nil {
		t.Fatalf("RunExample failed: %v", err)
	}
}

func TestJsonSourceEmbeddedQueries(t *testing.T) {
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
	defer func() { _ = database.Close() }()

	// Query active users from embedded JSON
	var users []User
	err = database.Query().
		Model(neat.NewJsonFSSource(jsonFS, "data/users.json")).
		Where("active = ?", true).
		OrderBy("name", "asc").
		Get(&users)
	if err != nil {
		t.Fatalf("failed to query users: %v", err)
	}
	if len(users) != 3 {
		t.Fatalf("expected 3 active users, got %d", len(users))
	}
	if users[0].Name != "Alice" {
		t.Errorf("expected first user 'Alice', got '%s'", users[0].Name)
	}

	// Query all users
	var allUsers []User
	err = database.Query().
		Model(neat.NewJsonFSSource(jsonFS, "data/users.json")).
		OrderBy("id", "asc").
		Get(&allUsers)
	if err != nil {
		t.Fatalf("failed to query all: %v", err)
	}
	if len(allUsers) != 5 {
		t.Fatalf("expected 5 total users, got %d", len(allUsers))
	}

	// Query purchase events from embedded JSONL
	var events []Event
	err = database.Query().
		Model(neat.NewJsonFSSource(jsonFS, "data/events.jsonl")).
		Where("type = ?", "purchase").
		OrderBy("id", "asc").
		Get(&events)
	if err != nil {
		t.Fatalf("failed to query events: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 purchase events, got %d", len(events))
	}
}
