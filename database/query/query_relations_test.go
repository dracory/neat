package query

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	_ "modernc.org/sqlite"
)

func TestInitializeRelations(t *testing.T) {
	q := NewQuery(nil, nil, nil, "", nil, nil)
	q.withRelations = []string{"User", "Posts"}

	type TestModel struct {
		User  *string
		Posts []int
	}

	model := &TestModel{}
	v := reflect.ValueOf(model)

	q.initializeRelations(v)

	// Check that relations are initialized (should be non-nil after initialization)
	userField := v.Elem().FieldByName("User")
	if !userField.IsValid() {
		t.Error("Expected User field to be valid")
	}
	if userField.Kind() != reflect.Ptr {
		t.Error("Expected User field to be a pointer")
	}
	if userField.IsNil() {
		t.Error("Expected User field to be non-nil after initialization")
	}

	postsField := v.Elem().FieldByName("Posts")
	if !postsField.IsValid() {
		t.Error("Expected Posts field to be valid")
	}
	if postsField.Kind() != reflect.Slice {
		t.Error("Expected Posts field to be a slice")
	}
	if postsField.IsNil() {
		t.Error("Expected Posts field to be non-nil after initialization")
	}
}

func TestInitializeRelationsWithNilValue(t *testing.T) {
	q := NewQuery(nil, nil, nil, "", nil, nil)
	q.withRelations = []string{"User"}

	var model *struct{}
	v := reflect.ValueOf(model)

	// Should not panic
	q.initializeRelations(v)
}

func TestInitializeRelationsWithNonStruct(t *testing.T) {
	q := NewQuery(nil, nil, nil, "", nil, nil)
	q.withRelations = []string{"User"}

	model := "not a struct"
	v := reflect.ValueOf(model)

	// Should not panic
	q.initializeRelations(v)
}

func TestLoadRelations(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create tables
	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}
	_, err = db.Exec("CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT, user_id INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create posts table: %v", err)
	}

	// Insert test data
	_, err = db.Exec("INSERT INTO users (id, name) VALUES (1, 'John Doe')")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}
	_, err = db.Exec("INSERT INTO posts (id, title, user_id) VALUES (1, 'Test Post', 1)")
	if err != nil {
		t.Fatalf("Failed to insert post: %v", err)
	}

	type User struct {
		ID   int
		Name string
	}

	type Post struct {
		ID     int
		Title  string
		UserID int
		User   *User
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.table = "posts"
	q.withRelations = []string{"User"}

	post := &Post{ID: 1, Title: "Test Post", UserID: 1}
	v := reflect.ValueOf(post)

	err = q.loadRelations(v)
	if err != nil {
		t.Errorf("loadRelations failed: %v", err)
	}

	// Check that User was loaded
	userField := v.Elem().FieldByName("User")
	if !userField.IsValid() {
		t.Error("Expected User field to be valid")
	}
	if userField.IsNil() {
		t.Error("Expected User field to be non-nil after loading")
	}

	// Check that User data is correct
	user := userField.Interface().(*User)
	if user.ID != 1 {
		t.Errorf("Expected User ID 1, got %d", user.ID)
	}
	if user.Name != "John Doe" {
		t.Errorf("Expected User name 'John Doe', got '%s'", user.Name)
	}
}

func TestLoadRelationsWithConstraintCallback(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create tables
	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, status TEXT)")
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}
	_, err = db.Exec("CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT, user_id INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create posts table: %v", err)
	}

	// Insert test data
	_, err = db.Exec("INSERT INTO users (id, name, status) VALUES (1, 'John Doe', 'active')")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}
	_, err = db.Exec("INSERT INTO users (id, name, status) VALUES (2, 'Jane Doe', 'inactive')")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}
	_, err = db.Exec("INSERT INTO posts (id, title, user_id) VALUES (1, 'Test Post', 1)")
	if err != nil {
		t.Fatalf("Failed to insert post: %v", err)
	}

	type User struct {
		ID     int
		Name   string
		Status string
	}

	type Post struct {
		ID     int
		Title  string
		UserID int
		User   *User
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.table = "posts"
	q.withRelations = []string{"User"}
	q.relationConstraints = map[string]func(contractsorm.Query) contractsorm.Query{
		"User": func(q contractsorm.Query) contractsorm.Query {
			return q.Where("status = ?", "active")
		},
	}

	post := &Post{ID: 1, Title: "Test Post", UserID: 1}
	v := reflect.ValueOf(post)

	err = q.loadRelations(v)
	if err != nil {
		t.Errorf("loadRelations with constraint failed: %v", err)
	}

	// Check that User was loaded with constraint applied
	userField := v.Elem().FieldByName("User")
	if userField.IsNil() {
		t.Error("Expected User field to be non-nil after loading with constraint")
	}

	user := userField.Interface().(*User)
	if user.Status != "active" {
		t.Errorf("Expected User status 'active', got '%s'", user.Status)
	}
}

