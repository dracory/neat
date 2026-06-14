package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/schema/constants"
	"github.com/dracory/neat/database/soft_delete"
)

// Post model using NULL-based soft delete strategy with deleted_at column
// This is Laravel-compatible and uses the traditional constants.DeletedAtColumnName column name
type Post struct {
	soft_delete.DeletedAt
	ID        uint      `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	Author    string    `json:"author" db:"author"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// This example demonstrates NULL-based soft delete with deleted_at column (Laravel-compatible)
func main() {
	if err := RunExample("sqlite://./deleted_at_example.db"); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

// RunExample demonstrates NULL-based soft delete with deleted_at column usage
func RunExample(dsn string) error {
	db, err := neat.NewFromDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Create posts table with nullable deleted_at column
	// Default value is NULL (indicates active record)
	err = db.Schema().Create("posts", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("title")
		blueprint.Text("content")
		blueprint.String("author")
		blueprint.Timestamp(constants.DefaultCreatedAtColumn)
		// For Laravel compatibility, deleted_at should be nullable
		blueprint.Timestamp(constants.DeletedAtColumnName).Nullable()
	})
	if err != nil {
		return fmt.Errorf("failed to create posts table: %w", err)
	}

	// Create some posts
	fmt.Println("=== Creating Posts ===")
	posts := []map[string]any{
		{
			"title":      "Getting Started with Go",
			"content":    "Go is a programming language...",
			"author":     "John Doe",
			"created_at": time.Now(),
		},
		{
			"title":      "Advanced Go Patterns",
			"content":    "Learn advanced Go programming patterns...",
			"author":     "Jane Smith",
			"created_at": time.Now(),
		},
		{
			"title":      "Go Web Development",
			"content":    "Building web applications with Go...",
			"author":     "Bob Johnson",
			"created_at": time.Now(),
		},
	}

	for _, p := range posts {
		err = db.Query().Table("posts").Create(p)
		if err != nil {
			return fmt.Errorf("failed to create post: %w", err)
		}
	}
	fmt.Printf("Created %d posts\n", len(posts))

	// Count all active posts (not soft deleted)
	fmt.Println("\n=== Active Posts (Default Query) ===")
	var count int64
	err = db.Query().Model(&Post{}).Count(&count)
	if err != nil {
		return fmt.Errorf("failed to count posts: %w", err)
	}
	fmt.Printf("Active posts: %d\n", count)

	// List all active posts
	var activePosts []Post
	err = db.Query().Model(&Post{}).Get(&activePosts)
	if err != nil {
		return fmt.Errorf("failed to get posts: %w", err)
	}
	for _, p := range activePosts {
		fmt.Printf("  - %s by %s\n", p.Title, p.Author)
	}

	// Soft delete a post (Getting Started with Go)
	fmt.Println("\n=== Soft Deleting 'Getting Started with Go' ===")
	result, err := db.Query().Model(&Post{}).Where("title = ?", "Getting Started with Go").Delete()
	if err != nil {
		return fmt.Errorf("failed to soft delete: %w", err)
	}
	fmt.Printf("Soft deleted %d post(s)\n", result.RowsAffected)

	// Count active posts after soft delete
	fmt.Println("\n=== Active Posts After Soft Delete ===")
	err = db.Query().Model(&Post{}).Count(&count)
	if err != nil {
		return fmt.Errorf("failed to count: %w", err)
	}
	fmt.Printf("Active posts: %d (Getting Started with Go is now hidden)\n", count)

	// List active posts (should not include Getting Started with Go)
	var remainingPosts []Post
	err = db.Query().Model(&Post{}).Get(&remainingPosts)
	if err != nil {
		return fmt.Errorf("failed to get posts: %w", err)
	}
	for _, p := range remainingPosts {
		fmt.Printf("  - %s by %s\n", p.Title, p.Author)
	}

	// Include soft deleted posts with WithTrashed()
	fmt.Println("\n=== All Posts Including Soft Deleted (WithTrashed) ===")
	var allPosts []Post
	err = db.Query().Model(&Post{}).WithSoftDeleted().Get(&allPosts)
	if err != nil {
		return fmt.Errorf("failed to get all posts: %w", err)
	}
	fmt.Printf("Total posts (including soft deleted): %d\n", len(allPosts))
	for _, p := range allPosts {
		status := "active"
		if p.IsSoftDeleted() {
			status = "soft deleted"
		}
		fmt.Printf("  - %s by %s [%s]\n", p.Title, p.Author, status)
	}

	// Query only soft deleted posts
	fmt.Println("\n=== Only Soft Deleted Posts (OnlySoftDeleted) ===")
	var deletedPosts []Post
	err = db.Query().Model(&Post{}).OnlySoftDeleted().Get(&deletedPosts)
	if err != nil {
		return fmt.Errorf("failed to get deleted posts: %w", err)
	}
	fmt.Printf("Soft deleted posts: %d\n", len(deletedPosts))
	for _, p := range deletedPosts {
		if p.DeletedAt.DeletedAt != nil {
			fmt.Printf("  - %s was soft deleted at %s\n", p.Title, p.DeletedAt.DeletedAt.Format(time.RFC3339))
		}
	}

	// Restore the soft deleted post
	fmt.Println("\n=== Restoring Soft Deleted Post ===")
	var gettingStarted Post
	err = db.Query().Model(&Post{}).OnlySoftDeleted().Where("title = ?", "Getting Started with Go").First(&gettingStarted)
	if err != nil {
		return fmt.Errorf("failed to find soft deleted post: %w", err)
	}

	restoreResult, err := db.Query().Model(&Post{}).Where("id = ?", gettingStarted.ID).RestoreSoftDeleted()
	if err != nil {
		return fmt.Errorf("failed to restore: %w", err)
	}
	fmt.Printf("Restored %d post(s)\n", restoreResult.RowsAffected)

	// Verify restoration
	fmt.Println("\n=== Active Posts After Restore ===")
	err = db.Query().Model(&Post{}).Count(&count)
	if err != nil {
		return fmt.Errorf("failed to count: %w", err)
	}
	fmt.Printf("Active posts: %d (Getting Started with Go is back!)\n", count)

	// Show the restored post
	var restoredPost Post
	err = db.Query().Model(&Post{}).Where("title = ?", "Getting Started with Go").First(&restoredPost)
	if err != nil {
		return fmt.Errorf("failed to find restored post: %w", err)
	}
	fmt.Printf("Restored: %s by %s - DeletedAt: %v\n",
		restoredPost.Title,
		restoredPost.Author,
		restoredPost.DeletedAt.DeletedAt)

	// Demonstrate force delete (permanent deletion)
	fmt.Println("\n=== Force Delete (Permanent Deletion) ===")
	forceResult, err := db.Query().Model(&Post{}).Where("title = ?", "Advanced Go Patterns").ForceDelete()
	if err != nil {
		return fmt.Errorf("failed to force delete: %w", err)
	}
	fmt.Printf("Force deleted %d post(s) permanently\n", forceResult.RowsAffected)

	// Final count
	fmt.Println("\n=== Final Active Post Count ===")
	err = db.Query().Model(&Post{}).WithSoftDeleted().Count(&count)
	if err != nil {
		return fmt.Errorf("failed to count: %w", err)
	}
	fmt.Printf("Total posts in database: %d (Advanced Go Patterns was permanently deleted)\n", count)

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("Laravel-compatible soft delete benefits:")
	fmt.Println("  - Uses traditional 'deleted_at' column name")
	fmt.Println("  - Compatible with Laravel schemas and conventions")
	fmt.Println("  - Easy migration from Laravel to Go")
	fmt.Println("  - Familiar pattern for Laravel developers")

	return nil
}
