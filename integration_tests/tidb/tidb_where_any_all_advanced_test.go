//go:build integration

package tidb_test

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestTiDBIntegrationWhereAnyAdvanced(t *testing.T) {

	db := SetupTiDBTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "test1", Avatar: "other"},
		{Name: "other", Avatar: "test1"},
		{Name: "test2", Avatar: "other"},
		{Name: "other", Avatar: "test2"},
		{Name: "test1", Avatar: "test1"},
		{Name: "test1", Avatar: "test2"},
	}
	err := query.Model(&models.User{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var foundUsers []models.User
	err = query.Model(&models.User{}).
		WhereAny([]string{"name", "avatar"}, "=", "test1").
		WhereAny([]string{"name", "avatar"}, "=", "test2").
		Find(&foundUsers)

	if err != nil {
		t.Fatalf("WhereAny advanced failed: %v", err)
	}
	if len(foundUsers) != 1 {
		t.Errorf("Expected 1 user for multiple WhereAny, got %d", len(foundUsers))
	} else if foundUsers[0].Name != "test1" || foundUsers[0].Avatar != "test2" {
		t.Errorf("Expected user test1/test2, got %s/%s", foundUsers[0].Name, foundUsers[0].Avatar)
	}

	var likeUsers []models.User
	err = query.Model(&models.User{}).WhereAny([]string{"name", "avatar"}, "LIKE", "test%").Find(&likeUsers)
	if err != nil {
		t.Fatalf("WhereAny LIKE failed: %v", err)
	}
	if len(likeUsers) != 6 {
		t.Errorf("Expected 6 users for WhereAny LIKE, got %d", len(likeUsers))
	}

	err = query.Model(&models.User{}).WhereAny([]string{"name"}, "INVALID", "test1").Find(&foundUsers)
	if err == nil {
		t.Error("Expected error for invalid operator in WhereAny, got nil")
	}
}

func TestTiDBIntegrationWhereAllAdvanced(t *testing.T) {

	db := SetupTiDBTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "test", Avatar: "test"},
		{Name: "test", Avatar: "test"},
		{Name: "test", Avatar: "other"},
	}
	err := query.Model(&models.User{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var foundUsers []models.User
	err = query.Model(&models.User{}).
		WhereAll([]string{"name"}, "=", "test").
		WhereAll([]string{"avatar"}, "=", "test").
		Find(&foundUsers)

	if err != nil {
		t.Fatalf("WhereAll advanced failed: %v", err)
	}
	if len(foundUsers) != 2 {
		t.Errorf("Expected 2 users for multiple WhereAll, got %d", len(foundUsers))
	}

	var likeUsers []models.User
	err = query.Model(&models.User{}).WhereAll([]string{"name", "avatar"}, "LIKE", "tes%").Find(&likeUsers)
	if err != nil {
		t.Fatalf("WhereAll LIKE failed: %v", err)
	}
	if len(likeUsers) != 2 {
		t.Errorf("Expected 2 users for WhereAll LIKE, got %d", len(likeUsers))
	}
}

func TestTiDBIntegrationWhereNone(t *testing.T) {

	db := SetupTiDBTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "test", Avatar: "other"},
		{Name: "other", Avatar: "test"},
		{Name: "test", Avatar: "test"},
		{Name: "none", Avatar: "none"},
	}
	err := query.Model(&models.User{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var noneUsers []models.User
	err = query.Model(&models.User{}).WhereNone([]string{"name", "avatar"}, "=", "test").Find(&noneUsers)
	if err != nil {
		t.Fatalf("WhereNone failed: %v", err)
	}
	if len(noneUsers) != 1 || noneUsers[0].Name != "none" {
		t.Errorf("Expected 'none' user for WhereNone, got %v", noneUsers)
	}

	var noneLikeUsers []models.User
	err = query.Model(&models.User{}).WhereNone([]string{"name", "avatar"}, "LIKE", "test%").Find(&noneLikeUsers)
	if err != nil {
		t.Fatalf("WhereNone LIKE failed: %v", err)
	}
	if len(noneLikeUsers) != 1 || noneLikeUsers[0].Name != "none" {
		t.Errorf("Expected 'none' user for WhereNone LIKE, got %v", noneLikeUsers)
	}

	var combinedUsers []models.User
	err = query.Model(&models.User{}).Where("name = ?", "none").WhereNone([]string{"avatar"}, "=", "test").Find(&combinedUsers)
	if err != nil {
		t.Fatalf("WhereNone combined with Where failed: %v", err)
	}
	if len(combinedUsers) != 1 || combinedUsers[0].Name != "none" {
		t.Errorf("Expected 'none' user for combined WhereNone, got %v", combinedUsers)
	}
}

func TestTiDBIntegrationWhereNoneAdvanced(t *testing.T) {

	db := SetupTiDBTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "test1", Avatar: "test1"},
		{Name: "test1", Avatar: "test2"},
		{Name: "test2", Avatar: "test1"},
		{Name: "test2", Avatar: "test2"},
		{Name: "none", Avatar: "none"},
	}
	err := query.Model(&models.User{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var foundUsers []models.User
	err = query.Model(&models.User{}).
		WhereNone([]string{"name", "avatar"}, "=", "test1").
		WhereNone([]string{"name", "avatar"}, "=", "test2").
		Find(&foundUsers)

	if err != nil {
		t.Fatalf("WhereNone advanced failed: %v", err)
	}
	if len(foundUsers) != 1 || foundUsers[0].Name != "none" {
		t.Errorf("Expected 'none' user for multiple WhereNone, got %v", foundUsers)
	}
}

func TestTiDBIntegrationWhereEdgeCases(t *testing.T) {

	db := SetupTiDBTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "test1", Avatar: "test1"},
	}
	err := query.Model(&models.User{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}
}

