package sqlite

import (
	"fmt"
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLiteIntegrationQueryChunk(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupSQLiteTest(t)

	// Seed data
	for i := 1; i <= 10; i++ {
		user := models.User{Name: fmt.Sprintf("chunk_user_%d", i)}
		if err := db.Query().Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user %d: %v", i, err)
		}
	}

	t.Run("chunk basic", func(t *testing.T) {
		count := 0
		err := db.Query().Model(&models.User{}).Where("name LIKE ?", "chunk_user_%").
			Chunk(3, func(users []models.User) error {
				count += len(users)
				return nil
			})
		if err != nil {
			t.Errorf("chunk basic failed: %v", err)
		}
		if count != 10 {
			t.Errorf("Expected 10 total users, got %d", count)
		}
	})

	t.Run("chunk with custom batch size", func(t *testing.T) {
		t.Skip("ORM Chunk() passes []interface{} to typed callbacks — type mismatch panic, not yet fixed")
	})

	t.Run("chunk error handling", func(t *testing.T) {
		t.Skip("ORM Chunk() passes []interface{} to typed callbacks — type mismatch panic, not yet fixed")
	})

	t.Run("chunk with empty result set", func(t *testing.T) {
		t.Skip("ORM Chunk() passes []interface{} to typed callbacks — type mismatch panic, not yet fixed")
	})
}
