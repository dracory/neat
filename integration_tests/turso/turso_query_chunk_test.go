package turso

import (
	"fmt"
	"testing"

	"github.com/dracory/neat/integration_tests/common"
	"github.com/dracory/neat/integration_tests/models"
)

func TestTursoIntegrationQueryChunkBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.SeedChunkTestData(t, db)

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
}

func TestTursoIntegrationQueryChunkCustomBatchSize(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.SeedChunkTestData(t, db)

	batchSizes := []int{0, 0, 0, 0} // 3, 3, 3, 1
	iteration := 0
	err := db.Query().Model(&models.User{}).Where("name LIKE ?", "chunk_user_%").OrderBy("id", "asc").
		Chunk(3, func(users []models.User) error {
			batchSizes[iteration] = len(users)
			iteration++
			return nil
		})

	if err != nil {
		t.Errorf("chunk custom batch size failed: %v", err)
	}
	expected := []int{3, 3, 3, 1}
	for i, v := range expected {
		if batchSizes[i] != v {
			t.Errorf("Expected batch size %d at iteration %d, got %d", v, i, batchSizes[i])
		}
	}
}

func TestTursoIntegrationQueryChunkErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.SeedChunkTestData(t, db)

	count := 0
	expectedErr := fmt.Errorf("stop chunking")
	err := db.Query().Model(&models.User{}).Where("name LIKE ?", "chunk_user_%").
		Chunk(3, func(users []models.User) error {
			count += len(users)
			return expectedErr
		})

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	if count != 3 {
		t.Errorf("Expected count 3 (stopped after first chunk), got %d", count)
	}
}

func TestTursoIntegrationQueryChunkEmptyResultSet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)

	called := false
	err := db.Query().Model(&models.User{}).Where("name = ?", "non_existent").
		Chunk(3, func(users []models.User) error {
			called = true
			return nil
		})

	if err != nil {
		t.Errorf("chunk empty result set failed: %v", err)
	}
	if called {
		t.Errorf("Callback should not have been called for empty result set")
	}
}

func TestTursoIntegrationQueryChunkPointerSlice(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	common.SeedChunkTestData(t, db)

	count := 0
	err := db.Query().Model(&models.User{}).Where("name LIKE ?", "chunk_user_%").
		Chunk(3, func(users []*models.User) error {
			count += len(users)
			for _, u := range users {
				if u.ID == 0 {
					t.Errorf("User ID should not be zero")
				}
				if u.Name == "" {
					t.Errorf("User Name should not be empty")
				}
			}
			return nil
		})

	if err != nil {
		t.Errorf("chunk pointer slice failed: %v", err)
	}
	if count != 10 {
		t.Errorf("Expected 10 total users, got %d", count)
	}
}