func TestTiDBIntegrationWhereAnyEmptyColumns(t *testing.T) {

	db := SetupTiDBTest(t)
	var found []models.User
	err := db.Query().Model(&models.User{}).WhereAny([]string{}, "=", "test1").Find(&found)
	if err != nil {
		t.Errorf("WhereAny with empty columns failed: %v", err)
	}
}

func TestTiDBIntegrationWhereAnyInvalidOperator(t *testing.T) {

	db := SetupTiDBTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "test1", Avatar: "test1"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var found []models.User
	err := query.Model(&models.User{}).WhereAny([]string{"name"}, "INVALID", "test1").Find(&found)
	if err == nil {
		t.Error("Expected error for invalid operator in WhereAny, got nil")
	}
}

func TestTiDBIntegrationWhereAnyInvalidColumn(t *testing.T) {

	db := SetupTiDBTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "test1", Avatar: "test1"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var found []models.User
	err := query.Model(&models.User{}).WhereAny([]string{"invalid column"}, "=", "test1").Find(&found)
	if err == nil {
		t.Error("Expected error for invalid column in WhereAny, got nil")
	}
}

func TestTiDBIntegrationWhereAllEmptyColumns(t *testing.T) {

	db := SetupTiDBTest(t)
	var found []models.User
	err := db.Query().Model(&models.User{}).WhereAll([]string{}, "=", "test1").Find(&found)
	if err != nil {
		t.Errorf("WhereAll with empty columns failed: %v", err)
	}
}

func TestTiDBIntegrationWhereAllInvalidOperator(t *testing.T) {

	db := SetupTiDBTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "test1", Avatar: "test1"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var found []models.User
	err := query.Model(&models.User{}).WhereAll([]string{"name"}, "INVALID", "test1").Find(&found)
	if err == nil {
		t.Error("Expected error for invalid operator in WhereAll, got nil")
	}
}

func TestTiDBIntegrationWhereAllInvalidColumn(t *testing.T) {

	db := SetupTiDBTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "test1", Avatar: "test1"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var found []models.User
	err := query.Model(&models.User{}).WhereAll([]string{"invalid column"}, "=", "test1").Find(&found)
	if err == nil {
		t.Error("Expected error for invalid column in WhereAll, got nil")
	}
}

func TestTiDBIntegrationWhereNoneEmptyColumns(t *testing.T) {

	db := SetupTiDBTest(t)
	var found []models.User
	err := db.Query().Model(&models.User{}).WhereNone([]string{}, "=", "test1").Find(&found)
	if err != nil {
		t.Errorf("WhereNone with empty columns failed: %v", err)
	}
}

func TestTiDBIntegrationWhereNoneInvalidOperator(t *testing.T) {

	db := SetupTiDBTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "test1", Avatar: "test1"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var found []models.User
	err := query.Model(&models.User{}).WhereNone([]string{"name"}, "INVALID", "test1").Find(&found)
	if err == nil {
		t.Error("Expected error for invalid operator in WhereNone, got nil")
	}
}

func TestTiDBIntegrationWhereNoneInvalidColumn(t *testing.T) {

	db := SetupTiDBTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "test1", Avatar: "test1"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var found []models.User
	err := query.Model(&models.User{}).WhereNone([]string{"invalid column"}, "=", "test1").Find(&found)
	if err == nil {
		t.Error("Expected error for invalid column in WhereNone, got nil")
	}
}
