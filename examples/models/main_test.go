package main_test

import (
	"testing"
	"time"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/database/schema"
	mainpkg "github.com/dracory/neat/examples/models"
)

func TestRunExample(t *testing.T) {
	// Use in-memory SQLite for testing
	err := mainpkg.RunExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunExample failed: %v", err)
	}
}

// localUser mirrors the example model for test assertions
type localUser struct {
	ID        uint       `gorm:"column:id;primaryKey"`
	Name      string     `gorm:"column:name"`
	Email     string     `gorm:"column:email"`
	Age       int        `gorm:"column:age"`
	Status    string     `gorm:"column:status"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

func (localUser) TableName() string { return "users" }

type localPost struct {
	ID      uint   `gorm:"column:id;primaryKey"`
	UserID  uint   `gorm:"column:user_id"`
	Title   string `gorm:"column:title"`
	Content string `gorm:"column:content"`
}

func (localPost) TableName() string { return "posts" }

func setupModelsDB(t *testing.T) *neat.Database {
	t.Helper()
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	err = db.Schema().Create("users", func(bp schema.Blueprint) {
		bp.ID()
		bp.String("name")
		bp.String("email")
		bp.Integer("age")
		bp.String("status")
		bp.Timestamps()
		bp.SoftDeletes()
	})
	if err != nil {
		t.Fatalf("failed to create users: %v", err)
	}

	err = db.Schema().Create("posts", func(bp schema.Blueprint) {
		bp.ID()
		bp.Integer("user_id")
		bp.String("title")
		bp.Text("content")
		bp.Timestamps()
	})
	if err != nil {
		t.Fatalf("failed to create posts: %v", err)
	}

	return db
}

func TestModels_Create_IDAssigned(t *testing.T) {
	db := setupModelsDB(t)
	defer func() { _ = db.Close() }()

	user := &localUser{Name: "John", Email: "john@example.com", Age: 30, Status: "active"}
	err := db.Query().Model(&localUser{}).Create(user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if user.ID == 0 {
		t.Error("expected ID to be assigned after Create, got 0")
	}
}

func TestModels_FindByID(t *testing.T) {
	db := setupModelsDB(t)
	defer func() { _ = db.Close() }()

	user := &localUser{Name: "Alice", Email: "alice@example.com", Age: 25, Status: "active"}
	if err := db.Query().Model(&localUser{}).Create(user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var found localUser
	err := db.Query().Model(&localUser{}).Where("id = ?", user.ID).First(&found)
	if err != nil {
		t.Fatalf("First failed: %v", err)
	}
	if found.Name != "Alice" {
		t.Errorf("expected name 'Alice', got '%s'", found.Name)
	}
	if found.Email != "alice@example.com" {
		t.Errorf("expected email 'alice@example.com', got '%s'", found.Email)
	}
}

func TestModels_Update(t *testing.T) {
	db := setupModelsDB(t)
	defer func() { _ = db.Close() }()

	user := &localUser{Name: "Bob", Email: "bob@example.com", Age: 30, Status: "active"}
	if err := db.Query().Model(&localUser{}).Create(user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	user.Name = "Robert"
	user.Age = 31
	_, err := db.Query().Model(&localUser{}).Where("id = ?", user.ID).Update(user)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	var found localUser
	if err = db.Query().Model(&localUser{}).Where("id = ?", user.ID).First(&found); err != nil {
		t.Fatalf("First after update failed: %v", err)
	}
	if found.Name != "Robert" {
		t.Errorf("expected name 'Robert' after update, got '%s'", found.Name)
	}
	if found.Age != 31 {
		t.Errorf("expected age 31 after update, got %d", found.Age)
	}
}

func TestModels_Post_ForeignKey(t *testing.T) {
	db := setupModelsDB(t)
	defer func() { _ = db.Close() }()

	user := &localUser{Name: "Carol", Email: "carol@example.com", Age: 28, Status: "active"}
	if err := db.Query().Model(&localUser{}).Create(user); err != nil {
		t.Fatalf("Create user failed: %v", err)
	}

	post := &localPost{UserID: user.ID, Title: "Hello World", Content: "Some content"}
	if err := db.Query().Model(&localPost{}).Create(post); err != nil {
		t.Fatalf("Create post failed: %v", err)
	}
	if post.ID == 0 {
		t.Error("expected post ID to be assigned, got 0")
	}

	var found localPost
	if err := db.Query().Model(&localPost{}).Where("user_id = ?", user.ID).First(&found); err != nil {
		t.Fatalf("First post failed: %v", err)
	}
	if found.Title != "Hello World" {
		t.Errorf("expected title 'Hello World', got '%s'", found.Title)
	}
}
