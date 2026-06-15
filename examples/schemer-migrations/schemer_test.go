package main_test

import (
	"context"
	"testing"

	"github.com/dracory/neat"
	"github.com/dracory/neat/database/schemer"
	mainpkg "github.com/dracory/neat/examples/schemer-migrations"
)

func TestSchemerPackage_AllMigrations(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create schemer instance
	schemerInstance := schemer.NewSchemer(db)

	// Add migrations
	schemerInstance.AddMigration(&mainpkg.CreateMigrationTrackerTable{})
	schemerInstance.AddMigration(&mainpkg.CreateUsersTable{})
	schemerInstance.AddMigration(&mainpkg.CreatePostsTable{})
	schemerInstance.AddMigration(&mainpkg.CreateCommentsTable{})
	schemerInstance.AddMigration(&mainpkg.AddPostsIndexes{})
	schemerInstance.AddMigration(&mainpkg.AddPublishedToPosts{})

	// Run migrations
	ctx := context.Background()
	if err := schemerInstance.Up(ctx); err != nil {
		t.Fatalf("migration up failed: %v", err)
	}

	// Verify tables exist
	schema := db.Schema()
	if !schema.HasTable("migration_tracker") {
		t.Error("Expected 'migration_tracker' table to exist after migrations")
	}
	if !schema.HasTable("users") {
		t.Error("Expected 'users' table to exist after migrations")
	}
	if !schema.HasTable("posts") {
		t.Error("Expected 'posts' table to exist after migrations")
	}
	if !schema.HasTable("comments") {
		t.Error("Expected 'comments' table to exist after migrations")
	}

	// Check migration status
	status, err := schemerInstance.Status()
	if err != nil {
		t.Fatalf("failed to get migration status: %v", err)
	}
	if len(status) != 6 {
		t.Errorf("Expected 6 migration statuses, got %d", len(status))
	}

	// Rollback last migration
	if err := schemerInstance.Down(ctx); err != nil {
		t.Fatalf("migration down failed: %v", err)
	}
}

func TestSchemerPackage_RollbackSteps(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create schemer instance
	schemerInstance := schemer.NewSchemer(db)

	// Add migrations
	schemerInstance.AddMigration(&mainpkg.CreateMigrationTrackerTable{})
	schemerInstance.AddMigration(&mainpkg.CreateUsersTable{})
	schemerInstance.AddMigration(&mainpkg.CreatePostsTable{})

	// Run migrations
	ctx := context.Background()
	if err := schemerInstance.Up(ctx); err != nil {
		t.Fatalf("migration up failed: %v", err)
	}

	// Rollback 2 migrations
	if err := schemerInstance.RollbackSteps(ctx, 2); err != nil {
		t.Fatalf("rollback steps failed: %v", err)
	}
}

func TestSchemerPackage_AddMigrations(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create schemer instance
	schemerInstance := schemer.NewSchemer(db)

	// Add migrations individually
	schemerInstance.AddMigration(&mainpkg.CreateMigrationTrackerTable{})
	schemerInstance.AddMigration(&mainpkg.CreateUsersTable{})

	// Run migrations
	ctx := context.Background()
	if err := schemerInstance.Up(ctx); err != nil {
		t.Fatalf("migration up failed: %v", err)
	}

	// Verify tables exist
	schema := db.Schema()
	if !schema.HasTable("users") {
		t.Error("Expected 'users' table to exist after migrations")
	}
}
