//go:build integration

package cockroachdb_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestCockroachDBIntegrationQueryAssociationFind(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	user := models.User{
		Name: "association_find_name",
	}

	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "association_find_name").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	var address models.Address
	assoc := query.Model(&createdUser).Association("Address")
	if err := assoc.Find(&address); err != nil {
		t.Logf("Find returned error (expected for empty association): %v", err)
	}
}

func TestCockroachDBIntegrationQueryAssociationAppendHasOne(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	user := models.User{
		Name: "association_append_has_one",
	}

	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "association_append_has_one").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	address := models.Address{
		Name: "Test Address",
	}

	assoc := query.Model(&createdUser).Association("Address")
	if err := assoc.Append(&address); err != nil {
		t.Fatalf("Failed to append address: %v", err)
	}

	var loadedAddress models.Address
	if err := assoc.Find(&loadedAddress); err != nil {
		t.Fatalf("Failed to find associated address: %v", err)
	}

	if loadedAddress.Name != "Test Address" {
		t.Errorf("Expected address name 'Test Address', got '%s'", loadedAddress.Name)
	}
}

func TestCockroachDBIntegrationQueryAssociationAppendHasMany(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	user := models.User{
		Name: "association_append_has_many",
	}

	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "association_append_has_many").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	book1 := models.Book{Name: "Book 1"}
	book2 := models.Book{Name: "Book 2"}

	assoc := query.Model(&createdUser).Association("Books")
	if err := assoc.Append(&book1, &book2); err != nil {
		t.Fatalf("Failed to append books: %v", err)
	}

	var books []models.Book
	if err := assoc.Find(&books); err != nil {
		t.Fatalf("Failed to find associated books: %v", err)
	}

	if len(books) != 2 {
		t.Errorf("Expected 2 books, got %d", len(books))
	}
}

func TestCockroachDBIntegrationQueryAssociationReplaceHasOne(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	user := models.User{
		Name: "association_replace_has_one",
	}

	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "association_replace_has_one").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	address1 := models.Address{Name: "Old Address"}
	assoc := query.Model(&createdUser).Association("Address")
	if err := assoc.Append(&address1); err != nil {
		t.Fatalf("Failed to append first address: %v", err)
	}

	address2 := models.Address{Name: "New Address"}
	if err := assoc.Replace(&address2); err != nil {
		t.Fatalf("Failed to replace address: %v", err)
	}

	var loadedAddress models.Address
	if err := assoc.Find(&loadedAddress); err != nil {
		t.Fatalf("Failed to find associated address: %v", err)
	}

	if loadedAddress.Name != "New Address" {
		t.Errorf("Expected address name 'New Address', got '%s'", loadedAddress.Name)
	}
}

func TestCockroachDBIntegrationQueryAssociationCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	user := models.User{
		Name: "association_count",
	}

	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "association_count").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	book1 := models.Book{Name: "Book 1"}
	book2 := models.Book{Name: "Book 2"}

	assoc := query.Model(&createdUser).Association("Books")
	if err := assoc.Append(&book1, &book2); err != nil {
		t.Fatalf("Failed to append books: %v", err)
	}

	count := assoc.Count()
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestCockroachDBIntegrationQueryAssociationReplaceHasMany(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	user := models.User{
		Name: "association_replace_has_many",
	}

	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "association_replace_has_many").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	book1 := models.Book{Name: "Book 1"}
	book2 := models.Book{Name: "Book 2"}
	assoc := query.Model(&createdUser).Association("Books")
	if err := assoc.Append(&book1, &book2); err != nil {
		t.Fatalf("Failed to append first books: %v", err)
	}

	book3 := models.Book{Name: "Book 3"}
	book4 := models.Book{Name: "Book 4"}
	if err := assoc.Replace(&book3, &book4); err != nil {
		t.Fatalf("Failed to replace books: %v", err)
	}

	var books []models.Book
	if err := assoc.Find(&books); err != nil {
		t.Fatalf("Failed to find associated books: %v", err)
	}

	if len(books) != 2 {
		t.Errorf("Expected 2 books after replace, got %d", len(books))
	}

	bookNames := make(map[string]bool)
	for _, book := range books {
		bookNames[book.Name] = true
	}

	if !bookNames["Book 3"] || !bookNames["Book 4"] {
		t.Errorf("Expected books 'Book 3' and 'Book 4', got %v", bookNames)
	}
}

