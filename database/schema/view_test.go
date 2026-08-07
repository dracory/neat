package schema_test

import (
	"testing"

	_ "modernc.org/sqlite"

	"github.com/dracory/neat"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
)

// TestSchemaCreateViewRaw verifies CreateViewRaw with a raw SQL string creates a view
// that is visible via HasView.
func TestSchemaCreateViewRaw(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()

	// Create a base table so the view has something to select from
	if err := schema.Create("users", func(table contractsschema.Blueprint) {
		table.ID()
		table.String("name")
	}); err != nil {
		t.Fatalf("failed to create base table: %v", err)
	}

	// CreateViewRaw with a valid select
	if err := schema.CreateViewRaw("active_users", "select * from users"); err != nil {
		t.Fatalf("CreateViewRaw returned error: %v", err)
	}

	if !schema.HasView("active_users") {
		t.Error("expected view 'active_users' to exist after CreateViewRaw")
	}

	// DropViewIfExists should remove it
	if err := schema.DropViewIfExists("active_users"); err != nil {
		t.Fatalf("DropViewIfExists returned error: %v", err)
	}

	if schema.HasView("active_users") {
		t.Error("expected view 'active_users' to be gone after DropViewIfExists")
	}
}

// TestSchemaCreateViewRawEmpty verifies CreateViewRaw with empty SQL returns an error.
func TestSchemaCreateViewRawEmpty(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()

	if err := schema.CreateViewRaw("empty_view", ""); err == nil {
		t.Error("expected error for empty select SQL, got nil")
	}
}

// TestSchemaDropView verifies DropView removes an existing view.
func TestSchemaDropView(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()

	if err := schema.Create("users", func(table contractsschema.Blueprint) {
		table.ID()
		table.String("name")
	}); err != nil {
		t.Fatalf("failed to create base table: %v", err)
	}

	if err := schema.CreateViewRaw("v_users", "select * from users"); err != nil {
		t.Fatalf("CreateViewRaw returned error: %v", err)
	}

	if !schema.HasView("v_users") {
		t.Fatal("expected view 'v_users' to exist before drop")
	}

	if err := schema.DropView("v_users"); err != nil {
		t.Fatalf("DropView returned error: %v", err)
	}

	if schema.HasView("v_users") {
		t.Error("expected view 'v_users' to be gone after DropView")
	}
}

// TestSchemaCreateViewWithQueryBuilder verifies CreateView with a query builder.
func TestSchemaCreateViewWithQueryBuilder(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	schema := db.Schema()

	if err := schema.Create("users", func(table contractsschema.Blueprint) {
		table.ID()
		table.String("name")
	}); err != nil {
		t.Fatalf("failed to create base table: %v", err)
	}

	// Insert a row so the view has data
	_ = schema.Orm().Query().Table("users").Create(map[string]any{"name": "alice"})

	// Build a select query and pass it to CreateView
	selectQuery := schema.Orm().Query().Table("users")
	if err := schema.CreateView("user_names", selectQuery); err != nil {
		t.Fatalf("CreateView returned error: %v", err)
	}

	if !schema.HasView("user_names") {
		t.Error("expected view 'user_names' to exist after CreateView")
	}

	// Cleanup
	if err := schema.DropViewIfExists("user_names"); err != nil {
		t.Fatalf("DropViewIfExists returned error: %v", err)
	}
}
