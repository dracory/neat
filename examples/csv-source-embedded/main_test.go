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

func TestCsvSourceEmbeddedQueries(t *testing.T) {
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

	// Query active users
	var users []User
	err = database.Query().
		Model(neat.NewCsvFSSource(csvFS, "data/users.csv")).
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

	// Query all users
	var allUsers []User
	err = database.Query().
		Model(neat.NewCsvFSSource(csvFS, "data/users.csv")).
		OrderBy("id", "asc").
		Get(&allUsers)
	if err != nil {
		t.Fatalf("failed to query all: %v", err)
	}
	if len(allUsers) != 5 {
		t.Fatalf("expected 5 total users, got %d", len(allUsers))
	}
}
