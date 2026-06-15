package main_test

import (
	"testing"

	"github.com/dracory/neat"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	mainpkg "github.com/dracory/neat/examples/schema-migrations"
)

func TestRunInterfaceBasedMigrations(t *testing.T) {
	// Use in-memory SQLite for testing
	err := mainpkg.RunInterfaceBasedMigrations("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunInterfaceBasedMigrations failed: %v", err)
	}
}

func TestSchemaMigrations_AllMigrations(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create migration instances
	migrations := []contractsschema.MigrationInterface{
		&mainpkg.CreateUsersTable{},
		&mainpkg.CreatePostsTable{},
		&mainpkg.CreateCommentsTable{},
		&mainpkg.AddPostsIndexes{},
		&mainpkg.AddPublishedToPosts{},
	}

	// Register migrations with schema
	schema := db.Schema()
	schema.Register(migrations)

	// Run all migrations
	for _, migration := range migrations {
		if err := migration.Up(); err != nil {
			t.Fatalf("migration %s failed: %v", migration.Signature(), err)
		}
	}

	// Verify tables exist
	if !schema.HasTable("users") {
		t.Error("Expected 'users' table to exist after migrations")
	}
	if !schema.HasTable("posts") {
		t.Error("Expected 'posts' table to exist after migrations")
	}
	if !schema.HasTable("comments") {
		t.Error("Expected 'comments' table to exist after migrations")
	}

	// Verify columns in users table
	columns, err := schema.GetColumns("users")
	if err != nil {
		t.Fatalf("failed to get columns for users table: %v", err)
	}
	expectedColumns := []string{"id", "name", "email", "password", "status", "created_at", "updated_at"}
	for _, col := range expectedColumns {
		found := false
		for _, c := range columns {
			if c.Name == col {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected column '%s' in users table", col)
		}
	}

	// Verify columns in posts table
	columns, err = schema.GetColumns("posts")
	if err != nil {
		t.Fatalf("failed to get columns for posts table: %v", err)
	}
	expectedColumns = []string{"id", "user_id", "title", "content", "status", "created_at", "updated_at", "published_at"}
	for _, col := range expectedColumns {
		found := false
		for _, c := range columns {
			if c.Name == col {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected column '%s' in posts table", col)
		}
	}
}

func TestSchemaMigrations_Rollback(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create migration instances
	migrations := []contractsschema.MigrationInterface{
		&mainpkg.CreateUsersTable{},
		&mainpkg.CreatePostsTable{},
		&mainpkg.AddPublishedToPosts{},
	}

	// Register migrations with schema
	schema := db.Schema()
	schema.Register(migrations)

	// Run all migrations
	for _, migration := range migrations {
		if err := migration.Up(); err != nil {
			t.Fatalf("migration %s failed: %v", migration.Signature(), err)
		}
	}

	// Verify published_at column exists
	columns, err := schema.GetColumns("posts")
	if err != nil {
		t.Fatalf("failed to get columns for posts table: %v", err)
	}
	hasPublishedAt := false
	for _, c := range columns {
		if c.Name == "published_at" {
			hasPublishedAt = true
			break
		}
	}
	if !hasPublishedAt {
		t.Error("Expected 'published_at' column to exist after migration")
	}

	// Rollback last migration
	lastMigration := migrations[len(migrations)-1]
	if err := lastMigration.Down(); err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	// Verify published_at column no longer exists
	columns, err = schema.GetColumns("posts")
	if err != nil {
		t.Fatalf("failed to get columns for posts table after rollback: %v", err)
	}
	hasPublishedAt = false
	for _, c := range columns {
		if c.Name == "published_at" {
			hasPublishedAt = true
			break
		}
	}
	if hasPublishedAt {
		t.Error("Expected 'published_at' column to be removed after rollback")
	}
}

func TestSchemaMigrations_DropTable(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create migration instance
	migration := &mainpkg.CreateUsersTable{}

	// Register migration with schema
	schema := db.Schema()
	schema.Register([]contractsschema.MigrationInterface{migration})

	// Run migration
	if err := migration.Up(); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Verify table exists
	if !schema.HasTable("users") {
		t.Error("Expected 'users' table to exist after migration")
	}

	// Rollback migration
	if err := migration.Down(); err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	// Verify table no longer exists
	if schema.HasTable("users") {
		t.Error("Expected 'users' table to be removed after rollback")
	}
}

func TestSchemaMigrations_MultipleMigrations(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create migration instances
	migrations := []contractsschema.MigrationInterface{
		&mainpkg.CreateUsersTable{},
		&mainpkg.CreatePostsTable{},
		&mainpkg.CreateCommentsTable{},
	}

	// Register migrations with schema
	schema := db.Schema()
	schema.Register(migrations)

	// Run migrations in order
	for i, migration := range migrations {
		if err := migration.Up(); err != nil {
			t.Fatalf("migration %d (%s) failed: %v", i, migration.Signature(), err)
		}
	}

	// Verify all tables exist
	tables := []string{"users", "posts", "comments"}
	for _, table := range tables {
		if !schema.HasTable(table) {
			t.Errorf("Expected '%s' table to exist after migrations", table)
		}
	}

	// Rollback in reverse order
	for i := len(migrations) - 1; i >= 0; i-- {
		if err := migrations[i].Down(); err != nil {
			t.Fatalf("rollback %d (%s) failed: %v", i, migrations[i].Signature(), err)
		}
	}

	// Verify all tables are removed
	for _, table := range tables {
		if schema.HasTable(table) {
			t.Errorf("Expected '%s' table to be removed after rollback", table)
		}
	}
}
