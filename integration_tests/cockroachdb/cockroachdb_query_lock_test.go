//go:build integration

package cockroachdb_test

import (
	"sync"
	"testing"
	"time"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/integration_tests/models"
)

func TestCockroachDBLockForUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)

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
		t.Errorf("Transaction failed: %v", err)
	}
}

func TestCockroachDBSharedLock(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)

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

func TestCockroachDBConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)

	user := models.User{Name: "concurrent_user"}
	if err := db.Query().Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	userID := user.ID

	var wg sync.WaitGroup
	wg.Add(2)

	start := make(chan struct{})

	go func() {
		defer wg.Done()
		<-start

		err := db.Transaction(func(tx contractsorm.Query) error {
			var result models.User
			err := tx.Model(&models.User{}).LockForUpdate().Where("id = ?", userID).First(&result)
			if err != nil {
				return err
			}

			time.Sleep(200 * time.Millisecond)

			_, err = tx.Model(&models.User{}).Where("id = ?", userID).Update("name", "updated_by_tx1")
			return err
		})
		if err != nil {
			t.Errorf("Transaction 1 failed: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		<-start

		time.Sleep(50 * time.Millisecond)

		err := db.Transaction(func(tx contractsorm.Query) error {
			var result models.User
			err := tx.Model(&models.User{}).LockForUpdate().Where("id = ?", userID).First(&result)
			if err != nil {
				return err
			}

			if result.Name != "updated_by_tx1" {
				t.Errorf("Expected 'updated_by_tx1', got '%s'", result.Name)
			}
			_, err = tx.Model(&models.User{}).Where("id = ?", userID).Update("name", "updated_by_tx2")
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
