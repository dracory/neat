package database

import (
	"testing"
	"time"

	"github.com/dracory/neat/contracts/log"
	contractsdb "github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat/database/db"
	"github.com/dracory/neat/support/arraysource"
)

type userArraySource struct {
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
				Driver: contractsdb.DriverArray,
				Database: ":memory:",
			},
		},
	}

	database, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer func() {
		_ = database.Close()
	}()

	var users []userModel
	err = database.Query().Model(&userArraySource{}).Get(&users)
	if err != nil {
		t.Fatalf("failed to get users: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
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

type modelWithNullableFields struct {
	ID      int        `db:"id"`
	Name    *string    `db:"name"`
	Age     *int       `db:"age"`
	Created *time.Time `db:"created"`
}

func TestNewArraySourceFrom_NullableFields_Integration(t *testing.T) {
	config := db.DBConfig{
		Default: "array_connection",
		Connections: map[string]db.ConnectionConfig{
			"array_connection": {
				Driver: contractsdb.DriverArray,
				Database: ":memory:",
			},
		},
	}

	database, err := New(config, WithLogger(log.NewNoopLogger()))
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer func() {
		_ = database.Close()
	}()

	strVal := "Alice"
	ageVal := 30
	now := time.Now().Truncate(time.Second)

	staticData := []modelWithNullableFields{
		{ID: 1, Name: &strVal, Age: &ageVal, Created: &now},
		{ID: 2, Name: nil, Age: nil, Created: nil},
	}

	var results []modelWithNullableFields
	err = database.Query().Model(arraysource.NewArraySourceFrom(staticData)).OrderBy("id", "asc").Get(&results)
	if err != nil {
		t.Fatalf("failed to query static data: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Verify non-null row
	r1 := results[0]
	if r1.ID != 1 {
		t.Errorf("expected ID 1, got %d", r1.ID)
	}
	if r1.Name == nil || *r1.Name != "Alice" {
		t.Errorf("expected Name 'Alice', got %v", r1.Name)
	}
	if r1.Age == nil || *r1.Age != 30 {
		t.Errorf("expected Age 30, got %v", r1.Age)
	}
	if r1.Created == nil || !r1.Created.Equal(now) {
		t.Errorf("expected Created %v, got %v", now, r1.Created)
	}

	// Verify null row
	r2 := results[1]
	if r2.ID != 2 {
		t.Errorf("expected ID 2, got %d", r2.ID)
	}
	if r2.Name != nil {
		t.Errorf("expected Name nil, got %v", *r2.Name)
	}
	if r2.Age != nil {
		t.Errorf("expected Age nil, got %v", *r2.Age)
	}
	if r2.Created != nil {
		t.Errorf("expected Created nil, got %v", r2.Created)
	}
}
