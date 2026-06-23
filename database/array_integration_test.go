package database

import (
	"testing"

	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database/db"
)

type userArraySource struct {
	id   int
	name string
}

func (u *userArraySource) TableName() string {
	return "users"
}

func (u *userArraySource) Rows() ([]map[string]any, error) {
	return []map[string]any{
		{"id": 1, "name": "John"},
		{"id": 2, "name": "Jane"},
	}, nil
}

type userModel struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

func TestArrayDriverIntegration(t *testing.T) {
	config := db.DBConfig{
		Default: "array_connection",
		Connections: map[string]db.ConnectionConfig{
			"array_connection": {
				Driver:   "array",
				Database: ":memory:",
			},
		},
	}

	database, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer database.Close()

	var users []userModel
	err = database.Query().Model(&userArraySource{}).Get(&users)
	if err != nil {
		t.Fatalf("failed to get users: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}

	if users[0].Name != "John" || users[1].Name != "Jane" {
		t.Errorf("unexpected user data: %+v", users)
	}

	// Test querying again to ensure it doesn't try to re-create/re-populate and fail
	var users2 []userModel
	err = database.Query().Model(&userArraySource{}).Get(&users2)
	if err != nil {
		t.Fatalf("second query failed: %v", err)
	}
	if len(users2) != 2 {
		t.Errorf("expected 2 users on second query, got %d", len(users2))
	}
}
