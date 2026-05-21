//go:build integration

package sqlite

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteLockForUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	user := models.User{Name: "lock_user"}
	if err := db.Query().Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Verify that LockForUpdate can be called without errors on SQLite
	// SQLite doesn't support SELECT ... FOR UPDATE, but the API should be available
	var result models.User
	err := db.Query().LockForUpdate().Where("name = ?", "lock_user").First(&result)
	if err != nil {
		t.Errorf("LockForUpdate failed: %v", err)
	}
	if result.Name != "lock_user" {
		t.Errorf("Expected 'lock_user', got '%s'", result.Name)
	}
}

func TestSQLiteSharedLock(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	user := models.User{Name: "shared_lock_user"}
	if err := db.Query().Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Verify that SharedLock can be called without errors on SQLite
	var result models.User
	err := db.Query().SharedLock().Where("name = ?", "shared_lock_user").First(&result)
	if err != nil {
		t.Errorf("SharedLock failed: %v", err)
	}
	if result.Name != "shared_lock_user" {
		t.Errorf("Expected 'shared_lock_user', got '%s'", result.Name)
	}
}