func TestLoadRelationsWithForeignKeyResolution(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create tables with snake_case foreign key
	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}
	_, err = db.Exec("CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT, user_id INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create posts table: %v", err)
	}

	// Insert test data
	_, err = db.Exec("INSERT INTO users (id, name) VALUES (1, 'John Doe')")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}
	_, err = db.Exec("INSERT INTO posts (id, title, user_id) VALUES (1, 'Test Post', 1)")
	if err != nil {
		t.Fatalf("Failed to insert post: %v", err)
	}

	type User struct {
		ID   int
		Name string
	}

	type Post struct {
		ID     int
		Title  string
		UserID int
		User   *User
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.table = "posts"
	q.withRelations = []string{"User"}

	post := &Post{ID: 1, Title: "Test Post", UserID: 1}
	v := reflect.ValueOf(post)

	err = q.loadRelations(v)
	if err != nil {
		t.Errorf("loadRelations with foreign key resolution failed: %v", err)
	}

	// Check that User was loaded using snake_case foreign key
	userField := v.Elem().FieldByName("User")
	if userField.IsNil() {
		t.Error("Expected User field to be non-nil after loading")
	}

	user := userField.Interface().(*User)
	if user.ID != 1 {
		t.Errorf("Expected User ID 1, got %d", user.ID)
	}
}

func TestLoadRelationsRecursivePrevention(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create tables
	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}
	_, err = db.Exec("CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT, user_id INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create posts table: %v", err)
	}

	// Insert test data
	_, err = db.Exec("INSERT INTO users (id, name) VALUES (1, 'John Doe')")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}
	_, err = db.Exec("INSERT INTO posts (id, title, user_id) VALUES (1, 'Test Post', 1)")
	if err != nil {
		t.Fatalf("Failed to insert post: %v", err)
	}

	type User struct {
		ID    int
		Name  string
		Posts interface{}
	}

	type Post struct {
		ID     int
		Title  string
		UserID int
		User   *User
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.table = "posts"
	q.withRelations = []string{"User"}

	post := &Post{ID: 1, Title: "Test Post", UserID: 1}
	v := reflect.ValueOf(post)

	// This should not cause infinite recursion
	err = q.loadRelations(v)
	if err != nil {
		t.Errorf("loadRelations failed: %v", err)
	}

	// Check that User was loaded
	userField := v.Elem().FieldByName("User")
	if userField.IsNil() {
		t.Error("Expected User field to be non-nil after loading")
	}

	user := userField.Interface().(*User)
	if user.ID != 1 {
		t.Errorf("Expected User ID 1, got %d", user.ID)
	}
}

func TestLoadRelationsWithDifferentModelTypes(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create tables
	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}
	_, err = db.Exec("CREATE TABLE comments (id INTEGER PRIMARY KEY, text TEXT, user_id INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create comments table: %v", err)
	}

	// Insert test data
	_, err = db.Exec("INSERT INTO users (id, name) VALUES (1, 'John Doe')")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}
	_, err = db.Exec("INSERT INTO comments (id, text, user_id) VALUES (1, 'Test Comment', 1)")
	if err != nil {
		t.Fatalf("Failed to insert comment: %v", err)
	}

	type User struct {
		ID   int
		Name string
	}

	type Comment struct {
		ID     int
		Text   string
		UserID int
		User   *User
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.table = "comments"
	q.withRelations = []string{"User"}

	comment := &Comment{ID: 1, Text: "Test Comment", UserID: 1}
	v := reflect.ValueOf(comment)

	err = q.loadRelations(v)
	if err != nil {
		t.Errorf("loadRelations with different model type failed: %v", err)
	}

	// Check that User was loaded
	userField := v.Elem().FieldByName("User")
	if userField.IsNil() {
		t.Error("Expected User field to be non-nil after loading")
	}

	user := userField.Interface().(*User)
	if user.ID != 1 {
		t.Errorf("Expected User ID 1, got %d", user.ID)
	}
	if user.Name != "John Doe" {
		t.Errorf("Expected User name 'John Doe', got '%s'", user.Name)
	}
}

func TestLoadRelationsWithZeroForeignKey(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create tables
	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}
	_, err = db.Exec("CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT, user_id INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create posts table: %v", err)
	}

	type User struct {
		ID   int
		Name string
	}

	type Post struct {
		ID     int
		Title  string
		UserID int
		User   *User
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.table = "posts"
	q.withRelations = []string{"User"}

	post := &Post{ID: 1, Title: "Test Post", UserID: 0}
	v := reflect.ValueOf(post)

	err = q.loadRelations(v)
	if err != nil {
		t.Errorf("loadRelations with zero foreign key failed: %v", err)
	}

	// User should remain nil when foreign key is zero
	userField := v.Elem().FieldByName("User")
	if !userField.IsNil() {
		t.Error("Expected User field to be nil when foreign key is zero")
	}
}

func TestLoadRelationsWithMissingForeignKeyField(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create tables
	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}
	_, err = db.Exec("CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT)")
	if err != nil {
		t.Fatalf("Failed to create posts table: %v", err)
	}

	type User struct {
		ID   int
		Name string
	}

	type Post struct {
		ID    int
		Title string
		User  *User
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.table = "posts"
	q.withRelations = []string{"User"}

	post := &Post{ID: 1, Title: "Test Post"}
	v := reflect.ValueOf(post)

	// Should not panic when foreign key field is missing
	err = q.loadRelations(v)
	if err != nil {
		t.Errorf("loadRelations with missing foreign key field failed: %v", err)
	}
}
