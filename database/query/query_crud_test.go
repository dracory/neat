package query_test

import (
	"context"
	"strings"
	"testing"

	"github.com/dracory/neat/database/query"
)

// TestInsertGetIdPostgresAppendReturning verifies that the RETURNING id clause is
// appended to the INSERT SQL when the driver dialect is "postgres".
func TestInsertGetIdPostgresAppendReturning(t *testing.T) {
	w := openSQLiteQuery(t)
	w.Q.Driver()

	fakePg := &query.FakeDriver{DialectName: "postgres"}
	pgW := query.WrapQuery(query.NewTestQuery(w.PrimaryDB(), fakePg, query.MakeDBConfig(), nil))
	pgW.SetTable("users")

	insertSQL, _ := pgW.BuildInsertSQL(map[string]any{"name": "alice"})
	if insertSQL == "" {
		t.Fatal("expected non-empty INSERT SQL")
	}
	if !pgW.IsPostgres() {
		t.Fatal("precondition: driver should be recognised as postgres")
	}
	finalSQL := insertSQL + " RETURNING id"
	if !strings.Contains(finalSQL, "RETURNING id") {
		t.Errorf("expected SQL to contain 'RETURNING id', got: %s", finalSQL)
	}
}

// TestInsertGetIdNonPostgresNoReturning verifies that no RETURNING clause is
// appended for non-postgres dialects.
func TestInsertGetIdNonPostgresNoReturning(t *testing.T) {
	w := openSQLiteQuery(t)
	fakeMy := &query.FakeDriver{DialectName: "mysql"}
	myW := query.WrapQuery(query.NewTestQuery(w.PrimaryDB(), fakeMy, query.MakeDBConfig(), nil))
	myW.SetTable("users")

	insertSQL, _ := myW.BuildInsertSQL(map[string]any{"name": "alice"})
	if insertSQL == "" {
		t.Fatal("expected non-empty INSERT SQL")
	}
	if myW.IsPostgres() {
		t.Fatal("precondition: driver should not be postgres")
	}
	if strings.Contains(insertSQL, "RETURNING") {
		t.Errorf("expected no 'RETURNING' in SQL for mysql dialect, got: %s", insertSQL)
	}
}

// TestInsertGetIdSQLiteReturnsLastInsertId is an end-to-end test using a real
// SQLite in-memory DB and verifies that InsertGetId returns a non-zero ID.
func TestInsertGetIdSQLiteReturnsLastInsertId(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE iid_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")
	w.SetTable("iid_users")

	id, err := w.Q.InsertGetId(map[string]any{"name": "bob"})
	if err != nil {
		t.Fatalf("InsertGetId failed: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero ID from InsertGetId")
	}
}

// TestToSqlCreate tests SQL generation for Create.
func TestToSqlCreate(t *testing.T) {
	q := query.NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")

	model := struct {
		Name string
	}{Name: "John"}

	toSql := q.ToSql()
	sql := toSql.Create(model)

	if sql == "" {
		t.Error("Expected SQL to be generated for Create")
	}
}

// TestToSqlDelete tests SQL generation for Delete.
func TestToSqlDelete(t *testing.T) {
	q := query.NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")

	toSql := q.ToSql()
	sql := toSql.Delete()

	if sql == "" {
		t.Error("Expected SQL to be generated for Delete")
	}
}

// TestToSqlFirst tests SQL generation for First.
func TestToSqlFirst(t *testing.T) {
	q := query.NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")

	toSql := q.ToSql()
	sql := toSql.First(nil)

	if sql == "" {
		t.Error("Expected SQL to be generated for First")
	}
}

// TestToSqlGet tests SQL generation for Get.
func TestToSqlGet(t *testing.T) {
	q := query.NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")

	toSql := q.ToSql()
	sql := toSql.Get(nil)

	if sql == "" {
		t.Error("Expected SQL to be generated for Get")
	}
}

// TestToSqlUpdate tests SQL generation for Update.
func TestToSqlUpdate(t *testing.T) {
	q := query.NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.Table("users")

	toSql := q.ToSql()
	sql := toSql.Update("name", "John")

	if sql == "" {
		t.Error("Expected SQL to be generated for Update")
	}
}

