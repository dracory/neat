package main

import (
	"os"
	"testing"

	"github.com/dracory/neat"
	_ "modernc.org/sqlite"
)

func TestXMLFile(t *testing.T) {
	xmlPath := resolvePath("data", "users.xml")
	if _, err := os.Stat(xmlPath); err != nil {
		t.Fatalf("sample XML file not found: %s", xmlPath)
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
		Model(neat.NewXmlFileSource(xmlPath)).
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
		Model(neat.NewXmlFileSource(xmlPath)).
		OrderBy("id", "asc").
		Get(&allUsers)
	if err != nil {
		t.Fatalf("failed to query all: %v", err)
	}
	if len(allUsers) != 5 {
		t.Fatalf("expected 5 total users, got %d", len(allUsers))
	}

	// Verify ID is parsed as int from attribute
	if allUsers[0].ID != 1 {
		t.Errorf("expected first user ID=1, got %d", allUsers[0].ID)
	}
	// Verify created is parsed as time.Time
	if allUsers[0].Created.IsZero() {
		t.Errorf("expected created to be parsed as time.Time, got zero value")
	}
}
