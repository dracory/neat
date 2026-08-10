//go:build integration

package oracle_test

import (
	"testing"
)

// TestOracleSchemaViewCreateRawRoundTrip verifies the full CreateViewRaw → HasView →
// DropViewIfExists → !HasView round-trip on Oracle.
//
// Note: Oracle stores unquoted identifiers in uppercase. The grammar's Wrap.Table
// uppercases identifier names, so views are created as uppercase. HasView compares
// directly against the name returned by GetViews (which is uppercase from ALL_VIEWS).
// Therefore uppercase view names are used throughout these tests.
func TestOracleSchemaViewCreateRawRoundTrip(t *testing.T) {
	db := SetupOracleTest(t)
	viewName := "TEST_VIEW_RAW"

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

// TestOracleSchemaViewDropView verifies DropView on an existing view.
func TestOracleSchemaViewDropView(t *testing.T) {
	db := SetupOracleTest(t)
	viewName := "TEST_VIEW_DROP"

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

// TestOracleSchemaViewCreateWithQuery verifies CreateView with a query builder.
func TestOracleSchemaViewCreateWithQuery(t *testing.T) {
	db := SetupOracleTest(t)
	viewName := "TEST_VIEW_QUERY"

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

// TestOracleSchemaViewCreateOrReplace verifies that CreateViewRaw can be called
// twice (Oracle supports CREATE OR REPLACE FORCE VIEW).
func TestOracleSchemaViewCreateOrReplace(t *testing.T) {
	db := SetupOracleTest(t)
	viewName := "TEST_VIEW_REPLACE"

	_ = db.Schema().DropViewIfExists(viewName)

	if err := db.Schema().CreateViewRaw(viewName, "select * from users"); err != nil {
		t.Fatalf("first CreateViewRaw failed: %v", err)
	}

	// Second call should succeed without dropping (CREATE OR REPLACE FORCE VIEW)
	if err := db.Schema().CreateViewRaw(viewName, "select id, name from users"); err != nil {
		t.Fatalf("second CreateViewRaw (replace) failed: %v", err)
	}

	if !db.Schema().HasView(viewName) {
		t.Error("view should exist after CREATE OR REPLACE")
	}

	_ = db.Schema().DropViewIfExists(viewName)
}

// TestOracleSchemaViewCreateRawEmpty verifies that empty select SQL is rejected.
func TestOracleSchemaViewCreateRawEmpty(t *testing.T) {
	db := SetupOracleTest(t)

	if err := db.Schema().CreateViewRaw("EMPTY_VIEW", ""); err == nil {
		t.Error("expected error for empty select SQL, got nil")
	}
}
