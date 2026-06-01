package main_test

import (
	"testing"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/database/schema"
	mainpkg "github.com/dracory/neat/examples/basic-orm"
)

func TestRunExample(t *testing.T) {
	// Use in-memory SQLite for testing
	err := mainpkg.RunExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunExample failed: %v", err)
	}
}

func TestBasicORM_CreateAndQuery(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	err = db.Schema().Create("users", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.String("email")
		blueprint.Integer("age")
		blueprint.String("status")
		blueprint.Timestamp("created_at")
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Table must exist
	if !db.Schema().HasTable("users") {
		t.Fatal("Expected 'users' table to exist after Create")
	}

	// Insert a record
	err = db.Query().Table("users").Create(map[string]any{
		"name": "John Doe", "email": "john@example.com",
		"age": 30, "status": "active", "created_at": "2026-05-12 18:00:00",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v", err)
	}

	// Record count must be 1
	var count int64
	err = db.Query().Table("users").Count(&count)
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 user after create, got %d", count)
	}

	// Update the record
	_, err = db.Query().Table("users").Where("id = ?", 1).Update(map[string]any{"name": "Jane Doe"})
	if err != nil {
		t.Fatalf("failed to update: %v", err)
	}

	// Verify update
	var user map[string]any
	err = db.Query().Table("users").Where("id = ?", 1).Get(&user)
	if err != nil {
		t.Fatalf("failed to get updated user: %v", err)
	}
	if user["name"] != "Jane Doe" {
		t.Errorf("expected name 'Jane Doe' after update, got '%v'", user["name"])
	}

	// Delete the record
	_, err = db.Query().Table("users").Where("id = ?", 1).Delete()
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	// Count must be 0 after delete
	err = db.Query().Table("users").Count(&count)
	if err != nil {
		t.Fatalf("failed to count after delete: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 users after delete, got %d", count)
	}
}

func TestBasicORM_AdvancedQuery(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	err = db.Schema().Create("users", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.String("email")
		blueprint.Integer("age")
		blueprint.String("status")
		blueprint.Timestamp("created_at")
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Seed data: 2 active adults, 1 minor, 1 inactive adult
	rows := []map[string]any{
		{"name": "Alice", "email": "alice@example.com", "age": 25, "status": "active", "created_at": "2026-01-01 00:00:00"},
		{"name": "Bob", "email": "bob@example.com", "age": 30, "status": "active", "created_at": "2026-01-02 00:00:00"},
		{"name": "Charlie", "email": "charlie@example.com", "age": 16, "status": "active", "created_at": "2026-01-03 00:00:00"},
		{"name": "Dave", "email": "dave@example.com", "age": 40, "status": "inactive", "created_at": "2026-01-04 00:00:00"},
	}
	for _, row := range rows {
		if err = db.Query().Table("users").Create(row); err != nil {
			t.Fatalf("failed to seed: %v", err)
		}
	}

	// Only active users over 18: Alice and Bob
	var results []map[string]any
	err = db.Query().Table("users").
		Where("age > ?", 18).
		Where("status = ?", "active").
		OrderBy("created_at", "desc").
		Limit(10).
		Get(&results)
	if err != nil {
		t.Fatalf("advanced query failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 active users over 18, got %d", len(results))
	}
}
