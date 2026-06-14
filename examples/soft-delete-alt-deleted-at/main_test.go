package main_test

import (
	"testing"
	"time"

	"github.com/dracory/neat"
	"github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/schema/constants"
	"github.com/dracory/neat/database/soft_delete"
	mainpkg "github.com/dracory/neat/examples/soft-delete-alt-deleted-at"
)

// Post model using NULL-based soft delete strategy with deleted_at column
type Post struct {
	soft_delete.DeletedAt
	ID        uint      `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	Author    string    `json:"author" db:"author"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func TestRunExample(t *testing.T) {
	err := mainpkg.RunExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunExample failed: %v", err)
	}
}

func TestDeletedAtSoftDelete_CreateAndSoftDelete(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	err = db.Schema().Create("posts", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("title")
		blueprint.Text("content")
		blueprint.String("author")
		blueprint.Timestamp(constants.DefaultCreatedAtColumn)
		blueprint.Timestamp(constants.DeletedAtColumnName).Nullable()
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Create a post
	err = db.Query().Table("posts").Create(map[string]any{
		"title":      "Test Post",
		"content":    "A test post",
		"author":     "Test Author",
		"created_at": time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}

	// Count should be 1
	var count int64
	err = db.Query().Model(&Post{}).Count(&count)
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 post, got %d", count)
	}

	// Soft delete the post
	_, err = db.Query().Model(&Post{}).Where("title = ?", "Test Post").Delete()
	if err != nil {
		t.Fatalf("failed to soft delete: %v", err)
	}

	// Count should be 0 (soft deleted posts are hidden by default)
	err = db.Query().Model(&Post{}).Count(&count)
	if err != nil {
		t.Fatalf("failed to count after soft delete: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 active posts after soft delete, got %d", count)
	}

	// WithTrashed should show 1
	err = db.Query().Model(&Post{}).WithSoftDeleted().Count(&count)
	if err != nil {
		t.Fatalf("failed to count with trashed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 post with trashed, got %d", count)
	}
}

func TestDeletedAtSoftDelete_Restore(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	err = db.Schema().Create("posts", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("title")
		blueprint.Text("content")
		blueprint.String("author")
		blueprint.Timestamp(constants.DeletedAtColumnName).Nullable()
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Create and soft delete a post
	err = db.Query().Table("posts").Create(map[string]any{
		"title":   "Restorable Post",
		"content": "This post can be restored",
		"author":  "Test Author",
	})
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	_, err = db.Query().Model(&Post{}).Where("title = ?", "Restorable Post").Delete()
	if err != nil {
		t.Fatalf("failed to soft delete: %v", err)
	}

	// Verify it's soft deleted
	var post Post
	err = db.Query().Model(&Post{}).OnlySoftDeleted().Where("title = ?", "Restorable Post").First(&post)
	if err != nil {
		t.Fatalf("failed to find soft deleted post: %v", err)
	}
	if !post.IsSoftDeleted() {
		t.Error("expected post to be soft deleted")
	}

	// Restore the post
	_, err = db.Query().Model(&Post{}).Where("title = ?", "Restorable Post").RestoreSoftDeleted()
	if err != nil {
		t.Fatalf("failed to restore: %v", err)
	}

	// Verify it's restored
	var restored Post
	err = db.Query().Model(&Post{}).Where("title = ?", "Restorable Post").First(&restored)
	if err != nil {
		t.Fatalf("failed to find restored post: %v", err)
	}
	if restored.IsSoftDeleted() {
		t.Error("expected post to be restored (not soft deleted)")
	}

	// Verify DeletedAt is NULL (nil)
	if restored.DeletedAt.DeletedAt != nil {
		t.Errorf("expected DeletedAt to be nil, got %v", restored.DeletedAt.DeletedAt)
	}
}

func TestDeletedAtSoftDelete_ForceDelete(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	err = db.Schema().Create("posts", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("title")
		blueprint.Text("content")
		blueprint.String("author")
		blueprint.Timestamp(constants.DeletedAtColumnName).Nullable()
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Create a post
	err = db.Query().Table("posts").Create(map[string]any{
		"title":   "Deletable Post",
		"content": "This post will be permanently deleted",
		"author":  "Test Author",
	})
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	// Force delete (permanent)
	_, err = db.Query().Model(&Post{}).Where("title = ?", "Deletable Post").ForceDelete()
	if err != nil {
		t.Fatalf("failed to force delete: %v", err)
	}

	// Count with trashed should be 0 (permanently deleted)
	var count int64
	err = db.Query().Model(&Post{}).WithSoftDeleted().Count(&count)
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 posts after force delete, got %d", count)
	}
}

func TestDeletedAtSoftDelete_OnlySoftDeleted(t *testing.T) {
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	err = db.Schema().Create("posts", func(blueprint schema.Blueprint) {
		blueprint.ID()
		blueprint.String("title")
		blueprint.Text("content")
		blueprint.String("author")
		blueprint.Timestamp(constants.DeletedAtColumnName).Nullable()
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Create active and soft deleted posts
	posts := []map[string]any{
		{"title": "Active Post", "content": "Active content", "author": "Author 1"},
		{"title": "Deleted Post 1", "content": "Deleted content 1", "author": "Author 2"},
		{"title": "Deleted Post 2", "content": "Deleted content 2", "author": "Author 3"},
	}
	for _, p := range posts {
		err = db.Query().Table("posts").Create(p)
		if err != nil {
			t.Fatalf("failed to create: %v", err)
		}
	}

	// Soft delete two posts
	_, err = db.Query().Model(&Post{}).Where("title LIKE ?", "Deleted%").Delete()
	if err != nil {
		t.Fatalf("failed to soft delete: %v", err)
	}

	// OnlySoftDeleted should return 2
	var deletedPosts []Post
	err = db.Query().Model(&Post{}).OnlySoftDeleted().Get(&deletedPosts)
	if err != nil {
		t.Fatalf("failed to get deleted posts: %v", err)
	}
	if len(deletedPosts) != 2 {
		t.Errorf("expected 2 soft deleted posts, got %d", len(deletedPosts))
	}

	// Default query should return 1
	var activePosts []Post
	err = db.Query().Model(&Post{}).Get(&activePosts)
	if err != nil {
		t.Fatalf("failed to get active posts: %v", err)
	}
	if len(activePosts) != 1 {
		t.Errorf("expected 1 active post, got %d", len(activePosts))
	}
	if activePosts[0].Title != "Active Post" {
		t.Errorf("expected 'Active Post', got '%s'", activePosts[0].Title)
	}
}

func TestDeletedAtSoftDelete_IsSoftDeleted(t *testing.T) {
	// Verify IsSoftDeleted logic for NULL-based strategy with deleted_at column
	post := Post{}

	// NULL (nil) means NOT soft deleted
	if post.IsSoftDeleted() {
		t.Error("post with nil DeletedAt should not be soft deleted")
	}

	// Non-NULL timestamp means soft deleted
	now := time.Now()
	post.DeletedAt.DeletedAt = &now
	if !post.IsSoftDeleted() {
		t.Error("post with non-nil DeletedAt should be soft deleted")
	}
}

func TestDeletedAtSoftDelete_DeletedAtField(t *testing.T) {
	// Verify the DeletedAt field exists and is properly named
	post := Post{}

	// Set DeletedAt to a specific time
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	post.DeletedAt.DeletedAt = &testTime

	if post.DeletedAt.DeletedAt == nil || !post.DeletedAt.DeletedAt.Equal(testTime) {
		t.Errorf("expected DeletedAt to be %v, got %v", testTime, post.DeletedAt.DeletedAt)
	}
}

func TestDeletedAtSoftDelete_ColumnName(t *testing.T) {
	// Verify the column name is constants.DeletedAtColumnName for Laravel compatibility
	post := Post{}

	columnName := post.SoftDeletedAtColumn()
	if columnName != constants.DeletedAtColumnName {
		t.Errorf("expected column name to be 'deleted_at', got '%s'", columnName)
	}
}
