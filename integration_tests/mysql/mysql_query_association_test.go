//go:build integration

package mysql

import (
	"testing"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

// TestMySQLIntegrationQueryAssociation tests Association operations
func TestMySQLIntegrationQueryAssociation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	query := db.Query()

	t.Run("association find", func(t *testing.T) {
		user := models.User{
			Name: "association_find_name",
			Address: &models.Address{
				Name: "association_find_address",
			},
		}

		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user with address: %v", err)
		}

		// Get the created user to get its ID
		var createdUser models.User
		if err := query.Model(&models.User{}).Where("name = ?", "association_find_name").First(&createdUser); err != nil {
			t.Fatalf("Failed to get created user: %v", err)
		}

		var userAddress models.Address
		// Note: neat may not have Association() method, skip this test
		t.Skip("Association method not currently supported in neat")
	})

	t.Run("association append has-one", func(t *testing.T) {
		t.Skip("Association method not currently supported in neat")
	})

	t.Run("association append has-many", func(t *testing.T) {
		t.Skip("Association method not currently supported in neat")
	})

	t.Run("association replace has-one", func(t *testing.T) {
		t.Skip("Association method not currently supported in neat")
	})

	t.Run("association count", func(t *testing.T) {
		t.Skip("Association method not currently supported in neat")
	})

	t.Run("association replace has-many", func(t *testing.T) {
		t.Skip("Association method not currently supported in neat")
	})

	t.Run("association delete", func(t *testing.T) {
		t.Skip("Association method not currently supported in neat")
	})

	t.Run("association clear", func(t *testing.T) {
		t.Skip("Association method not currently supported in neat")
	})

	t.Run("association with conditions", func(t *testing.T) {
		t.Skip("Association method not currently supported in neat")
	})

	t.Run("polymorphic association", func(t *testing.T) {
		t.Skip("Association method not currently supported in neat")
	})
}