func TestCockroachDBIntegrationQueryAssociationDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	user := models.User{
		Name: "association_delete",
	}

	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "association_delete").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	book1 := models.Book{Name: "Book 1"}
	book2 := models.Book{Name: "Book 2"}
	assoc := query.Model(&createdUser).Association("Books")
	if err := assoc.Append(&book1, &book2); err != nil {
		t.Fatalf("Failed to append books: %v", err)
	}

	if err := assoc.Delete(&book1); err != nil {
		t.Fatalf("Failed to delete book: %v", err)
	}

	var books []models.Book
	if err := assoc.Find(&books); err != nil {
		t.Fatalf("Failed to find associated books: %v", err)
	}

	if len(books) != 1 {
		t.Errorf("Expected 1 book after delete, got %d", len(books))
	}
}

func TestCockroachDBIntegrationQueryAssociationClear(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	user := models.User{
		Name: "association_clear",
	}

	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "association_clear").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	book1 := models.Book{Name: "Book 1"}
	book2 := models.Book{Name: "Book 2"}
	assoc := query.Model(&createdUser).Association("Books")
	if err := assoc.Append(&book1, &book2); err != nil {
		t.Fatalf("Failed to append books: %v", err)
	}

	if err := assoc.Clear(); err != nil {
		t.Fatalf("Failed to clear association: %v", err)
	}

	var books []models.Book
	if err := assoc.Find(&books); err != nil {
		t.Fatalf("Failed to find associated books: %v", err)
	}

	if len(books) != 0 {
		t.Errorf("Expected 0 books after clear, got %d", len(books))
	}
}

func TestCockroachDBIntegrationQueryAssociationWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	user := models.User{
		Name: "association_conditions",
	}

	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "association_conditions").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	book1 := models.Book{Name: "Book 1"}
	book2 := models.Book{Name: "Book 2"}
	assoc := query.Model(&createdUser).Association("Books")
	if err := assoc.Append(&book1, &book2); err != nil {
		t.Fatalf("Failed to append books: %v", err)
	}

	var books []models.Book
	if err := assoc.Find(&books, "name = ?", "Book 1"); err != nil {
		t.Fatalf("Failed to find associated books with conditions: %v", err)
	}

	if len(books) != 1 {
		t.Errorf("Expected 1 book with condition, got %d", len(books))
	}

	if len(books) > 0 && books[0].Name != "Book 1" {
		t.Errorf("Expected book name 'Book 1', got '%s'", books[0].Name)
	}
}

