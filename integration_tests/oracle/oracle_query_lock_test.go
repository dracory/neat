package oracle_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/integration_tests/models"
)

func TestOracleLockForUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)

	user := models.User{Name: "lock_user"}
	if err := db.Query().Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	err := db.Transaction(func(tx contractsorm.Query) error {
		var result models.User
		err := tx.Model(&models.User{}).LockForUpdate().Where("name = ?", "lock_user").First(&result)
		if err != nil {
			return err
		}
		if result.Name != "lock_user" {
			t.Errorf("Expected 'lock_user', got '%s'", result.Name)
		}
		return nil
	})
	if err != nil {
		t.Errorf("LockForUpdate transaction failed: %v", err)
	}
}

func TestOracleSharedLock(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)

	user := models.User{Name: "shared_lock_user"}
	if err := db.Query().Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	err := db.Transaction(func(tx contractsorm.Query) error {
		var result models.User
		err := tx.Model(&models.User{}).SharedLock().Where("name = ?", "shared_lock_user").First(&result)
		if err != nil {
			return err
		}
		if result.Name != "shared_lock_user" {
			t.Errorf("Expected 'shared_lock_user', got '%s'", result.Name)
		}
		return nil
	})
	if err != nil {
		t.Errorf("SharedLock transaction failed: %v", err)
	}
}

func TestOracleConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupOracleTest(t)

	user := models.User{Name: "concurrent_user"}
	if err := db.Query().Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	userID := user.ID // Store the ID for use in WHERE clauses

	var wg sync.WaitGroup
	wg.Add(2)

	start := make(chan struct{})

	// Goroutine 1: Lock row and sleep
	go func() {
		defer wg.Done()
		<-start

		err := db.Transaction(func(tx contractsorm.Query) error {
			var result []models.User
			// Use WithTrashed to bypass soft delete filter so FOR UPDATE works on Oracle
			// Use Get() instead of First() to avoid LIMIT which Oracle doesn't support with FOR UPDATE
			err := tx.Model(&models.User{}).WithTrashed().LockForUpdate().Where("id = ?", userID).Get(&result)
			if err != nil {
				return err
			}
			if len(result) == 0 {
				return fmt.Errorf("no user found with id %d", userID)
			}

			// Hold the lock for a short duration
			time.Sleep(200 * time.Millisecond)

			_, err = tx.Model(&models.User{}).WithTrashed().Where("id = ?", userID).Update("name", "updated_by_tx1")
			return err
		})
		if err != nil {
			t.Errorf("Transaction 1 failed: %v", err)
		}
	}()

	// Goroutine 2: Try to lock the same row
	go func() {
		defer wg.Done()
		<-start

		// Wait a bit to ensure Goroutine 1 starts first
		time.Sleep(50 * time.Millisecond)

		err := db.Transaction(func(tx contractsorm.Query) error {
			var result []models.User
			// This should block until Goroutine 1 commits
			// Use WithTrashed to bypass soft delete filter so FOR UPDATE works on Oracle
			// Use Get() instead of First() to avoid LIMIT which Oracle doesn't support with FOR UPDATE
			err := tx.Model(&models.User{}).WithTrashed().LockForUpdate().Where("id = ?", userID).Get(&result)
			if err != nil {
				return err
			}
			if len(result) == 0 {
				return fmt.Errorf("no user found with id %d", userID)
			}
			if result[0].Name != "updated_by_tx1" {
				t.Errorf("Expected 'updated_by_tx1', got '%s'", result[0].Name)
			}
			_, err = tx.Model(&models.User{}).WithTrashed().Where("id = ?", userID).Update("name", "updated_by_tx2")
			return err
		})
		if err != nil {
			t.Errorf("Transaction 2 failed: %v", err)
		}
	}()

	close(start)
	wg.Wait()

	var finalResult models.User
	err := db.Query().Model(&models.User{}).Where("id = ?", userID).First(&finalResult)
	if err != nil {
		t.Errorf("Failed to find final result: %v", err)
	}

	if finalResult.Name != "updated_by_tx2" {
		t.Errorf("Expected 'updated_by_tx2', got '%s'", finalResult.Name)
	}
}
