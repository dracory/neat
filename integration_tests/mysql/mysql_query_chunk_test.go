package mysql

import (
	"fmt"
	"testing"
	"time"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

func seedChunkTestData(t *testing.T, db *database.Database) {
	query := db.Query()
	now := time.Now()

	for i := 1; i <= 10; i++ {
		user := models.User{Name: fmt.Sprintf("chunk_user_%d", i), CreatedAt: now, UpdatedAt: now}
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user %d: %v", i, err)
		}
	}
}

func TestMySQLIntegrationQueryChunkBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	if db == nil {
		t.Skip("MySQL not available")
	}
	_, _ = db.Query().Table("users").Where("name LIKE ?", "chunk_user_%").Delete()
	query := db.Query()
	seedChunkTestData(t, db)

	var totalCount int
	err := query.Model(&models.User{}).Where("name LIKE ?", "chunk_user_%").OrderBy("id", "asc").Chunk(3, func(users []models.User) error {
		totalCount += len(users)
		if totalCount <= 3 {
			if len(users) != 3 {
				t.Errorf("Expected 3 users in first chunk, got %d", len(users))
			}
		} else if totalCount <= 6 {
			if len(users) != 3 {
				t.Errorf("Expected 3 users in second chunk, got %d", len(users))
			}
		} else if totalCount <= 9 {
			if len(users) != 3 {
				t.Errorf("Expected 3 users in third chunk, got %d", len(users))
			}
		} else {
			if len(users) != 1 {
				t.Errorf("Expected 1 user in last chunk, got %d", len(users))
			}
		}
		return nil
	})

	if err != nil {
		t.Errorf("Chunk failed: %v", err)
	}
	if totalCount != 10 {
		t.Errorf("Expected total count 10, got %d", totalCount)
	}
}

func TestMySQLIntegrationQueryChunkCustomBatchSize(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	if db == nil {
		t.Skip("MySQL not available")
	}
	_, _ = db.Query().Table("users").Where("name LIKE ?", "chunk_user_%").Delete()
	query := db.Query()
	seedChunkTestData(t, db)

	var totalCount int
	err := query.Model(&models.User{}).Where("name LIKE ?", "chunk_user_%").Chunk(5, func(users []models.User) error {
		totalCount += len(users)
		if len(users) != 5 {
			t.Errorf("Expected 5 users per chunk, got %d", len(users))
		}
		return nil
	})

	if err != nil {
		t.Errorf("Chunk failed: %v", err)
	}
	if totalCount != 10 {
		t.Errorf("Expected total count 10, got %d", totalCount)
	}
}

func TestMySQLIntegrationQueryChunkErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupMySQLTest(t)
	if db == nil {
		t.Skip("MySQL not available")
	}
	query := db.Query()
	seedChunkTestData(t, db)

	var totalCount int
	err := query.Model(&models.User{}).Where("name LIKE ?", "chunk_user_%").OrderBy("id", "asc").Chunk(3, func(users []models.User) error {
		totalCount += len(users)
		if totalCount >= 6 {
			return fmt.Errorf("stop chunking")
		}
		return nil
	})

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err != nil && err.Error() != "stop chunking" {
		t.Errorf("Expected 'stop chunking' error, got '%s'", err.Error())
	}
	if totalCount != 6 {
		t.Errorf("Expected total count 6, got %d", totalCount)
	}
}
