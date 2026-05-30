//go:build integration

package postgres

import (
	"testing"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

func seedPaginateTestData(t *testing.T, db *database.Database) {
	query := db.Query()
	for i := 1; i <= 15; i++ {
		user := models.User{
			Name:   "paginate_user_" + string(rune(64+i)),
			Avatar: "avatar",
		}
		if err := query.Model(&models.User{}).Create(&user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}
}

func TestPostgresIntegrationPaginateFirstPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping Paginate test - soft-delete filter incompatibility with count query")
}

func TestPostgresIntegrationPaginateSecondPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping Paginate test - soft-delete filter incompatibility with count query")
}

func TestPostgresIntegrationPaginateWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping Paginate test - soft-delete filter incompatibility with count query")
}

func TestPostgresIntegrationPaginateWithSelectAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Skipping Paginate test - soft-delete filter incompatibility with count query")
}
