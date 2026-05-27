

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

	// Create a test user
	user := models.User{Name: "lock_user"}
	if err := db.Query().Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// SQLite doesn't support FOR UPDATE, but the ORM should handle it gracefully
	// by ignoring the lock clause for SQLite dialect
	var result models.User
	err := db.Query().Model(&models.User{}).LockForUpdate().Where("name = ?", "lock_user").First(&result)
	if err != nil {
		t.Errorf("LockForUpdate query failed: %v", err)
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

	// Create a test user
	user := models.User{Name: "shared_lock_user"}
	if err := db.Query().Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// SQLite doesn't support LOCK IN SHARE MODE, but the ORM should handle it gracefully
	// by ignoring the lock clause for SQLite dialect
	var result models.User
	err := db.Query().Model(&models.User{}).SharedLock().Where("name = ?", "shared_lock_user").First(&result)
	if err != nil {
		t.Errorf("SharedLock query failed: %v", err)
	}
	if result.Name != "shared_lock_user" {
		t.Errorf("Expected 'shared_lock_user', got '%s'", result.Name)
	}
}
