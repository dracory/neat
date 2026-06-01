package main_test

import (
	"testing"

	mainpkg "github.com/dracory/neat/examples/seeders"
)

func TestRunExample(t *testing.T) {
	// Use in-memory SQLite for testing
	err := mainpkg.RunExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunExample failed: %v", err)
	}
}

func TestRunExampleWithAssertions(t *testing.T) {
	// Use in-memory SQLite for testing
	db, err := mainpkg.RunExampleForTest("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunExampleForTest failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Verify final row counts: SeedOnce (3) + Seed (3) = 6 total
	var users []map[string]any
	err = db.Query().Table("users").Get(&users)
	if err != nil {
		t.Fatalf("Failed to get users: %v", err)
	}
	if len(users) != 6 {
		t.Errorf("Expected 6 users (3 from SeedOnce + 3 from Seed), got %d", len(users))
	}

	var roles []map[string]any
	err = db.Query().Table("roles").Get(&roles)
	if err != nil {
		t.Fatalf("Failed to get roles: %v", err)
	}
	if len(roles) != 6 {
		t.Errorf("Expected 6 roles (3 from SeedOnce + 3 from Seed), got %d", len(roles))
	}

	// Verify facade behavior
	facade := db.Seeder()
	s := facade.GetSeeder("user_seeder")
	if s == nil {
		t.Error("Expected GetSeeder to return non-nil for 'user_seeder'")
	}
	if s != nil && s.Signature() != "user_seeder" {
		t.Errorf("Expected signature 'user_seeder', got '%s'", s.Signature())
	}

	allSeeders := facade.GetSeeders()
	if len(allSeeders) != 2 {
		t.Errorf("Expected 2 registered seeders, got %d", len(allSeeders))
	}
}
