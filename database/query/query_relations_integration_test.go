package query

import (
	"context"
	"database/sql"
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	_ "modernc.org/sqlite"
)

func TestEagerLoadingWithSingleRelation(t *testing.T) {
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

	var posts []Post
	err = q.Get(&posts)
	if err != nil {
		t.Fatalf("Get with eager loading failed: %v", err)
	}

	if len(posts) != 1 {
		t.Errorf("Expected 1 post, got %d", len(posts))
	}

	if posts[0].User == nil {
		t.Error("Expected User to be loaded via eager loading")
	}

	if posts[0].User.Name != "John Doe" {
		t.Errorf("Expected user name 'John Doe', got '%s'", posts[0].User.Name)
	}
}

func TestEagerLoadingWithMultipleRelations(t *testing.T) {
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
	_, err = db.Exec("CREATE TABLE comments (id INTEGER PRIMARY KEY, text TEXT, post_id INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create comments table: %v", err)
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
	_, err = db.Exec("INSERT INTO comments (id, text, post_id) VALUES (1, 'Test Comment', 1)")
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
		PostID int
	}

	type Post struct {
		ID       int
		Title    string
		UserID   int
		User     *User
		Comments []Comment
	}

	q := NewQuery(context.Background(), db, nil, "", nil, nil)
	q.table = "posts"
	q.withRelations = []string{"User", "Comments"}

	// Verify tables exist
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='comments'").Scan(&tableName)
	if err != nil {
		t.Fatalf("comments table not found: %v", err)
	}

	var posts []Post
	err = q.Get(&posts)
	if err != nil {
		t.Fatalf("Get with multiple eager loading failed: %v", err)
	}

	if len(posts) != 1 {
		t.Errorf("Expected 1 post, got %d", len(posts))
	}

	if posts[0].User == nil {
		t.Error("Expected User to be loaded via eager loading")
	}

	if len(posts[0].Comments) == 0 {
		t.Error("Expected Comments to be loaded via eager loading")
	}
}

func TestEagerLoadingWithWhereClause(t *testing.T) {
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

	var posts []Post
	err = q.Get(&posts)
	if err != nil {
		t.Fatalf("Get with eager loading and constraint failed: %v", err)
	}

	if len(posts) != 1 {
		t.Errorf("Expected 1 post, got %d", len(posts))
	}

	if posts[0].User == nil {
		t.Error("Expected User to be loaded via eager loading")
	}

	if posts[0].User.Status != "active" {
		t.Errorf("Expected user status 'active', got '%s'", posts[0].User.Status)
	}
}

func TestEagerLoadingWithNoMatchingRelation(t *testing.T) {
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

	// Insert test data - post with non-existent user
	_, err = db.Exec("INSERT INTO posts (id, title, user_id) VALUES (1, 'Test Post', 999)")
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

	var posts []Post
	err = q.Get(&posts)
	if err != nil {
		t.Fatalf("Get with eager loading failed: %v", err)
	}

	if len(posts) != 1 {
		t.Errorf("Expected 1 post, got %d", len(posts))
	}

	// User should be nil when no matching relation exists
	if posts[0].User != nil {
		t.Error("Expected User to be nil when no matching relation exists")
	}
}

func TestEagerLoadingWithFirst(t *testing.T) {
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

	var post Post
	err = q.First(&post)
	if err != nil {
		t.Fatalf("First with eager loading failed: %v", err)
	}

	if post.User == nil {
		t.Error("Expected User to be loaded via eager loading with First")
	}

	if post.User.Name != "John Doe" {
		t.Errorf("Expected user name 'John Doe', got '%s'", post.User.Name)
	}
}
