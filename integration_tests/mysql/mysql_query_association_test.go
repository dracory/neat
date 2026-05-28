//go:build integration

package mysql

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestMySQLIntegrationQueryAssociationFind(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Association method not currently supported in neat")

	db := SetupMySQLTest(t)
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
}

func TestMySQLIntegrationQueryAssociationAppendHasOne(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Association method not currently supported in neat")
}

func TestMySQLIntegrationQueryAssociationAppendHasMany(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Association method not currently supported in neat")
}

func TestMySQLIntegrationQueryAssociationReplaceHasOne(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Association method not currently supported in neat")
}

func TestMySQLIntegrationQueryAssociationCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Association method not currently supported in neat")
}

func TestMySQLIntegrationQueryAssociationReplaceHasMany(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Association method not currently supported in neat")
}

func TestMySQLIntegrationQueryAssociationDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Association method not currently supported in neat")
}

func TestMySQLIntegrationQueryAssociationClear(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Association method not currently supported in neat")
}

func TestMySQLIntegrationQueryAssociationWithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Association method not currently supported in neat")
}

func TestMySQLIntegrationQueryPolymorphicAssociation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Skip("Association method not currently supported in neat")
}
