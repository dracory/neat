package main_test

import (
	"testing"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/database/schema"
	mainpkg "github.com/dracory/neat/examples/schema-builder"
)

func TestRunExample(t *testing.T) {
	// Use in-memory SQLite for testing
	err := mainpkg.RunExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunExample failed: %v", err)
	}
}

func TestSchemaBuilder_CreateTable(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	err = db.Schema().Create("users", func(bp schema.Blueprint) {
		bp.ID()
		bp.String("name")
		bp.String("email")
		bp.Integer("age")
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if !db.Schema().HasTable("users") {
		t.Error("expected 'users' table to exist after Create")
	}

	for _, col := range []string{"id", "name", "email", "age"} {
		if !db.Schema().HasColumn("users", col) {
			t.Errorf("expected column '%s' to exist", col)
		}
	}
}

func TestSchemaBuilder_ModifyTable(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	err = db.Schema().Create("users", func(bp schema.Blueprint) {
		bp.ID()
		bp.String("name")
		bp.String("phone")
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Add address, drop phone
	err = db.Schema().Table("users", func(bp schema.Blueprint) {
		bp.String("address")
		bp.DropColumn("phone")
	})
	if err != nil {
		t.Fatalf("Table modify failed: %v", err)
	}

	if !db.Schema().HasColumn("users", "address") {
		t.Error("expected 'address' column to exist after Table modify")
	}
	if db.Schema().HasColumn("users", "phone") {
		t.Error("expected 'phone' column to be dropped after Table modify")
	}
}

func TestSchemaBuilder_DropTable(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	err = db.Schema().Create("users", func(bp schema.Blueprint) {
		bp.ID()
		bp.String("name")
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err = db.Schema().Drop("users")
	if err != nil {
		t.Fatalf("Drop failed: %v", err)
	}

	if db.Schema().HasTable("users") {
		t.Error("expected 'users' table to be gone after Drop")
	}
}

func TestSchemaBuilder_DropIfExists_NonExistent(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	// Must not return an error for a table that does not exist
	err = db.Schema().DropIfExists("nonexistent_table")
	if err != nil {
		t.Errorf("DropIfExists on non-existent table returned error: %v", err)
	}
}