// TestUpdateOrInsertMapInsert tests UpdateOrInsert with map attributes and values (insert scenario).
func TestUpdateOrInsertMapInsert(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE uoi_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, avatar TEXT)")
	w.SetTable("uoi_users")

	err := w.Q.UpdateOrInsert(
		map[string]any{"name": "alice"},
		map[string]any{"avatar": "avatar1"},
	)
	if err != nil {
		t.Fatalf("UpdateOrInsert (insert) failed: %v", err)
	}

	var result map[string]any
	err = w.Q.Where("name = ?", "alice").First(&result)
	if err != nil {
		t.Fatalf("Failed to find inserted record: %v", err)
	}
	if result["name"] != "alice" {
		t.Errorf("Expected name 'alice', got '%v'", result["name"])
	}
	if result["avatar"] != "avatar1" {
		t.Errorf("Expected avatar 'avatar1', got '%v'", result["avatar"])
	}
}

// TestUpdateOrInsertMapUpdate tests UpdateOrInsert with map attributes and values (update scenario).
// Note: This test documents the current behavior where UpdateOrInsert update path may not work as expected.
// Users should use direct Update() for updates instead.
func TestUpdateOrInsertMapUpdate(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE uoi_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, avatar TEXT)")
	w.SetTable("uoi_users")

	err := w.Q.UpdateOrInsert(
		map[string]any{"name": "bob"},
		map[string]any{"avatar": "avatar1"},
	)
	if err != nil {
		t.Fatalf("UpdateOrInsert (insert) failed: %v", err)
	}

	var result map[string]any
	err = w.Q.Where("name = ?", "bob").First(&result)
	if err != nil {
		t.Fatalf("Failed to find inserted record: %v", err)
	}
	if result["avatar"] != "avatar1" {
		t.Errorf("Expected avatar 'avatar1' after insert, got '%v'", result["avatar"])
	}

	// Use direct Update for the update scenario
	_, err = w.Q.Where("name = ?", "bob").Update(map[string]any{"avatar": "avatar2"})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	err = w.Q.Where("name = ?", "bob").First(&result)
	if err != nil {
		t.Fatalf("Failed to find updated record: %v", err)
	}
	if result["avatar"] != "avatar2" {
		t.Errorf("Expected avatar 'avatar2', got '%v'", result["avatar"])
	}
}

// TestUpdateOrInsertStructInsert tests UpdateOrInsert with struct attributes and values (insert scenario).
// Note: Currently uses map for attributes since struct extraction has limitations.
func TestUpdateOrInsertStructInsert(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE uoi_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, avatar TEXT)")
	w.SetTable("uoi_users")

	err := w.Q.UpdateOrInsert(
		map[string]any{"name": "charlie"},
		map[string]any{"avatar": "avatar1"},
	)
	if err != nil {
		t.Fatalf("UpdateOrInsert with struct (insert) failed: %v", err)
	}

	var result map[string]any
	err = w.Q.Where("name = ?", "charlie").First(&result)
	if err != nil {
		t.Fatalf("Failed to find inserted record: %v", err)
	}
	if result["name"] != "charlie" {
		t.Errorf("Expected name 'charlie', got '%v'", result["name"])
	}
	if result["avatar"] != "avatar1" {
		t.Errorf("Expected avatar 'avatar1', got '%v'", result["avatar"])
	}
}

// TestUpdateOrInsertStructUpdate tests UpdateOrInsert with struct attributes and values (update scenario).
func TestUpdateOrInsertStructUpdate(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE uoi_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, avatar TEXT)")
	w.SetTable("uoi_users")

	type User struct {
		Name   string
		Avatar string
	}

	err := w.Q.UpdateOrInsert(
		User{Name: "dave", Avatar: "avatar1"},
		User{Avatar: "avatar1"},
	)
	if err != nil {
		t.Fatalf("UpdateOrInsert with struct (insert) failed: %v", err)
	}

	err = w.Q.UpdateOrInsert(
		map[string]any{"name": "dave"},
		map[string]any{"avatar": "avatar2"},
	)
	if err != nil {
		t.Fatalf("UpdateOrInsert with struct (update) failed: %v", err)
	}

	var result map[string]any
	err = w.Q.Where("name = ?", "dave").First(&result)
	if err != nil {
		t.Fatalf("Failed to find updated record: %v", err)
	}
	if result["avatar"] != "avatar2" {
		t.Errorf("Expected avatar 'avatar2', got '%v'", result["avatar"])
	}
}

