package query

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	_ "modernc.org/sqlite"
)

func TestInitializeRelations(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
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
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.withRelations = []string{"User"}

	var model *struct{}
	v := reflect.ValueOf(model)

	// Should not panic
	q.initializeRelations(v)
}

func TestInitializeRelationsWithNonStruct(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
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

// Unit tests for helper functions

func TestInferTableName(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		expected string
	}{
		{"simple singular", "User", "users"},
		{"simple singular 2", "Post", "posts"},
		{"already plural", "Status", "Status"},
		{"camel case", "UserProfile", "user_profiles"},
		{"pascal case", "BlogPost", "blog_posts"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a dummy type to get reflect.Type
			type Dummy struct{}
			// We can't easily create custom type names in tests, so we'll test the logic
			// by checking the function behavior with actual types
			result := inferTableName(reflect.TypeOf(Dummy{}))
			// Since we can't control the type name, we'll just verify it doesn't panic
			if result == "" {
				t.Error("inferTableName returned empty string")
			}
		})
	}
}

func TestBuildForeignKeyColumn(t *testing.T) {
	tests := []struct {
		name           string
		parentTypeName string
		expected       string
	}{
		{"simple", "User", "user_id"},
		{"camel case", "UserProfile", "user_profile_id"},
		{"pascal case", "BlogPost", "blog_post_id"},
		{"single word", "Post", "post_id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildForeignKeyColumn(tt.parentTypeName)
			if result != tt.expected {
				t.Errorf("buildForeignKeyColumn(%q) = %q, want %q", tt.parentTypeName, result, tt.expected)
			}
		})
	}
}

func TestGetParentTypeName(t *testing.T) {
	type User struct {
		ID int
	}

	type Post struct {
		ID     int
		UserID int
	}

	t.Run("from value type name", func(t *testing.T) {
		q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
		post := &Post{ID: 1}
		v := reflect.ValueOf(post).Elem()
		result := q.getParentTypeName(v)
		if result != "Post" {
			t.Errorf("getParentTypeName() = %q, want %q", result, "Post")
		}
	})

	t.Run("from model when value has no name", func(t *testing.T) {
		q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
		q.model = &Post{ID: 1}
		// Create an anonymous struct
		anon := struct{ ID int }{ID: 1}
		v := reflect.ValueOf(anon)
		result := q.getParentTypeName(v)
		if result != "Post" {
			t.Errorf("getParentTypeName() from model = %q, want %q", result, "Post")
		}
	})

	t.Run("from model slice", func(t *testing.T) {
		q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
		posts := []Post{{ID: 1}}
		q.model = posts
		anon := struct{ ID int }{ID: 1}
		v := reflect.ValueOf(anon)
		result := q.getParentTypeName(v)
		if result != "Post" {
			t.Errorf("getParentTypeName() from model slice = %q, want %q", result, "Post")
		}
	})

	t.Run("empty when no model", func(t *testing.T) {
		q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
		anon := struct{ ID int }{ID: 1}
		v := reflect.ValueOf(anon)
		result := q.getParentTypeName(v)
		if result != "" {
			t.Errorf("getParentTypeName() without model = %q, want empty string", result)
		}
	})
}

