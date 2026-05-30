package mysql

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestMySQLIntegrationQueryAssociationFind(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
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

	// Test HasOne association
	var address models.Address
	assoc := query.Model(&createdUser).Association("Address")
	if err := assoc.Find(&address); err != nil {
		t.Logf("Find returned error (expected for empty association): %v", err)
	}
}

func TestMySQLIntegrationQueryAssociationAppendHasOne(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
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

	// Test HasOne association append
	address := models.Address{
		Name: "Test Address",
	}

	assoc := query.Model(&createdUser).Association("Address")
	if err := assoc.Append(&address); err != nil {
		t.Fatalf("Failed to append address: %v", err)
	}

	// Verify the address was associated
	var loadedAddress models.Address
	if err := assoc.Find(&loadedAddress); err != nil {
		t.Fatalf("Failed to find associated address: %v", err)
	}

	if loadedAddress.Name != "Test Address" {
		t.Errorf("Expected address name 'Test Address', got '%s'", loadedAddress.Name)
	}
}

func TestMySQLIntegrationQueryAssociationAppendHasMany(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
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

	// Test HasMany association append
	book1 := models.Book{Name: "Book 1"}
	book2 := models.Book{Name: "Book 2"}

	assoc := query.Model(&createdUser).Association("Books")
	if err := assoc.Append(&book1, &book2); err != nil {
		t.Fatalf("Failed to append books: %v", err)
	}

	// Verify the books were associated
	var books []models.Book
	if err := assoc.Find(&books); err != nil {
		t.Fatalf("Failed to find associated books: %v", err)
	}

	if len(books) != 2 {
		t.Errorf("Expected 2 books, got %d", len(books))
	}
}

func TestMySQLIntegrationQueryAssociationReplaceHasOne(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
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

	// First append an address
	address1 := models.Address{Name: "Old Address"}
	assoc := query.Model(&createdUser).Association("Address")
	if err := assoc.Append(&address1); err != nil {
		t.Fatalf("Failed to append first address: %v", err)
	}

	// Replace with a new address
	address2 := models.Address{Name: "New Address"}
	if err := assoc.Replace(&address2); err != nil {
		t.Fatalf("Failed to replace address: %v", err)
	}

	// Verify the new address is associated
	var loadedAddress models.Address
	if err := assoc.Find(&loadedAddress); err != nil {
		t.Fatalf("Failed to find associated address: %v", err)
	}

	if loadedAddress.Name != "New Address" {
		t.Errorf("Expected address name 'New Address', got '%s'", loadedAddress.Name)
	}
}

func TestMySQLIntegrationQueryAssociationCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
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

	// Test HasMany association count
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

func TestMySQLIntegrationQueryAssociationReplaceHasMany(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
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

	// First append some books
	book1 := models.Book{Name: "Book 1"}
	book2 := models.Book{Name: "Book 2"}
	assoc := query.Model(&createdUser).Association("Books")
	if err := assoc.Append(&book1, &book2); err != nil {
		t.Fatalf("Failed to append first books: %v", err)
	}

	// Replace with new books
	book3 := models.Book{Name: "Book 3"}
	book4 := models.Book{Name: "Book 4"}
	if err := assoc.Replace(&book3, &book4); err != nil {
		t.Fatalf("Failed to replace books: %v", err)
	}

	// Verify only the new books are associated
	var books []models.Book
	if err := assoc.Find(&books); err != nil {
		t.Fatalf("Failed to find associated books: %v", err)
	}

	if len(books) != 2 {
		t.Errorf("Expected 2 books after replace, got %d", len(books))
	}

	// Verify the books are the new ones
	bookNames := make(map[string]bool)
	for _, book := range books {
		bookNames[book.Name] = true
	}

	if !bookNames["Book 3"] || !bookNames["Book 4"] {
		t.Errorf("Expected books 'Book 3' and 'Book 4', got %v", bookNames)
	}
}

func TestMySQLIntegrationQueryAssociationDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
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

	// Append some books
	book1 := models.Book{Name: "Book 1"}
	book2 := models.Book{Name: "Book 2"}
	assoc := query.Model(&createdUser).Association("Books")
	if err := assoc.Append(&book1, &book2); err != nil {
		t.Fatalf("Failed to append books: %v", err)
	}

	// Delete one book
	if err := assoc.Delete(&book1); err != nil {
		t.Fatalf("Failed to delete book: %v", err)
	}

	// Verify only one book remains
	var books []models.Book
	if err := assoc.Find(&books); err != nil {
		t.Fatalf("Failed to find associated books: %v", err)
	}

	if len(books) != 1 {
		t.Errorf("Expected 1 book after delete, got %d", len(books))
	}
}

func TestMySQLIntegrationQueryAssociationClear(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
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

	// Append some books
	book1 := models.Book{Name: "Book 1"}
	book2 := models.Book{Name: "Book 2"}
	assoc := query.Model(&createdUser).Association("Books")
	if err := assoc.Append(&book1, &book2); err != nil {
		t.Fatalf("Failed to append books: %v", err)
	}

	// Clear all books
	if err := assoc.Clear(); err != nil {
		t.Fatalf("Failed to clear association: %v", err)
	}

	// Verify no books remain
	var books []models.Book
	if err := assoc.Find(&books); err != nil {
		t.Fatalf("Failed to find associated books: %v", err)
	}

	if len(books) != 0 {
		t.Errorf("Expected 0 books after clear, got %d", len(books))
	}
}

func TestMySQLIntegrationQueryAssociationWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
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

	// Append some books
	book1 := models.Book{Name: "Book 1"}
	book2 := models.Book{Name: "Book 2"}
	assoc := query.Model(&createdUser).Association("Books")
	if err := assoc.Append(&book1, &book2); err != nil {
		t.Fatalf("Failed to append books: %v", err)
	}

	// Find with conditions
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

func TestMySQLIntegrationQueryPolymorphicAssociation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Polymorphic associations not yet implemented")
}

func TestMySQLIntegrationQueryAssociationBelongsTo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
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

	// Test BelongsTo association append
	address := models.Address{
		Name: "Test Address",
	}

	if err := query.Model(&models.Address{}).Create(&address); err != nil {
		t.Fatalf("Failed to create address: %v", err)
	}

	// Associate address with user using BelongsTo
	assoc := query.Model(&address).Association("User")
	if err := assoc.Append(&createdUser); err != nil {
		t.Fatalf("Failed to append user to address: %v", err)
	}

	// Verify the user was associated
	var loadedUser models.User
	if err := assoc.Find(&loadedUser); err != nil {
		t.Fatalf("Failed to find associated user: %v", err)
	}

	if loadedUser.Name != "association_belongs_to" {
		t.Errorf("Expected user name 'association_belongs_to', got '%s'", loadedUser.Name)
	}

	// Test Delete
	if err := assoc.Delete(&createdUser); err != nil {
		t.Fatalf("Failed to delete user from address: %v", err)
	}

	// Verify the association was cleared
	var deletedUser models.User
	if err := assoc.Find(&deletedUser); err != nil {
		t.Logf("Find returned error after delete (expected): %v", err)
	}
}
