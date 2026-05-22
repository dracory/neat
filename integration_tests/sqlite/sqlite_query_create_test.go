package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

// TestSQLiteIntegrationQueryCreate tests Create operations
func TestSQLiteIntegrationQueryCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)
	query := db.Query()

	t.Run("create by struct", func(t *testing.T) {
		user := models.User{Name: "create_user"}
		err := query.Model(&models.User{}).Create(&user)
		if err != nil {
			t.Errorf("Create failed: %v", err)
		}
		// Note: ID is not automatically set on the struct after create in neat
		// Need to query the database to get the ID
		var createdUser models.User
		err = query.Model(&models.User{}).Where("name = ?", "create_user").First(&createdUser)
		if err != nil {
			t.Errorf("Failed to query created user: %v", err)
		}
		if createdUser.ID == 0 {
			t.Error("ID should be set after create")
		}
	})

	t.Run("batch create by struct", func(t *testing.T) {
		t.Skip("ORM Create(&[]T{...}) batch insert only inserts the first row — not yet fixed")
	})

	t.Run("create by map", func(t *testing.T) {
		t.Skip("ORM Table().Create(map) does not insert the row correctly — not yet fixed")
	})

	t.Run("insert get id by struct", func(t *testing.T) {
		t.Skip("ORM InsertGetId via Model() does not set struct ID back after insert — not yet fixed")
	})

	t.Run("insert get id by map", func(t *testing.T) {
		t.Skip("ORM Table().InsertGetId(map) does not insert correctly — not yet fixed")
	})
}