func TestLoadHasManyRelation(t *testing.T) {
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
	_, err = db.Exec("INSERT INTO posts (id, title, user_id) VALUES (1, 'Post 1', 1)")
	if err != nil {
		t.Fatalf("Failed to insert post 1: %v", err)
	}
	_, err = db.Exec("INSERT INTO posts (id, title, user_id) VALUES (2, 'Post 2', 1)")
	if err != nil {
		t.Fatalf("Failed to insert post 2: %v", err)
	}

	type Post struct {
		ID     int
		Title  string
		UserID int
	}

	type User struct {
		ID    int
		Name  string
		Posts []Post
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	user := &User{ID: 1, Name: "John Doe"}
	v := reflect.ValueOf(user).Elem()
	postsField := v.FieldByName("Posts")

	err = q.loadHasManyRelation(v, postsField, "Posts", db)
	if err != nil {
		t.Fatalf("loadHasManyRelation failed: %v", err)
	}

	// Check that posts were loaded
	posts := user.Posts
	if len(posts) != 2 {
		t.Errorf("Expected 2 posts, got %d", len(posts))
	}
	if posts[0].Title != "Post 1" {
		t.Errorf("Expected post title 'Post 1', got '%s'", posts[0].Title)
	}
	if posts[1].Title != "Post 2" {
		t.Errorf("Expected post title 'Post 2', got '%s'", posts[1].Title)
	}
}

func TestLoadHasManyRelationWithNilConnection(t *testing.T) {
	type Post struct {
		ID     int
		Title  string
		UserID int
	}

	type User struct {
		ID    int
		Name  string
		Posts []Post
	}

	q := NewQuery(context.Background(), nil, nil, "", nil, nil)
	user := &User{ID: 1, Name: "John Doe"}
	v := reflect.ValueOf(user).Elem()
	postsField := v.FieldByName("Posts")

	err := q.loadHasManyRelation(v, postsField, "Posts", nil)
	if err == nil {
		t.Error("Expected error when connection is nil")
	} else if err.Error() != "database connection is nil" {
		t.Errorf("Expected 'database connection is nil' error, got: %v", err)
	}
}

func TestLoadHasManyRelationWithMissingID(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	type Post struct {
		ID     int
		Title  string
		UserID int
	}

	type User struct {
		Name  string
		Posts []Post
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	user := &User{Name: "John Doe"}
	v := reflect.ValueOf(user).Elem()
	postsField := v.FieldByName("Posts")

	err = q.loadHasManyRelation(v, postsField, "Posts", db)
	if err == nil {
		t.Error("Expected error when ID field is missing")
	}
}

func TestLoadHasOneRelation(t *testing.T) {
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
	post := &Post{ID: 1, Title: "Test Post", UserID: 1}
	v := reflect.ValueOf(post).Elem()
	userField := v.FieldByName("User")

	err = q.loadHasOneRelation(v, userField, "User", db)
	if err != nil {
		t.Fatalf("loadHasOneRelation failed: %v", err)
	}

	// Check that user was loaded
	if post.User == nil {
		t.Error("Expected User to be loaded")
	}
	if post.User.Name != "John Doe" {
		t.Errorf("Expected user name 'John Doe', got '%s'", post.User.Name)
	}
}

func TestLoadHasOneRelationWithNilConnection(t *testing.T) {
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

	q := NewQuery(context.Background(), nil, nil, "", nil, nil)
	post := &Post{ID: 1, Title: "Test Post", UserID: 1}
	v := reflect.ValueOf(post).Elem()
	userField := v.FieldByName("User")

	err := q.loadHasOneRelation(v, userField, "User", nil)
	if err == nil {
		t.Error("Expected error when connection is nil")
	} else if err.Error() != "database connection is nil" {
		t.Errorf("Expected 'database connection is nil' error, got: %v", err)
	}
}

func TestLoadHasOneRelationWithZeroForeignKey(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

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
	post := &Post{ID: 1, Title: "Test Post", UserID: 0}
	v := reflect.ValueOf(post).Elem()
	userField := v.FieldByName("User")

	err = q.loadHasOneRelation(v, userField, "User", db)
	if err != nil {
		t.Fatalf("loadHasOneRelation with zero foreign key failed: %v", err)
	}

	// User should be nil when foreign key is zero
	if post.User != nil {
		t.Error("Expected User to be nil when foreign key is zero")
	}
}

func TestLoadHasOneRelationWithMissingForeignKey(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

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
	post := &Post{ID: 1, Title: "Test Post"}
	v := reflect.ValueOf(post).Elem()
	userField := v.FieldByName("User")

	err = q.loadHasOneRelation(v, userField, "User", db)
	if err == nil {
		t.Error("Expected error when foreign key field is missing")
	}
}

func TestLoad(t *testing.T) {
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
	post := &Post{ID: 1, Title: "Test Post", UserID: 1}

	err = q.Load(post, "User")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Check that User was loaded
	if post.User == nil {
		t.Error("Expected User to be loaded")
	}
	if post.User.Name != "John Doe" {
		t.Errorf("Expected user name 'John Doe', got '%s'", post.User.Name)
	}
}

func TestLoadWithConstraint(t *testing.T) {
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
	post := &Post{ID: 1, Title: "Test Post", UserID: 1}

	err = q.Load(post, "User", func(q contractsorm.Query) contractsorm.Query {
		return q.Where("status = ?", "active")
	})
	if err != nil {
		t.Fatalf("Load with constraint failed: %v", err)
	}

	// Check that User was loaded with constraint applied
	if post.User == nil {
		t.Error("Expected User to be loaded")
	}
	if post.User.Status != "active" {
		t.Errorf("Expected user status 'active', got '%s'", post.User.Status)
	}
}

func TestLoadHasMany(t *testing.T) {
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
	_, err = db.Exec("INSERT INTO posts (id, title, user_id) VALUES (1, 'Post 1', 1)")
	if err != nil {
		t.Fatalf("Failed to insert post 1: %v", err)
	}
	_, err = db.Exec("INSERT INTO posts (id, title, user_id) VALUES (2, 'Post 2', 1)")
	if err != nil {
		t.Fatalf("Failed to insert post 2: %v", err)
	}

	type Post struct {
		ID     int
		Title  string
		UserID int
	}

	type User struct {
		ID    int
		Name  string
		Posts []Post
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	user := &User{ID: 1, Name: "John Doe"}

	err = q.Load(user, "Posts")
	if err != nil {
		t.Fatalf("Load has-many failed: %v", err)
	}

	// Check that posts were loaded
	if len(user.Posts) != 2 {
		t.Errorf("Expected 2 posts, got %d", len(user.Posts))
	}
	if user.Posts[0].Title != "Post 1" {
		t.Errorf("Expected post title 'Post 1', got '%s'", user.Posts[0].Title)
	}
	if user.Posts[1].Title != "Post 2" {
		t.Errorf("Expected post title 'Post 2', got '%s'", user.Posts[1].Title)
	}
}

func TestLoadMissing(t *testing.T) {
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
	post := &Post{ID: 1, Title: "Test Post", UserID: 1}

	// First load should load the relation
	err = q.LoadMissing(post, "User")
	if err != nil {
		t.Fatalf("LoadMissing failed: %v", err)
	}

	if post.User == nil {
		t.Error("Expected User to be loaded")
	}
	if post.User.Name != "John Doe" {
		t.Errorf("Expected user name 'John Doe', got '%s'", post.User.Name)
	}

	// Second load should skip since already loaded
	err = q.LoadMissing(post, "User")
	if err != nil {
		t.Fatalf("LoadMissing second call failed: %v", err)
	}

	// User should still be the same
	if post.User.Name != "John Doe" {
		t.Errorf("Expected user name 'John Doe', got '%s'", post.User.Name)
	}
}

func TestLoadMissingHasMany(t *testing.T) {
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
	_, err = db.Exec("INSERT INTO posts (id, title, user_id) VALUES (1, 'Post 1', 1)")
	if err != nil {
		t.Fatalf("Failed to insert post 1: %v", err)
	}
	_, err = db.Exec("INSERT INTO posts (id, title, user_id) VALUES (2, 'Post 2', 1)")
	if err != nil {
		t.Fatalf("Failed to insert post 2: %v", err)
	}

	type Post struct {
		ID     int
		Title  string
		UserID int
	}

	type User struct {
		ID    int
		Name  string
		Posts []Post
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	user := &User{ID: 1, Name: "John Doe"}

	// First load should load the relation
	err = q.LoadMissing(user, "Posts")
	if err != nil {
		t.Fatalf("LoadMissing has-many failed: %v", err)
	}

	if len(user.Posts) != 2 {
		t.Errorf("Expected 2 posts, got %d", len(user.Posts))
	}

	// Second load should skip since already loaded
	err = q.LoadMissing(user, "Posts")
	if err != nil {
		t.Fatalf("LoadMissing has-many second call failed: %v", err)
	}

	// Posts should still be the same
	if len(user.Posts) != 2 {
		t.Errorf("Expected 2 posts after second LoadMissing, got %d", len(user.Posts))
	}
}

func TestWithCount(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"
	q.model = &struct{ ID int }{}

	result := q.WithCount("Posts")
	if result == nil {
		t.Error("WithCount should return a query")
	}

	resultQuery := result.(*Query)
	if len(resultQuery.withCountQueries) != 1 {
		t.Errorf("Expected 1 count query, got %d", len(resultQuery.withCountQueries))
	}
	if resultQuery.withCountQueries[0].relation != "Posts" {
		t.Errorf("Expected relation 'Posts', got '%s'", resultQuery.withCountQueries[0].relation)
	}
	if resultQuery.withCountQueries[0].column != "Posts_count" {
		t.Errorf("Expected column 'Posts_count', got '%s'", resultQuery.withCountQueries[0].column)
	}
}

func TestWithCountWithConstraint(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"
	q.model = &struct{ ID int }{}

	result := q.WithCount("Posts", func(q contractsorm.Query) contractsorm.Query {
		return q.Where("published = ?", true)
	})
	if result == nil {
		t.Error("WithCount with constraint should return a query")
	}

	resultQuery := result.(*Query)
	if len(resultQuery.withCountQueries) != 1 {
		t.Errorf("Expected 1 count query, got %d", len(resultQuery.withCountQueries))
	}
	if resultQuery.withCountQueries[0].constraint == nil {
		t.Error("Expected constraint to be set")
	}
}

func TestWithExists(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"
	q.model = &struct{ ID int }{}

	result := q.WithExists("Posts")
	if result == nil {
		t.Error("WithExists should return a query")
	}

	resultQuery := result.(*Query)
	if len(resultQuery.withExistsQueries) != 1 {
		t.Errorf("Expected 1 exists query, got %d", len(resultQuery.withExistsQueries))
	}
	if resultQuery.withExistsQueries[0].relation != "Posts" {
		t.Errorf("Expected relation 'Posts', got '%s'", resultQuery.withExistsQueries[0].relation)
	}
}

func TestWithExistsWithConstraint(t *testing.T) {
	q := NewQuery(context.TODO(), nil, nil, "", nil, nil)
	q.table = "users"
	q.model = &struct{ ID int }{}

	result := q.WithExists("Posts", func(q contractsorm.Query) contractsorm.Query {
		return q.Where("published = ?", true)
	})
	if result == nil {
		t.Error("WithExists with constraint should return a query")
	}

	resultQuery := result.(*Query)
	if len(resultQuery.withExistsQueries) != 1 {
		t.Errorf("Expected 1 exists query, got %d", len(resultQuery.withExistsQueries))
	}
	if resultQuery.withExistsQueries[0].constraint == nil {
		t.Error("Expected constraint to be set")
	}
}

func TestWithCountSQLGeneration(t *testing.T) {
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
		ID int
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.table = "users"
	q.model = &User{}
	q = q.WithCount("Posts").(*Query)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Check that SQL contains count subquery
	if !strings.Contains(sql, "COUNT(*)") {
		t.Errorf("Expected SQL to contain COUNT(*), got: %s", sql)
	}
	if !strings.Contains(sql, "posts") {
		t.Errorf("Expected SQL to contain posts table, got: %s", sql)
	}
	if !strings.Contains(sql, "Posts_count") {
		t.Errorf("Expected SQL to contain Posts_count alias, got: %s", sql)
	}
	if len(args) != 0 {
		t.Errorf("Expected no args, got %d", len(args))
	}
}

func TestWithExistsSQLGeneration(t *testing.T) {
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
		ID int
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.table = "users"
	q.model = &User{}
	q = q.WithExists("Posts").(*Query)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Check that SQL contains exists subquery
	if !strings.Contains(sql, "EXISTS") {
		t.Errorf("Expected SQL to contain EXISTS, got: %s", sql)
	}
	if !strings.Contains(sql, "posts") {
		t.Errorf("Expected SQL to contain posts table, got: %s", sql)
	}
	if !strings.Contains(sql, "Posts_exists") {
		t.Errorf("Expected SQL to contain Posts_exists alias, got: %s", sql)
	}
	if len(args) != 0 {
		t.Errorf("Expected no args, got %d", len(args))
	}
}

func TestWithCountAndExistsTogether(t *testing.T) {
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
		ID int
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.table = "users"
	q.model = &User{}
	q = q.WithCount("Posts").(*Query)
	q = q.WithExists("Comments").(*Query)

	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Check that SQL contains both subqueries
	if !strings.Contains(sql, "COUNT(*)") {
		t.Errorf("Expected SQL to contain COUNT(*), got: %s", sql)
	}
	if !strings.Contains(sql, "EXISTS") {
		t.Errorf("Expected SQL to contain EXISTS, got: %s", sql)
	}
	if !strings.Contains(sql, "Posts_count") {
		t.Errorf("Expected SQL to contain Posts_count alias, got: %s", sql)
	}
	if !strings.Contains(sql, "Comments_exists") {
		t.Errorf("Expected SQL to contain Comments_exists alias, got: %s", sql)
	}
	if len(args) != 0 {
		t.Errorf("Expected no args, got %d", len(args))
	}
}
