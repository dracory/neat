package postgres

import (
	"fmt"
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgresIntegrationQueryChunkBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	for i := 1; i <= 10; i++ {
		user := models.User{Name: fmt.Sprintf("chunk_user_%d", i)}
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user %d: %v", i, err)
		}
	}

	var totalCount int
	err := query.Model(&models.User{}).Where("name LIKE ?", "chunk_user_%").OrderBy("id", "asc").Chunk(3, func(users []models.User) error {
		totalCount += len(users)
		if totalCount <= 3 {
			if len(users) != 3 {
				t.Errorf("Expected 3 users, got %d", len(users))
			}
		} else if totalCount <= 6 {
			if len(users) != 3 {
				t.Errorf("Expected 3 users, got %d", len(users))
			}
		} else if totalCount <= 9 {
			if len(users) != 3 {
				t.Errorf("Expected 3 users, got %d", len(users))
			}
		} else {
			if len(users) != 1 {
				t.Errorf("Expected 1 user, got %d", len(users))
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

func TestPostgresIntegrationQueryChunkCustomBatchSize(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	for i := 1; i <= 10; i++ {
		user := models.User{Name: fmt.Sprintf("chunk_user_%d", i)}
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user %d: %v", i, err)
		}
	}

	var totalCount int
	err := query.Model(&models.User{}).Where("name LIKE ?", "chunk_user_%").Chunk(5, func(users []models.User) error {
		totalCount += len(users)
		if len(users) != 5 {
			t.Errorf("Expected 5 users, got %d", len(users))
		}
		return nil
	})

	if err != nil {
		t.Errorf("Chunk with custom batch size failed: %v", err)
	}
	if totalCount != 10 {
		t.Errorf("Expected total count 10, got %d", totalCount)
	}
}

func TestPostgresIntegrationQueryChunkErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	for i := 1; i <= 10; i++ {
		user := models.User{Name: fmt.Sprintf("chunk_user_%d", i)}
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user %d: %v", i, err)
		}
	}

	var totalCount int
	err := query.Model(&models.User{}).Where("name LIKE ?", "chunk_user_%").OrderBy("id", "asc").Chunk(3, func(users []models.User) error {
		totalCount += len(users)
		if totalCount >= 6 {
			return fmt.Errorf("stop chunking")
		}
		return nil
	})

	if err == nil {
		t.Error("Expected error from chunk callback")
	} else if err.Error() != "stop chunking" {
		t.Errorf("Expected error 'stop chunking', got '%s'", err.Error())
	}
	if totalCount != 6 {
		t.Errorf("Expected total count 6, got %d", totalCount)
	}
}
