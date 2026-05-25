//go:build disabled

package postgres

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestPostgreSQLIntegrationQueryAssociationFind(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupPostgresTest(t)
	query := db.Query()

	user := models.User{
		Name: "association_find_name",
	}

	if err := query.Model(&models.User{}).Create(&user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var createdUser models.User
	if err := query.Model(&models.User{}).Where("name = ?", "association_find_name").First(&createdUser); err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	t.Skip("Association method not currently supported in neat")
}

func TestPostgreSQLIntegrationQueryAssociationAppendHasOne(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Association method not currently supported in neat")
}

func TestPostgreSQLIntegrationQueryAssociationAppendHasMany(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Association method not currently supported in neat")
}

func TestPostgreSQLIntegrationQueryAssociationReplaceHasOne(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Association method not currently supported in neat")
}

func TestPostgreSQLIntegrationQueryAssociationCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Association method not currently supported in neat")
}

func TestPostgreSQLIntegrationQueryAssociationReplaceHasMany(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Association method not currently supported in neat")
}

func TestPostgreSQLIntegrationQueryAssociationDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Association method not currently supported in neat")
}

func TestPostgreSQLIntegrationQueryAssociationClear(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Association method not currently supported in neat")
}

func TestPostgreSQLIntegrationQueryAssociationWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Association method not currently supported in neat")
}

func TestPostgreSQLIntegrationQueryPolymorphicAssociation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Association method not currently supported in neat")
}
