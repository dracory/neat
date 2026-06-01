package main_test

import (
	"testing"

	"github.com/dracory/neat"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	mainpkg "github.com/dracory/neat/examples/migrations"
)

func TestRunSchemaBuilderExample(t *testing.T) {
	// Use in-memory SQLite for testing
	err := mainpkg.RunSchemaBuilderExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunSchemaBuilderExample failed: %v", err)
	}
}

func TestRunMigrationSystemExample(t *testing.T) {
	// Use in-memory SQLite for testing
	err := mainpkg.RunMigrationSystemExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunMigrationSystemExample failed: %v", err)
	}
}

func TestSchemaBuilder_TablesExist(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err = db.Schema().Create("users", func(bp contractsschema.Blueprint) {
		bp.ID()
		bp.String("name")
		bp.String("email")
		bp.Timestamps()
		bp.SoftDeletes()
	}); err != nil {
		t.Fatalf("Create users failed: %v", err)
	}

	if err = db.Schema().Create("posts", func(bp contractsschema.Blueprint) {
		bp.ID()
		bp.Integer("user_id")
		bp.String("title")
		bp.Text("content")
		bp.Timestamps()
	}); err != nil {
		t.Fatalf("Create posts failed: %v", err)
	}

	if err = db.Schema().Create("comments", func(bp contractsschema.Blueprint) {
		bp.ID()
		bp.Integer("post_id")
		bp.Integer("user_id")
		bp.Text("comment")
		bp.Timestamps()
	}); err != nil {
		t.Fatalf("Create comments failed: %v", err)
	}

	for _, tbl := range []string{"users", "posts", "comments"} {
		if !db.Schema().HasTable(tbl) {
			t.Errorf("expected table '%s' to exist after schema builder Create", tbl)
		}
	}

	if !db.Schema().HasColumn("users", "name") {
		t.Error("expected 'name' column on users")
	}
	if !db.Schema().HasColumn("posts", "user_id") {
		t.Error("expected 'user_id' column on posts")
	}
	if !db.Schema().HasColumn("comments", "post_id") {
		t.Error("expected 'post_id' column on comments")
	}
}

func TestMigrationSystem_AllRan(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err = db.Migrate(); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	status, err := db.MigrationStatus()
	if err != nil {
		t.Fatalf("MigrationStatus failed: %v", err)
	}

	if len(status) == 0 {
		t.Fatal("expected at least one migration in status, got 0")
	}

	for _, s := range status {
		if !s.Ran {
			t.Errorf("expected migration '%s' to be marked as ran, but it was not", s.Name)
		}
	}
}

func TestMigrationSystem_Rollback(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err = db.Migrate(); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	statusBefore, err := db.MigrationStatus()
	if err != nil {
		t.Fatalf("MigrationStatus before rollback failed: %v", err)
	}
	ranBefore := 0
	for _, s := range statusBefore {
		if s.Ran {
			ranBefore++
		}
	}

	if err = db.MigrateDown(1); err != nil {
		t.Fatalf("MigrateDown failed: %v", err)
	}

	statusAfter, err := db.MigrationStatus()
	if err != nil {
		t.Fatalf("MigrationStatus after rollback failed: %v", err)
	}
	ranAfter := 0
	for _, s := range statusAfter {
		if s.Ran {
			ranAfter++
		}
	}

	if ranAfter >= ranBefore {
		t.Errorf("expected fewer ran migrations after rollback: before=%d, after=%d", ranBefore, ranAfter)
	}
}