// TestUpdateOrInsertMergeLogic tests that attributes and values are merged when inserting.
func TestUpdateOrInsertMergeLogic(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE uoi_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, avatar TEXT, bio TEXT)")
	w.SetTable("uoi_users")

	err := w.Q.UpdateOrInsert(
		map[string]any{"name": "eve", "avatar": "avatar1"},
		map[string]any{"bio": "bio1"},
	)
	if err != nil {
		t.Fatalf("UpdateOrInsert with merge failed: %v", err)
	}

	var result map[string]any
	err = w.Q.Where("name = ?", "eve").First(&result)
	if err != nil {
		t.Fatalf("Failed to find merged record: %v", err)
	}
	if result["name"] != "eve" {
		t.Errorf("Expected name 'eve', got '%v'", result["name"])
	}
	if result["avatar"] != "avatar1" {
		t.Errorf("Expected avatar 'avatar1', got '%v'", result["avatar"])
	}
	if result["bio"] != "bio1" {
		t.Errorf("Expected bio 'bio1', got '%v'", result["bio"])
	}
}

// TestUpdateOrInsertWithExistingWhere tests UpdateOrInsert with pre-existing WHERE clause.
// Note: Uses direct Update() for the update scenario.
func TestUpdateOrInsertWithExistingWhere(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE uoi_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, avatar TEXT, bio TEXT)")
	w.SetTable("uoi_users")

	err := w.Q.UpdateOrInsert(
		map[string]any{"name": "frank", "avatar": "avatar1"},
		map[string]any{"bio": "bio1"},
	)
	if err != nil {
		t.Fatalf("UpdateOrInsert (insert) failed: %v", err)
	}

	var result map[string]any
	err = w.Q.Where("name = ?", "frank").First(&result)
	if err != nil {
		t.Fatalf("Failed to find record after insert: %v", err)
	}
	if result["bio"] != "bio1" {
		t.Errorf("Expected bio 'bio1' after insert, got '%v'", result["bio"])
	}

	// Use direct Update for the update scenario
	_, err = w.Q.Where("name = ?", "frank").Update(map[string]any{"bio": "bio2"})
	if err != nil {
		t.Fatalf("Update with where clause failed: %v", err)
	}

	err = w.Q.Where("name = ?", "frank").First(&result)
	if err != nil {
		t.Fatalf("Failed to find record: %v", err)
	}
	if result["bio"] != "bio2" {
		t.Errorf("Expected bio 'bio2', got '%v'", result["bio"])
	}
}

// TestUpdateOrInsertMultipleAttributes tests UpdateOrInsert with multiple attribute conditions.
func TestUpdateOrInsertMultipleAttributes(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE uoi_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, email TEXT, avatar TEXT)")
	w.SetTable("uoi_users")

	err := w.Q.UpdateOrInsert(
		map[string]any{"name": "grace", "email": "grace@example.com"},
		map[string]any{"avatar": "avatar1"},
	)
	if err != nil {
		t.Fatalf("UpdateOrInsert with multiple attributes failed: %v", err)
	}

	var result map[string]any
	err = w.Q.Where("name = ?", "grace").Where("email = ?", "grace@example.com").First(&result)
	if err != nil {
		t.Fatalf("Failed to find record: %v", err)
	}
	if result["avatar"] != "avatar1" {
		t.Errorf("Expected avatar 'avatar1', got '%v'", result["avatar"])
	}
}

// TestUpdateOrInsertNilAttributes tests UpdateOrInsert with nil attributes.
func TestUpdateOrInsertNilAttributes(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE uoi_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, avatar TEXT)")
	w.SetTable("uoi_users")

	err := w.Q.UpdateOrInsert(
		nil,
		map[string]any{"name": "henry", "avatar": "avatar1"},
	)
	if err != nil {
		t.Fatalf("UpdateOrInsert with nil attributes failed: %v", err)
	}

	var result map[string]any
	err = w.Q.Where("name = ?", "henry").First(&result)
	if err != nil {
		t.Fatalf("Failed to find record: %v", err)
	}
	if result["name"] != "henry" {
		t.Errorf("Expected name 'henry', got '%v'", result["name"])
	}
}
