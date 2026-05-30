
package postgres

import (
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgreSQLIntegrationQueryBelongsToWith(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Create user first
	user := &models.User{
		Name: "belongs_to_name",
	}
	if err := query.Model(&models.User{}).Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create address with user_id
	address := &models.Address{
		Name:   "belongs_to_address",
		UserID: user.ID,
	}
	if err := query.Model(&models.Address{}).Create(address); err != nil {
		t.Fatalf("Failed to create address: %v", err)
	}

	var userAddress models.Address
	if err := query.Model(&models.Address{}).With("User").Where("name = ?", "belongs_to_address").First(&userAddress); err != nil {
		t.Errorf("Failed to find address with user: %v", err)
	}
	if userAddress.ID == 0 {
		t.Error("Address ID should be set")
	}
	if userAddress.User == nil {
		t.Error("User should be loaded")
	}
	if userAddress.User.ID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, userAddress.User.ID)
	}
}

func TestPostgreSQLIntegrationQueryBelongsToWithout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Create user first
	user := &models.User{
		Name: "belongs_to_without_user",
	}
	if err := query.Model(&models.User{}).Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create address with user_id
	address := &models.Address{
		Name:   "belongs_to_without_address",
		UserID: user.ID,
	}
	if err := query.Model(&models.Address{}).Create(address); err != nil {
		t.Fatalf("Failed to create address: %v", err)
	}

	var userAddress models.Address
	err := query.Model(&models.Address{}).With("User").Without("User").Where("name = ?", "belongs_to_without_address").First(&userAddress)
	if err != nil {
		t.Errorf("Belongs to without failed: %v", err)
	}
	if userAddress.User != nil {
		t.Error("User should be nil")
	}
}

func TestPostgreSQLIntegrationQueryBelongsToWithConstraints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	// Create user first
	user := &models.User{
		Name: "constrained_user",
	}
	if err := query.Model(&models.User{}).Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create address with user_id
	address := &models.Address{
		Name:   "constrained_address",
		UserID: user.ID,
	}
	if err := query.Model(&models.Address{}).Create(address); err != nil {
		t.Fatalf("Failed to create address: %v", err)
	}

	var userAddress models.Address
	err := query.Model(&models.Address{}).With("User", func(q contractsorm.Query) contractsorm.Query {
		return q.Where("name = ?", "non_existent_user")
	}).Where("name = ?", "constrained_address").First(&userAddress)

	if err != nil {
		t.Errorf("Belongs to with constraints failed: %v", err)
	}
	if userAddress.User != nil {
		t.Error("User should be nil with non-existent constraint")
	}

	err = query.Model(&models.Address{}).With("User", func(q contractsorm.Query) contractsorm.Query {
		return q.Where("name = ?", "constrained_user")
	}).Where("name = ?", "constrained_address").First(&userAddress)

	if err != nil {
		t.Errorf("Belongs to with valid constraint failed: %v", err)
	}
	if userAddress.User == nil {
		t.Error("User should be loaded with valid constraint")
	}
	if userAddress.User.Name != "constrained_user" {
		t.Errorf("Expected 'constrained_user', got '%s'", userAddress.User.Name)
	}
}

func TestPostgreSQLIntegrationQueryMultipleBelongsTo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	user := &models.User{Name: "multi_belongs_user"}
	if err := query.Model(&models.User{}).Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	book := &models.Book{
		Name:   "multi_belongs_book",
		UserID: user.ID,
	}
	if err := query.Model(&models.Book{}).Create(book); err != nil {
		t.Fatalf("Failed to create book: %v", err)
	}

	var foundBook models.Book
	err := query.Model(&models.Book{}).With("User").Where("id = ?", book.ID).First(&foundBook)
	if err != nil {
		t.Errorf("Failed to find book with user: %v", err)
	}
	if foundBook.User == nil {
		t.Error("User should be loaded")
	}
	if foundBook.User.Name != user.Name {
		t.Errorf("Expected '%s', got '%s'", user.Name, foundBook.User.Name)
	}

	var foundAddress models.Address
	address := &models.Address{Name: "multi_belongs_address", UserID: user.ID}
	if err := query.Model(&models.Address{}).Create(address); err != nil {
		t.Fatalf("Failed to create address: %v", err)
	}

	err = query.Model(&models.Address{}).With("User").Where("id = ?", address.ID).First(&foundAddress)
	if err != nil {
		t.Errorf("Failed to find address with user: %v", err)
	}
	if foundAddress.User == nil {
		t.Error("User should be loaded")
	}
	if foundAddress.User.Name != user.Name {
		t.Errorf("Expected '%s', got '%s'", user.Name, foundAddress.User.Name)
	}
}
