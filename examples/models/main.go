package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/database/schema"
)

// User represents a user model
type User struct {
	ID        uint       `gorm:"column:id;primaryKey"`
	Name      string     `gorm:"column:name"`
	Email     string     `gorm:"column:email"`
	Age       int        `gorm:"column:age"`
	Status    string     `gorm:"column:status"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}

// Post represents a post model
type Post struct {
	ID        uint      `gorm:"column:id;primaryKey"`
	UserID    uint      `gorm:"column:user_id"`
	Title     string    `gorm:"column:title"`
	Content   string    `gorm:"column:content"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName specifies the table name for the Post model
func (Post) TableName() string {
	return "posts"
}

// This example demonstrates using struct-based models with the ORM
func main() {
	if err := RunExample("sqlite://./example.db"); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

// RunExample demonstrates struct-based model operations
func RunExample(dsn string) error {
	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Create tables for the example
	err = db.Schema().Create("users", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("name")
		blueprint.String("email")
		blueprint.Integer("age")
		blueprint.String("status")
		blueprint.Timestamps()
		blueprint.SoftDeletes()
	})
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	err = db.Schema().Create("posts", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.Integer("user_id")
		blueprint.String("title")
		blueprint.Text("content")
		blueprint.Timestamps()
	})
	if err != nil {
		return fmt.Errorf("failed to create posts table: %w", err)
	}

	// Create a new user using model
	fmt.Println("=== Create User with Model ===")
	user := &User{
		Name:   "John Doe",
		Email:  "john@example.com",
		Age:    30,
		Status: "active",
	}
	err = db.Query().Model(&User{}).Create(user)
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	} else {
		fmt.Printf("User created with ID: %d\n", user.ID)
	}

	// Find user by ID using model
	fmt.Println("\n=== Find User by ID ===")
	var foundUser User
	err = db.Query().Model(&User{}).Where("id = ?", user.ID).First(&foundUser)
	if err != nil {
		return fmt.Errorf("error finding user: %v", err)
	} else {
		fmt.Printf("Found user: %+v\n", foundUser)
	}

	// Update user using model
	fmt.Println("\n=== Update User ===")
	foundUser.Name = "Jane Doe"
	foundUser.Age = 31
	_, err = db.Query().Model(&User{}).Where("id = ?", foundUser.ID).Update(&foundUser)
	if err != nil {
		return fmt.Errorf("error updating user: %v", err)
	} else {
		fmt.Println("User updated successfully")
	}

	// Query multiple users
	fmt.Println("\n=== Query Multiple Users ===")
	var users []User
	err = db.Query().Model(&User{}).Where("age > ?", 18).Get(&users)
	if err != nil {
		return fmt.Errorf("error querying users: %v", err)
	} else {
		fmt.Printf("Found %d users\n", len(users))
	}

	// Create a post with foreign key
	fmt.Println("\n=== Create Post with Foreign Key ===")
	post := &Post{
		UserID:  user.ID,
		Title:   "My First Post",
		Content: "This is the content of my first post",
	}
	err = db.Query().Model(&Post{}).Create(post)
	if err != nil {
		return fmt.Errorf("error creating post: %v", err)
	} else {
		fmt.Printf("Post created with ID: %d\n", post.ID)
	}

	// Delete user (soft delete)
	fmt.Println("\n=== Soft Delete User ===")
	_, err = db.Query().Model(&User{}).Delete(&foundUser)
	if err != nil {
		return fmt.Errorf("error deleting user: %v", err)
	} else {
		fmt.Println("User soft deleted successfully")
	}

	return nil
}
