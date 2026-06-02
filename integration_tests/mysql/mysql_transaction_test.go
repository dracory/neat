package mysql_test

import (
	"database/sql"
	"errors"
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/integration_tests/models"
)

func TestMySQLTransactionCommit(t *testing.T) {

	db := SetupMySQLTest(t)

	// Test basic commit
	err := db.Transaction(func(tx contractsorm.Query) error {
		user := models.User{Name: "tx_commit_user"}
		return tx.Model(&models.User{}).Create(&user)
	})
	if err != nil {
		t.Errorf("Transaction commit failed: %v", err)
	}

	var user models.User
	err = db.Query().Model(&models.User{}).Where("name = ?", "tx_commit_user").First(&user)
	if err != nil {
		t.Errorf("Failed to find committed user: %v", err)
	}

	if user.Name != "tx_commit_user" {
		t.Errorf("Expected 'tx_commit_user', got '%s'", user.Name)
	}

	// Test commit with multiple operations
	err = db.Transaction(func(tx contractsorm.Query) error {
		user1 := models.User{Name: "tx_multi_user1"}
		if err := tx.Model(&models.User{}).Create(&user1); err != nil {
			return err
		}

		user2 := models.User{Name: "tx_multi_user2"}
		return tx.Model(&models.User{}).Create(&user2)
	})
	if err != nil {
		t.Errorf("Transaction commit with multiple operations failed: %v", err)
	}

	var count int64
	err = db.Query().Table("users").Where("name LIKE ?", "tx_multi_user%").Count(&count)
	if err != nil {
		t.Errorf("Failed to count users: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestMySQLTransactionRollback(t *testing.T) {

	db := SetupMySQLTest(t)

	// Test rollback on error
	err := db.Transaction(func(tx contractsorm.Query) error {
		user := models.User{Name: "tx_rollback_user"}
		if err := tx.Model(&models.User{}).Create(&user); err != nil {
			return err
		}
		return errors.New("some error")
	})
	if err == nil {
		t.Error("Expected error, got nil")
	} else if err.Error() != "some error" {
		t.Errorf("Expected 'some error', got '%s'", err.Error())
	}

	var count int64
	err = db.Query().Table("users").Where("name = ?", "tx_rollback_user").Count(&count)
	if err != nil {
		t.Errorf("Failed to count users: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	// Test rollback with explicit call (using Begin/Rollback)
	tx, err := db.Query().Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	if !tx.InTransaction() {
		t.Error("Expected InTransaction to be true")
	}

	user := models.User{Name: "tx_explicit_rollback"}
	err = tx.Model(&models.User{}).Create(&user)
	if err != nil {
		t.Fatalf("Failed to create user in transaction: %v", err)
	}

	err = tx.Rollback()
	if err != nil {
		t.Errorf("Failed to rollback: %v", err)
	}

	err = db.Query().Table("users").Where("name = ?", "tx_explicit_rollback").Count(&count)
	if err != nil {
		t.Errorf("Failed to count users: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}

func TestMySQLTransactionErrorHandling(t *testing.T) {

	db := SetupMySQLTest(t)

	// Test error propagation
	expectedErr := errors.New("propagation error")
	err := db.Transaction(func(tx contractsorm.Query) error {
		return expectedErr
	})
	if err != expectedErr {
		t.Errorf("Expected propagation error, got %v", err)
	}
}

func TestMySQLNestedTransactions(t *testing.T) {

	db := SetupMySQLTest(t)

	// Test nested transactions using db.Transaction
	err := db.Transaction(func(tx contractsorm.Query) error {
		user1 := models.User{Name: "tx_nested_outer"}
		if err := tx.Model(&models.User{}).Create(&user1); err != nil {
			return err
		}

		return tx.Transaction(func(innerTx contractsorm.Query) error {
			user2 := models.User{Name: "tx_nested_inner"}
			return innerTx.Model(&models.User{}).Create(&user2)
		})
	})
	if err != nil {
		t.Errorf("Nested transaction failed: %v", err)
	}

	var count int64
	err = db.Query().Table("users").Where("name LIKE ?", "tx_nested_%").Count(&count)
	if err != nil {
		t.Errorf("Failed to count users: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}

	// Test nested transaction rollback
	err = db.Transaction(func(tx contractsorm.Query) error {
		user1 := models.User{Name: "tx_nested_rollback_outer"}
		if err := tx.Model(&models.User{}).Create(&user1); err != nil {
			return err
		}

		_ = tx.Transaction(func(innerTx contractsorm.Query) error {
			user2 := models.User{Name: "tx_nested_rollback_inner"}
			_ = innerTx.Model(&models.User{}).Create(&user2)
			return errors.New("rollback inner")
		})

		return nil
	})
	if err != nil {
		t.Errorf("Nested transaction with rollback failed: %v", err)
	}

	err = db.Query().Table("users").Where("name = ?", "tx_nested_rollback_outer").Count(&count)
	if err != nil {
		t.Errorf("Failed to count users: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	err = db.Query().Table("users").Where("name = ?", "tx_nested_rollback_inner").Count(&count)
	if err != nil {
		t.Errorf("Failed to count users: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}

func TestMySQLTransactionIsolationLevels(t *testing.T) {

	db := SetupMySQLTest(t)

	levels := []sql.IsolationLevel{
		sql.LevelReadCommitted,
		sql.LevelRepeatableRead,
		sql.LevelSerializable,
	}

	for _, level := range levels {
		err := db.Transaction(func(tx contractsorm.Query) error {
			innerTx, err := db.Query().Begin(&sql.TxOptions{Isolation: level})
			if err != nil {
				return err
			}
			return innerTx.Rollback()
		})
		if err != nil {
			t.Errorf("Transaction with isolation level %v failed: %v", level, err)
		}
	}
}
