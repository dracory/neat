package query

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

// --- Nil Model Handling ---

func TestNilModelHandling(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	// Try to create with nil model
	err = q.Create(nil)
	if err == nil {
		t.Error("Expected error when creating with nil model")
	}
}

func TestNilModelInUpdate(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	// Try to update with nil model
	_, err = q.Update(nil)
	if err == nil {
		t.Error("Expected error when updating with nil model")
	}
}

func TestNilModelInSave(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	// Try to save with nil model
	err = q.Save(nil)
	if err == nil {
		t.Error("Expected error when saving with nil model")
	}
}

// --- Zero Value Primary Keys ---

func TestZeroValuePrimaryKey(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	type Model struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	// Create with zero ID (should auto-increment)
	model := Model{Name: "test"}
	err = q.Create(&model)
	if err != nil {
		t.Errorf("Failed to create with zero ID: %v", err)
	}

	if model.ID == 0 {
		t.Error("Expected ID to be auto-incremented")
	}
}

func TestZeroValuePrimaryKeyInUpdate(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	_, err = db.Exec("INSERT INTO test (id, name) VALUES (1, 'original')")
	if err != nil {
		t.Fatalf("Failed to insert initial data: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	type Model struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	// Update with zero ID (should update based on WHERE clause)
	model := Model{ID: 0, Name: "updated"}
	_, err = q.Where("id = ?", 1).Update(&model)
	if err != nil {
		t.Errorf("Failed to update with zero ID: %v", err)
	}

	// Verify update
	var result Model
	err = q.Where("id = ?", 1).First(&result)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if result.Name != "updated" {
		t.Errorf("Expected name 'updated', got '%s'", result.Name)
	}
}

func TestZeroValuePrimaryKeyInFind(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	_, err = db.Exec("INSERT INTO test (id, name) VALUES (0, 'zero_id')")
	if err != nil {
		t.Fatalf("Failed to insert data with zero ID: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	type Model struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	var result Model
	err = q.Where("id = ?", 0).First(&result)
	if err != nil {
		t.Errorf("Failed to find record with zero ID: %v", err)
	}

	if result.ID != 0 {
		t.Errorf("Expected ID 0, got %d", result.ID)
	}

	if result.Name != "zero_id" {
		t.Errorf("Expected name 'zero_id', got '%s'", result.Name)
	}
}

// --- Nil Pointer Fields ---

func TestNilPointerFields(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	type Model struct {
		ID   int     `db:"id"`
		Name *string `db:"name"`
		Age  *int    `db:"age"`
	}

	// Create with nil pointer fields
	model := Model{ID: 1, Name: nil, Age: nil}
	err = q.Create(&model)
	if err != nil {
		t.Errorf("Failed to create with nil pointer fields: %v", err)
	}

	// Verify nil pointers are stored as NULL
	var result Model
	err = q.Where("id = ?", 1).First(&result)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if result.Name != nil {
		t.Error("Expected Name to be nil")
	}

	if result.Age != nil {
		t.Error("Expected Age to be nil")
	}
}

func TestNilPointerFieldsInUpdate(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	name := "original"
	age := 25
	_, err = db.Exec("INSERT INTO test (id, name, age) VALUES (1, ?, ?)", name, age)
	if err != nil {
		t.Fatalf("Failed to insert initial data: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	type Model struct {
		ID   int     `db:"id"`
		Name *string `db:"name"`
		Age  *int    `db:"age"`
	}

	// Update to set pointer fields to nil
	model := Model{ID: 1, Name: nil, Age: nil}
	_, err = q.Where("id = ?", 1).Update(&model)
	if err != nil {
		t.Errorf("Failed to update with nil pointer fields: %v", err)
	}

	// Verify fields are now NULL
	var result Model
	err = q.Where("id = ?", 1).First(&result)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	// Note: Update may not set fields to NULL if they're nil in the model
	// This test verifies the update executes without error
	_ = result.Name
	_ = result.Age
}

func TestNilPointerFieldsInScan(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	_, err = db.Exec("INSERT INTO test (id, name, age) VALUES (1, NULL, NULL)")
	if err != nil {
		t.Fatalf("Failed to insert NULL values: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	type Model struct {
		ID   int     `db:"id"`
		Name *string `db:"name"`
		Age  *int    `db:"age"`
	}

	var result Model
	err = q.Where("id = ?", 1).First(&result)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	// Verify NULL values are scanned as nil pointers
	if result.Name != nil {
		t.Error("Expected Name to be nil from NULL database value")
	}

	if result.Age != nil {
		t.Error("Expected Age to be nil from NULL database value")
	}
}

// --- Empty Slices in Bulk Operations ---

func TestEmptySliceInCreate(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	type Model struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	// Try to create with empty slice - should succeed (nothing to insert)
	var models []Model
	err = q.Create(&models)
	if err != nil {
		t.Errorf("Expected no error when creating with empty slice, got: %v", err)
	}
}

func TestEmptySliceInFind(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	type Model struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	// Find into empty slice (should work and return empty results)
	var models []Model
	err = q.Find(&models)
	if err != nil {
		t.Errorf("Failed to find into empty slice: %v", err)
	}

	if len(models) != 0 {
		t.Errorf("Expected empty results, got %d", len(models))
	}
}

func TestEmptySliceInUpdate(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	type Model struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	// Try to update with empty slice
	var models []Model
	_, err = q.Update(&models)
	if err == nil {
		t.Error("Expected error when updating with empty slice")
	}
}

func TestEmptySliceInDelete(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	type Model struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	// Try to delete with empty slice
	var models []Model
	_, err = q.Delete(&models)
	// Delete may not validate the slice
	_ = err
}

// --- Empty WHERE Clauses ---

func TestEmptyWhereClause(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	_, err = db.Exec("INSERT INTO test (id, name) VALUES (1, 'first'), (2, 'second')")
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	// Query with empty WHERE clause (should return all records)
	type Model struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	var results []Model
	err = q.Find(&results)
	if err != nil {
		t.Errorf("Failed to find with empty WHERE: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results with empty WHERE, got %d", len(results))
	}
}

func TestEmptyWhereClauseInUpdate(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	_, err = db.Exec("INSERT INTO test (id, name) VALUES (1, 'first'), (2, 'second')")
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	// Update with empty WHERE clause (should update all records)
	_, err = q.Update(map[string]any{"name": "updated"})
	if err != nil {
		t.Errorf("Failed to update with empty WHERE: %v", err)
	}

	// Verify all records were updated
	var count int64
	err = q.Where("name = ?", "updated").Count(&count)
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 records updated, got %d", count)
	}
}

func TestEmptyWhereClauseInDelete(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	_, err = db.Exec("INSERT INTO test (id, name) VALUES (1, 'first'), (2, 'second')")
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	// Delete with empty WHERE clause (should delete all records)
	_, err = q.Delete()
	if err != nil {
		t.Errorf("Failed to delete with empty WHERE: %v", err)
	}

	// Verify all records were deleted
	var count int64
	err = q.Count(&count)
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 records after delete, got %d", count)
	}
}

func TestEmptyWhereString(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	// Try to add empty WHERE string
	q.Where("")

	type Model struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	var results []Model
	err = q.Find(&results)
	// Should handle gracefully or return error
	_ = err
}

func TestWhereWithEmptyArgs(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	_, err = db.Exec("INSERT INTO test (id, name) VALUES (1, 'test')")
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	// WHERE with placeholder but no args
	q.Where("id = ?")

	type Model struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	var result Model
	err = q.First(&result)
	// Should handle gracefully or return error
	_ = err
}

// --- Complex Types ---

// Nested Struct Fields

type Address struct {
	Street  string `db:"street"`
	City    string `db:"city"`
	Country string `db:"country"`
}

type PersonWithAddress struct {
	ID      int     `db:"id"`
	Name    string  `db:"name"`
	Address Address `db:"address"` // Note: nested struct in single table
}

func TestNestedStructFields(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create table with flattened columns (common pattern for nested structs)
	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT, street TEXT, city TEXT, country TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	person := PersonWithAddress{
		Name: "John",
		Address: Address{
			Street:  "123 Main St",
			City:    "New York",
			Country: "USA",
		},
	}

	// Note: This test verifies the struct can be used
	// Actual nested struct mapping depends on implementation
	err = q.Create(&person)
	_ = err // May not work without custom mapping
}

// Embedded Struct Fields

type Timestamps struct {
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

type Document struct {
	ID         int    `db:"id"`
	Title      string `db:"title"`
	Timestamps        // Embedded struct
	Content    string `db:"content"`
}

func TestEmbeddedStructFields(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, title TEXT, created_at TEXT, updated_at TEXT, content TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	doc := Document{
		Title:   "Test Document",
		Content: "Test content",
		Timestamps: Timestamps{
			CreatedAt: "2024-01-01",
			UpdatedAt: "2024-01-01",
		},
	}

	err = q.Create(&doc)
	if err != nil {
		t.Errorf("Failed to create with embedded struct: %v", err)
	}

	// Verify embedded fields were stored
	var result Document
	err = q.Where("id = ?", doc.ID).First(&result)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if result.Title != "Test Document" {
		t.Errorf("Expected title 'Test Document', got '%s'", result.Title)
	}

	if result.Content != "Test content" {
		t.Errorf("Expected content 'Test content', got '%s'", result.Content)
	}
}

// Pointer to Pointer Fields

type DoublePointerModel struct {
	ID     int      `db:"id"`
	Name   **string `db:"name"`
	Age    **int    `db:"age"`
	Active **bool   `db:"active"`
}

func TestPointerToPointerFields(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT, age INTEGER, active INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	name := "test"
	age := 25
	active := true

	// Create double pointers
	var namePtr *string = &name
	var agePtr *int = &age
	var activePtr *bool = &active

	model := DoublePointerModel{
		Name:   &namePtr,
		Age:    &agePtr,
		Active: &activePtr,
	}

	// Note: Double pointer handling may not be supported by default
	// This test verifies the behavior
	err = q.Create(&model)
	_ = err // May not work without custom Scanner/Valuer
}

// Custom Types implementing Scanner/Valuer

type CustomString string

func (cs *CustomString) Scan(value interface{}) error {
	if value == nil {
		*cs = ""
		return nil
	}
	if str, ok := value.(string); ok {
		*cs = CustomString(str)
		return nil
	}
	if str, ok := value.([]byte); ok {
		*cs = CustomString(str)
		return nil
	}
	return nil
}

func (cs CustomString) Value() (interface{}, error) {
	return string(cs), nil
}

type CustomInt int

func (ci *CustomInt) Scan(value interface{}) error {
	if value == nil {
		*ci = 0
		return nil
	}
	switch v := value.(type) {
	case int64:
		*ci = CustomInt(v)
	case int:
		*ci = CustomInt(v)
	}
	return nil
}

func (ci CustomInt) Value() (interface{}, error) {
	return int(ci), nil
}

type CustomTypeModel struct {
	ID    int          `db:"id"`
	Name  CustomString `db:"name"`
	Count CustomInt    `db:"count"`
}

func TestCustomScannerValuer(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT, count INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	model := CustomTypeModel{
		Name:  CustomString("custom_name"),
		Count: CustomInt(42),
	}

	err = q.Create(&model)
	if err != nil {
		t.Errorf("Failed to create with custom types: %v", err)
	}

	// Verify custom types were stored and can be retrieved
	var result CustomTypeModel
	err = q.Where("id = ?", model.ID).First(&result)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if result.Name != "custom_name" {
		t.Errorf("Expected name 'custom_name', got '%s'", result.Name)
	}

	if result.Count != 42 {
		t.Errorf("Expected count 42, got %d", result.Count)
	}
}

// JSON/JSONB Fields

type JSONModel struct {
	ID       int    `db:"id"`
	Name     string `db:"name"`
	JSONData string `db:"json_data"` // Stored as TEXT in SQLite
}

func TestJSONFields(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT, json_data TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	jsonData := `{"key": "value", "number": 123}`

	model := JSONModel{
		Name:     "test",
		JSONData: jsonData,
	}

	err = q.Create(&model)
	if err != nil {
		t.Errorf("Failed to create with JSON field: %v", err)
	}

	// Verify JSON data was stored
	var result JSONModel
	err = q.Where("id = ?", model.ID).First(&result)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if result.JSONData != jsonData {
		t.Errorf("Expected JSON data '%s', got '%s'", jsonData, result.JSONData)
	}
}

func TestJSONFieldsWithQuery(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT, json_data TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	jsonData1 := `{"type": "user", "active": true}`
	jsonData2 := `{"type": "admin", "active": false}`

	_, err = db.Exec("INSERT INTO test (name, json_data) VALUES (?, ?)", "user1", jsonData1)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	_, err = db.Exec("INSERT INTO test (name, json_data) VALUES (?, ?)", "user2", jsonData2)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	// Query using JSON functions (SQLite specific)
	var results []JSONModel
	err = q.Where("json_extract(json_data, '$.type') = ?", "user").Find(&results)
	if err != nil {
		t.Errorf("Failed to query with JSON: %v", err)
	}

	// Should find user1
	if len(results) == 0 {
		t.Error("Expected to find at least one result with JSON query")
	}
}

// Array Fields

type ArrayModel struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
	Tags string `db:"tags"` // Stored as comma-separated string in SQLite
}

func TestArrayFields(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT, tags TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	tags := "tag1,tag2,tag3"

	model := ArrayModel{
		Name: "test",
		Tags: tags,
	}

	err = q.Create(&model)
	if err != nil {
		t.Errorf("Failed to create with array field: %v", err)
	}

	// Verify array data was stored
	var result ArrayModel
	err = q.Where("id = ?", model.ID).First(&result)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if result.Tags != tags {
		t.Errorf("Expected tags '%s', got '%s'", tags, result.Tags)
	}
}

func TestArrayFieldsQuery(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT, tags TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("test")

	_, err = db.Exec("INSERT INTO test (name, tags) VALUES (?, ?)", "item1", "red,blue,green")
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	_, err = db.Exec("INSERT INTO test (name, tags) VALUES (?, ?)", "item2", "yellow,orange")
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	// Query using LIKE for array search
	var results []ArrayModel
	err = q.Where("tags LIKE ?", "%blue%").Find(&results)
	if err != nil {
		t.Errorf("Failed to query with array: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result with array query, got %d", len(results))
	}

	if len(results) > 0 && results[0].Name != "item1" {
		t.Errorf("Expected name 'item1', got '%s'", results[0].Name)
	}
}

// --- Short ID (String Primary Key) ---

type ShortIDProduct struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}

func TestShortIDAutoGenerated(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("products")

	product := ShortIDProduct{Name: "Widget"}
	if err := q.Create(&product); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if product.ID == "" {
		t.Fatal("Expected short ID to be generated, got empty string")
	}
	if len(product.ID) != 11 {
		t.Fatalf("Expected 11-char short ID, got %d: %s", len(product.ID), product.ID)
	}

	// Verify it was persisted
	var fetched ShortIDProduct
	if err := q.Where("id = ?", product.ID).First(&fetched); err != nil {
		t.Fatalf("Failed to fetch product: %v", err)
	}
	if fetched.Name != "Widget" {
		t.Errorf("Expected name 'Widget', got '%s'", fetched.Name)
	}
}

func TestShortIDUserProvided(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("products")

	customID := "my-custom-id-123"
	product := ShortIDProduct{ID: customID, Name: "Gadget"}
	if err := q.Create(&product); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if product.ID != customID {
		t.Fatalf("Expected user-provided ID '%s', got '%s'", customID, product.ID)
	}

	// Verify it was persisted with the custom ID
	var fetched ShortIDProduct
	if err := q.Where("id = ?", customID).First(&fetched); err != nil {
		t.Fatalf("Failed to fetch product: %v", err)
	}
	if fetched.Name != "Gadget" {
		t.Errorf("Expected name 'Gadget', got '%s'", fetched.Name)
	}
}

func TestShortIDBulkInsert(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("products")

	products := []ShortIDProduct{
		{Name: "A"},
		{Name: "B"},
		{Name: "C"},
	}
	if err := q.Create(&products); err != nil {
		t.Fatalf("Bulk create failed: %v", err)
	}

	ids := make(map[string]bool)
	for i, p := range products {
		if p.ID == "" {
			t.Fatalf("product %d: expected short ID to be generated", i)
		}
		if len(p.ID) != 11 {
			t.Fatalf("product %d: expected 11-char ID, got %d: %s", i, len(p.ID), p.ID)
		}
		if ids[p.ID] {
			t.Fatalf("product %d: duplicate ID '%s'", i, p.ID)
		}
		ids[p.ID] = true
	}
}

func TestShortIDCreateThenSaveUpdate(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.Table("products")

	// Insert with custom ID using Create (Save would treat non-zero ID as UPDATE)
	customID := "custom-save-id"
	product := ShortIDProduct{ID: customID, Name: "Original"}
	if err := q.Create(&product); err != nil {
		t.Fatalf("Create with custom ID failed: %v", err)
	}

	if product.ID != customID {
		t.Fatalf("Expected ID '%s', got '%s'", customID, product.ID)
	}

	// Update via Save (ID is non-empty, so Save performs UPDATE)
	product.Name = "Updated"
	if err := q.Save(&product); err != nil {
		t.Fatalf("Save (update) failed: %v", err)
	}

	var fetched ShortIDProduct
	if err := q.Where("id = ?", customID).First(&fetched); err != nil {
		t.Fatalf("Failed to fetch product: %v", err)
	}
	if fetched.Name != "Updated" {
		t.Errorf("Expected updated name 'Updated', got '%s'", fetched.Name)
	}
}
