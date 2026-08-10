//go:build integration

package cockroachdb_test

import (
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/integration_tests/models"
)

func TestCockroachDBIntegrationWhereColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	defer func() {
		_, _ = db.Query().Model(&models.User{}).Where("name IN ?", []string{"user1", "user2", "same"}).Delete(&models.User{})
	}()

	user1 := models.User{Name: "user1"}
	user2 := models.User{Name: "user2"}
	err := query.Model(&models.User{}).Create(&user1)
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	err = query.Model(&models.User{}).Create(&user2)
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	user2.Avatar = "user2_avatar"
	_, err = query.Model(&models.User{}).Where("id = ?", user2.ID).Update("avatar", "user2_avatar")
	if err != nil {
		t.Fatalf("Failed to update user2: %v", err)
	}

	var foundUsers []models.User
	err = query.Model(&models.User{}).WhereColumn("name", "=", "avatar").Find(&foundUsers)
	if err != nil {
		t.Fatalf("WhereColumn failed: %v", err)
	}
	if len(foundUsers) != 0 {
		t.Errorf("Expected 0 users matching name=avatar, got %d", len(foundUsers))
	}

	user3 := models.User{Name: "same", Avatar: "same"}
	err = query.Model(&models.User{}).Create(&user3)
	if err != nil {
		t.Fatalf("Failed to create user3: %v", err)
	}

	err = query.Model(&models.User{}).WhereColumn("name", "=", "avatar").Find(&foundUsers)
	if err != nil {
		t.Fatalf("WhereColumn failed: %v", err)
	}
	if len(foundUsers) != 1 {
		t.Errorf("Expected 1 user matching name=avatar, got %d", len(foundUsers))
	}

	err = query.Model(&models.User{}).WhereColumn("created_at", "<", "updated_at").Find(&foundUsers)
	if err != nil {
		t.Fatalf("WhereColumn with < failed: %v", err)
	}

	err = query.Model(&models.User{}).WhereColumn("invalid; column", "=", "name").Find(&foundUsers)
	if err == nil {
		t.Error("Expected error for invalid column in WhereColumn, got nil")
	}

	err = query.Model(&models.User{}).WhereColumn("name", "INVALID", "avatar").Find(&foundUsers)
	if err == nil {
		t.Error("Expected error for invalid operator in WhereColumn, got nil")
	}
}

func TestCockroachDBIntegrationOrWhereColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	defer func() {
		_, _ = db.Query().Model(&models.User{}).Where("name IN ?", []string{"match", "other", "user1"}).Delete(&models.User{})
	}()

	users := []models.User{
		{Name: "match", Avatar: "match"},
		{Name: "other", Avatar: "nomatch"},
		{Name: "user1", Avatar: "nomatch"},
	}
	err := query.Model(&models.User{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var foundUsers []models.User
	err = query.Model(&models.User{}).Where("name = ?", "user1").OrWhereColumn("name", "=", "avatar").Find(&foundUsers)
	if err != nil {
		t.Fatalf("OrWhereColumn failed: %v", err)
	}

	if len(foundUsers) != 2 {
		t.Errorf("Expected 2 users, got %d", len(foundUsers))
	}

	err = query.Model(&models.User{}).OrWhereColumn("invalid; column", "=", "name").Find(&foundUsers)
	if err == nil {
		t.Error("Expected error for invalid column in OrWhereColumn, got nil")
	}
}

func TestCockroachDBIntegrationWhereExists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	defer func() {
		_, _ = db.Query().Model(&models.User{}).Where("name = ?", "exists_user").Delete(&models.User{})
		_, _ = db.Query().Model(&models.Address{}).Where("name = ?", "work").Delete(&models.Address{})
	}()

	user := models.User{Name: "exists_user"}
	err := query.Model(&models.User{}).Create(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	address := models.Address{UserID: user.ID, Name: "work"}
	err = query.Model(&models.Address{}).Create(&address)
	if err != nil {
		t.Fatalf("Failed to create address: %v", err)
	}

	var foundUsers []models.User
	err = query.Model(&models.User{}).WhereExists(func(q contractsorm.Query) contractsorm.Query {
		return q.Table("addresses").WhereColumn("addresses.user_id", "=", "users.id")
	}).Find(&foundUsers)

	if err != nil {
		t.Fatalf("WhereExists failed: %v", err)
	}
	if len(foundUsers) != 1 {
		t.Errorf("Expected 1 user, got %d", len(foundUsers))
	}
}

func TestCockroachDBIntegrationWhereNot(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	defer func() {
		_, _ = db.Query().Model(&models.User{}).Where("name IN ?", []string{"user1", "user2"}).Delete(&models.User{})
	}()

	users := []models.User{
		{Name: "user1"},
		{Name: "user2"},
	}
	err := query.Model(&models.User{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var foundUsers []models.User
	err = query.Model(&models.User{}).WhereNot("name = ?", "user1").Find(&foundUsers)
	if err != nil {
		t.Fatalf("WhereNot failed: %v", err)
	}

	if len(foundUsers) != 1 || foundUsers[0].Name != "user2" {
		t.Errorf("Expected only user2, got %v", foundUsers)
	}

	var foundUsersNested []models.User
	err = query.Model(&models.User{}).WhereNot(func(q contractsorm.Query) contractsorm.Query {
		return q.Where("name = ?", "user1")
	}).Find(&foundUsersNested)

	if err != nil {
		t.Fatalf("WhereNot nested failed: %v", err)
	}
	if len(foundUsersNested) != 1 || foundUsersNested[0].Name != "user2" {
		t.Errorf("Expected only user2 for nested WhereNot, got %v", foundUsersNested)
	}
}

func TestCockroachDBIntegrationOrWhereNot(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	defer func() {
		_, _ = db.Query().Model(&models.User{}).Where("name IN ?", []string{"user1", "user2", "user3"}).Delete(&models.User{})
	}()

	users := []models.User{
		{Name: "user1", Avatar: "avatar1"},
		{Name: "user2", Avatar: "avatar2"},
		{Name: "user3", Avatar: "avatar3"},
	}
	err := query.Model(&models.User{}).Create(&users)
	if err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	var foundUsers []models.User
	err = query.Model(&models.User{}).Where("name = ?", "user1").OrWhereNot("avatar = ?", "avatar2").Find(&foundUsers)
	if err != nil {
		t.Fatalf("OrWhereNot failed: %v", err)
	}

	if len(foundUsers) != 2 {
		t.Errorf("Expected 2 users, got %d", len(foundUsers))
	}

	foundNames := make([]string, len(foundUsers))
	for i, u := range foundUsers {
		foundNames[i] = u.Name
	}
	hasUser1 := false
	hasUser3 := false
	for _, name := range foundNames {
		if name == "user1" {
			hasUser1 = true
		}
		if name == "user3" {
			hasUser3 = true
		}
	}
	if !hasUser1 || !hasUser3 {
		t.Errorf("Expected user1 and user3, got %v", foundNames)
	}

	var foundUsersClosure []models.User
	err = query.Model(&models.User{}).Where("name = ?", "user1").OrWhereNot(func(q contractsorm.Query) contractsorm.Query {
		return q.Where("avatar = ?", "avatar2")
	}).Find(&foundUsersClosure)

	if err != nil {
		t.Fatalf("OrWhereNot closure failed: %v", err)
	}
	if len(foundUsersClosure) != 3 {
		t.Errorf("Expected 3 users for OrWhereNot closure, got %d", len(foundUsersClosure))
	}
}

func TestCockroachDBIntegrationExists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupCockroachDBTest(t)
	query := db.Query()

	defer func() {
		_, _ = db.Query().Model(&models.User{}).Where("name = ?", "existent").Delete(&models.User{})
	}()

	var exists bool
	err := query.Model(&models.User{}).Where("name = ?", "non-existent").Exists(&exists)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Expected exists to be false")
	}

	user := models.User{Name: "existent"}
	err = query.Model(&models.User{}).Create(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	err = query.Model(&models.User{}).Where("name = ?", "existent").Exists(&exists)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Expected exists to be true")
	}
}
