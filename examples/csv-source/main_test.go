package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dracory/neat"
	_ "modernc.org/sqlite"
)

func TestRunExample(t *testing.T) {
	// Resolve the path to the sample CSV file
	csvPath := filepath.Join("data", "users.csv")
	if _, err := os.Stat(csvPath); err != nil {
		csvPath = filepath.Join("examples", "csv-source", "data", "users.csv")
	}

	// Create array-driver database
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
		Model(neat.NewCsvFileSource(csvPath)).
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
		Model(neat.NewCsvFileSource(csvPath)).
		OrderBy("id", "asc").
		Get(&allUsers)
	if err != nil {
		t.Fatalf("failed to query all: %v", err)
	}
	if len(allUsers) != 5 {
		t.Fatalf("expected 5 total users, got %d", len(allUsers))
	}
}