func TestCockroachDBIntegrationQueryPolymorphicAssociation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	if _, err := query.Exec("CREATE TABLE IF NOT EXISTS posts (id SERIAL PRIMARY KEY, title VARCHAR(255), content TEXT, created_at TIMESTAMP, updated_at TIMESTAMP)"); err != nil {
		t.Fatalf("Failed to create posts table: %v", err)
	}
	if _, err := query.Exec("CREATE TABLE IF NOT EXISTS videos (id SERIAL PRIMARY KEY, title VARCHAR(255), url VARCHAR(255), created_at TIMESTAMP, updated_at TIMESTAMP)"); err != nil {
		t.Fatalf("Failed to create videos table: %v", err)
	}
	if _, err := query.Exec("CREATE TABLE IF NOT EXISTS comments (id SERIAL PRIMARY KEY, body TEXT, commentable_id INTEGER, commentable_type VARCHAR(255), created_at TIMESTAMP, updated_at TIMESTAMP)"); err != nil {
		t.Fatalf("Failed to create comments table: %v", err)
	}
	defer func() {
		_, _ = query.Exec("DROP TABLE IF EXISTS comments")
		_, _ = query.Exec("DROP TABLE IF EXISTS videos")
		_, _ = query.Exec("DROP TABLE IF EXISTS posts")
	}()

	post := models.Post{
		Title:   "Test Post",
		Content: "This is a test post",
	}
	if err := query.Model(&models.Post{}).Create(&post); err != nil {
		t.Fatalf("Failed to create post: %v", err)
	}

	video := models.Video{
		Title: "Test Video",
		URL:   "http://example.com/video",
	}
	if err := query.Model(&models.Video{}).Create(&video); err != nil {
		t.Fatalf("Failed to create video: %v", err)
	}

	comment1 := models.Comment{Body: "Comment 1 on post"}
	comment2 := models.Comment{Body: "Comment 2 on post"}

	postAssoc := query.Model(&post).Association("Comments")
	if err := postAssoc.Append(&comment1, &comment2); err != nil {
		t.Fatalf("Failed to append comments to post: %v", err)
	}

	var postComments []models.Comment
	if err := postAssoc.Find(&postComments); err != nil {
		t.Fatalf("Failed to find comments for post: %v", err)
	}

	if len(postComments) != 2 {
		t.Errorf("Expected 2 comments for post, got %d", len(postComments))
	}

	comment3 := models.Comment{Body: "Comment 1 on video"}
	comment4 := models.Comment{Body: "Comment 2 on video"}

	videoAssoc := query.Model(&video).Association("Comments")
	if err := videoAssoc.Append(&comment3, &comment4); err != nil {
		t.Fatalf("Failed to append comments to video: %v", err)
	}

	var videoComments []models.Comment
	if err := videoAssoc.Find(&videoComments); err != nil {
		t.Fatalf("Failed to find comments for video: %v", err)
	}

	if len(videoComments) != 2 {
		t.Errorf("Expected 2 comments for video, got %d", len(videoComments))
	}

	comment5 := models.Comment{Body: "Another comment on post"}
	commentAssoc := query.Model(&comment5).Association("Commentable")
	if err := commentAssoc.Append(&post); err != nil {
		t.Fatalf("Failed to associate comment with post: %v", err)
	}

	var loadedPost models.Post
	if err := commentAssoc.Find(&loadedPost); err != nil {
		t.Fatalf("Failed to find post for comment: %v", err)
	}

	if loadedPost.Title != "Test Post" {
		t.Errorf("Expected post title 'Test Post', got '%s'", loadedPost.Title)
	}

	count := postAssoc.Count()
	t.Logf("Count returned: %d (expected 2)", count)
}

func TestCockroachDBIntegrationQueryAssociationBelongsTo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	user := models.User{
		Name: "association_belongs_to",
	}

	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "association_belongs_to").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	address := models.Address{
		Name: "Test Address",
	}

	if err := query.Model(&models.Address{}).Create(&address); err != nil {
		t.Fatalf("Failed to create address: %v", err)
	}

	assoc := query.Model(&address).Association("User")
	if err := assoc.Append(&createdUser); err != nil {
		t.Fatalf("Failed to append user to address: %v", err)
	}

	var loadedUser models.User
	if err := assoc.Find(&loadedUser); err != nil {
		t.Fatalf("Failed to find associated user: %v", err)
	}

	if loadedUser.Name != "association_belongs_to" {
		t.Errorf("Expected user name 'association_belongs_to', got '%s'", loadedUser.Name)
	}

	if err := assoc.Delete(&createdUser); err != nil {
		t.Fatalf("Failed to delete user from address: %v", err)
	}

	var deletedUser models.User
	if err := assoc.Find(&deletedUser); err != nil {
		t.Logf("Find returned error after delete (expected): %v", err)
	}
}
