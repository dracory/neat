package postgres_test

import (
	"testing"
)

// TestPostgresSchemaViewCreateRawRoundTrip verifies the full CreateViewRaw → HasView →
// DropViewIfExists → !HasView round-trip on PostgreSQL.
func TestPostgresSchemaViewCreateRawRoundTrip(t *testing.T) {
	db := SetupPostgresTest(t)
	viewName := "test_view_raw"

	_ = db.Schema().DropViewIfExists(viewName)

	if db.Schema().HasView(viewName) {
		t.Fatal("view should not exist before creation")
	}

	if err := db.Schema().CreateViewRaw(viewName, "select * from users"); err != nil {
		t.Fatalf("CreateViewRaw failed: %v", err)
	}

	if !db.Schema().HasView(viewName) {
		t.Error("view should exist after CreateViewRaw")
	}

	if err := db.Schema().DropViewIfExists(viewName); err != nil {
		t.Fatalf("DropViewIfExists failed: %v", err)
	}

	if db.Schema().HasView(viewName) {
		t.Error("view should not exist after DropViewIfExists")
	}
}

// TestPostgresSchemaViewDropView verifies DropView on an existing view.
func TestPostgresSchemaViewDropView(t *testing.T) {
	db := SetupPostgresTest(t)
	viewName := "test_view_drop"

	_ = db.Schema().DropViewIfExists(viewName)

	if err := db.Schema().CreateViewRaw(viewName, "select * from users"); err != nil {
		t.Fatalf("CreateViewRaw failed: %v", err)
	}

	if !db.Schema().HasView(viewName) {
		t.Fatal("view should exist before DropView")
	}

	if err := db.Schema().DropView(viewName); err != nil {
		t.Fatalf("DropView failed: %v", err)
	}

	if db.Schema().HasView(viewName) {
		t.Error("view should not exist after DropView")
	}
}

// TestPostgresSchemaViewCreateWithQuery verifies CreateView with a query builder.
func TestPostgresSchemaViewCreateWithQuery(t *testing.T) {
	db := SetupPostgresTest(t)
	viewName := "test_view_query"

	_ = db.Schema().DropViewIfExists(viewName)

	selectQuery := db.Schema().Orm().Query().Table("users")
	if err := db.Schema().CreateView(viewName, selectQuery); err != nil {
		t.Fatalf("CreateView failed: %v", err)
	}

	if !db.Schema().HasView(viewName) {
		t.Error("view should exist after CreateView with query builder")
	}

	_ = db.Schema().DropViewIfExists(viewName)
}

// TestPostgresSchemaViewCreateOrReplace verifies that CreateViewRaw can be called
// twice (PostgreSQL supports CREATE OR REPLACE VIEW).
// Note: PostgreSQL requires the new query to produce at least the same columns as
// the old one — columns can be added but not dropped.
func TestPostgresSchemaViewCreateOrReplace(t *testing.T) {
	db := SetupPostgresTest(t)
	viewName := "test_view_replace"

	_ = db.Schema().DropViewIfExists(viewName)

	if err := db.Schema().CreateViewRaw(viewName, "select * from users"); err != nil {
		t.Fatalf("first CreateViewRaw failed: %v", err)
	}

	// Second call should succeed without dropping (CREATE OR REPLACE)
	// Using the same SELECT to avoid the "cannot drop columns from view" restriction
	if err := db.Schema().CreateViewRaw(viewName, "select * from users"); err != nil {
		t.Fatalf("second CreateViewRaw (replace) failed: %v", err)
	}

	if !db.Schema().HasView(viewName) {
		t.Error("view should exist after CREATE OR REPLACE")
	}

	_ = db.Schema().DropViewIfExists(viewName)
}

// TestPostgresSchemaViewCreateRawEmpty verifies that empty select SQL is rejected.
func TestPostgresSchemaViewCreateRawEmpty(t *testing.T) {
	db := SetupPostgresTest(t)

	if err := db.Schema().CreateViewRaw("empty_view", ""); err == nil {
		t.Error("expected error for empty select SQL, got nil")
	}
}
