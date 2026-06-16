package main_test

import (
	"testing"

	"github.com/dracory/neat"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	mainpkg "github.com/dracory/neat/examples/schemer-migrations"
)

func TestRunSchemerBasedMigrations(t *testing.T) {
	// Use in-memory SQLite for testing
	err := mainpkg.RunSchemerBasedMigrations("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunSchemerBasedMigrations failed: %v", err)
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
		&mainpkg.CreateMigrationTrackerTable{},
		&mainpkg.CreateUsersTable{},
		&mainpkg.CreatePostsTable{},
		&mainpkg.CreateCommentsTable{},
		&mainpkg.AddPostsIndexes{},
		&mainpkg.AddPublishedToPosts{},
	}

	// Inject schema into migrations
	schema := db.Schema()
	for _, m := range migrations {
		 m.SetSchema(schema)
	}

	// Run all migrations
	for _, migration := range migrations {
		t.Logf("Running migration: %s - %s", migration.Signature(), migration.Description())
		if err := migration.Up(); err != nil {
			t.Fatalf("migration %s failed: %v", migration.Signature(), err)
		}
	}

	// Verify tables exist
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
		&mainpkg.CreateMigrationTrackerTable{},
		&mainpkg.CreateUsersTable{},
		&mainpkg.CreatePostsTable{},
		&mainpkg.AddPublishedToPosts{},
	}

	// Inject schema into migrations
	schema := db.Schema()
	for _, m := range migrations {
		 m.SetSchema(schema)
	}

	// Run all migrations
	for _, migration := range migrations {
		t.Logf("Running migration: %s - %s", migration.Signature(), migration.Description())
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

	// Inject schema into migrations
	schema := db.Schema()
	for _, m := range []contractsschema.MigrationInterface{&mainpkg.CreateMigrationTrackerTable{}, migration} {
		 m.SetSchema(schema)
	}

	// Run migration
	t.Logf("Running migration: %s - %s", migration.Signature(), migration.Description())
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
		&mainpkg.CreateMigrationTrackerTable{},
		&mainpkg.CreateUsersTable{},
		&mainpkg.CreatePostsTable{},
		&mainpkg.CreateCommentsTable{},
	}

	// Inject schema into migrations
	schema := db.Schema()
	for _, m := range migrations {
		 m.SetSchema(schema)
	}

	// Run migrations in order
	for i, migration := range migrations {
		t.Logf("Running migration %d: %s - %s", i, migration.Signature(), migration.Description())
		if err := migration.Up(); err != nil {
			t.Fatalf("migration %d (%s) failed: %v", i, migration.Signature(), err)
		}
	}

	// Verify all tables exist
	tables := []string{"migration_tracker", "users", "posts", "comments"}
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
