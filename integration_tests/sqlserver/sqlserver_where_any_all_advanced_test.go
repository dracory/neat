package sqlserver

import (
	"testing"

	"github.com/dracory/neat/integration_tests/models"
)

func TestSQLServerIntegrationWhereAnyAdvanced(t *testing.T) {

	db := SetupSQLServerTest(t)
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

func TestSQLServerIntegrationWhereAllAdvanced(t *testing.T) {

	db := SetupSQLServerTest(t)
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

func TestSQLServerIntegrationWhereNone(t *testing.T) {

	db := SetupSQLServerTest(t)
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

func TestSQLServerIntegrationWhereNoneAdvanced(t *testing.T) {

	db := SetupSQLServerTest(t)
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

func TestSQLServerIntegrationWhereEdgeCases(t *testing.T) {

	db := SetupSQLServerTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "test1", Avatar: "test1"},
	}
	err := query.Model(&models.User{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}
}

func TestSQLServerIntegrationWhereAnyEmptyColumns(t *testing.T) {

	db := SetupSQLServerTest(t)
	var found []models.User
	err := db.Query().Model(&models.User{}).WhereAny([]string{}, "=", "test1").Find(&found)
	if err != nil {
		t.Errorf("WhereAny with empty columns failed: %v", err)
	}
}

func TestSQLServerIntegrationWhereAnyInvalidOperator(t *testing.T) {

	db := SetupSQLServerTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "test1", Avatar: "test1"},
	}
	query.Model(&models.User{}).Create(&users)

	var found []models.User
	err := query.Model(&models.User{}).WhereAny([]string{"name"}, "INVALID", "test1").Find(&found)
	if err == nil {
		t.Error("Expected error for invalid operator in WhereAny, got nil")
	}
}

func TestSQLServerIntegrationWhereAnyInvalidColumn(t *testing.T) {

	db := SetupSQLServerTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "test1", Avatar: "test1"},
	}
	query.Model(&models.User{}).Create(&users)

	var found []models.User
	err := query.Model(&models.User{}).WhereAny([]string{"invalid column"}, "=", "test1").Find(&found)
	if err == nil {
		t.Error("Expected error for invalid column in WhereAny, got nil")
	}
}

func TestSQLServerIntegrationWhereAllEmptyColumns(t *testing.T) {

	db := SetupSQLServerTest(t)
	var found []models.User
	err := db.Query().Model(&models.User{}).WhereAll([]string{}, "=", "test1").Find(&found)
	if err != nil {
		t.Errorf("WhereAll with empty columns failed: %v", err)
	}
}

func TestSQLServerIntegrationWhereAllInvalidOperator(t *testing.T) {

	db := SetupSQLServerTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "test1", Avatar: "test1"},
	}
	query.Model(&models.User{}).Create(&users)

	var found []models.User
	err := query.Model(&models.User{}).WhereAll([]string{"name"}, "INVALID", "test1").Find(&found)
	if err == nil {
		t.Error("Expected error for invalid operator in WhereAll, got nil")
	}
}

func TestSQLServerIntegrationWhereAllInvalidColumn(t *testing.T) {

	db := SetupSQLServerTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "test1", Avatar: "test1"},
	}
	query.Model(&models.User{}).Create(&users)

	var found []models.User
	err := query.Model(&models.User{}).WhereAll([]string{"invalid column"}, "=", "test1").Find(&found)
	if err == nil {
		t.Error("Expected error for invalid column in WhereAll, got nil")
	}
}

func TestSQLServerIntegrationWhereNoneEmptyColumns(t *testing.T) {

	db := SetupSQLServerTest(t)
	var found []models.User
	err := db.Query().Model(&models.User{}).WhereNone([]string{}, "=", "test1").Find(&found)
	if err != nil {
		t.Errorf("WhereNone with empty columns failed: %v", err)
	}
}

func TestSQLServerIntegrationWhereNoneInvalidOperator(t *testing.T) {

	db := SetupSQLServerTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "test1", Avatar: "test1"},
	}
	query.Model(&models.User{}).Create(&users)

	var found []models.User
	err := query.Model(&models.User{}).WhereNone([]string{"name"}, "INVALID", "test1").Find(&found)
	if err == nil {
		t.Error("Expected error for invalid operator in WhereNone, got nil")
	}
}

func TestSQLServerIntegrationWhereNoneInvalidColumn(t *testing.T) {

	db := SetupSQLServerTest(t)
	query := db.Query()

	users := []models.User{
		{Name: "test1", Avatar: "test1"},
	}
	query.Model(&models.User{}).Create(&users)

	var found []models.User
	err := query.Model(&models.User{}).WhereNone([]string{"invalid column"}, "=", "test1").Find(&found)
	if err == nil {
		t.Error("Expected error for invalid column in WhereNone, got nil")
	}
}
