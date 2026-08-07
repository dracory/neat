package sqlite

import (
	"testing"
)

// TestSqliteSchemaViewCreateRawRoundTrip verifies the full CreateViewRaw → HasView →
// DropViewIfExists → !HasView round-trip on SQLite.
func TestSqliteSchemaViewCreateRawRoundTrip(t *testing.T) {
	db := SetupSQLiteTest(t)
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

// TestSqliteSchemaViewDropView verifies DropView on an existing view.
func TestSqliteSchemaViewDropView(t *testing.T) {
	db := SetupSQLiteTest(t)
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

// TestSqliteSchemaViewCreateWithQuery verifies CreateView with a query builder.
func TestSqliteSchemaViewCreateWithQuery(t *testing.T) {
	db := SetupSQLiteTest(t)
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

// TestSqliteSchemaViewCreateRawEmpty verifies that empty select SQL is rejected.
func TestSqliteSchemaViewCreateRawEmpty(t *testing.T) {
	db := SetupSQLiteTest(t)

	if err := db.Schema().CreateViewRaw("empty_view", ""); err == nil {
		t.Error("expected error for empty select SQL, got nil")
	}
}
