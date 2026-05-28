package turso

import (
	"testing"

	"github.com/dracory/neat/database"
	"github.com/dracory/neat/integration_tests/models"
)

func seedOrderLimitOffsetTestData(t *testing.T, db *database.Database) {
	query := db.Query()
	users := []models.User{
		{Name: "user_b", Avatar: "avatar2"},
		{Name: "user_a", Avatar: "avatar1"},
		{Name: "user_c", Avatar: "avatar3"},
		{Name: "user_d", Avatar: "avatar4"},
		{Name: "user_e", Avatar: "avatar5"},
	}
	if err := query.Model(&models.User{}).Create(&users); err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}
}

func TestTursoIntegrationOrderByAscending(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedOrderLimitOffsetTestData(t, db)

	results := make([]models.User, 0)
	err := db.Query().Model(&models.User{}).OrderBy("name", "asc").Get(&results)
	if err != nil {
		t.Errorf("OrderBy ascending failed: %v", err)
	}
	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}
	if results[0].Name != "user_a" {
		t.Errorf("Expected 'user_a', got '%s'", results[0].Name)
	}
	if results[1].Name != "user_b" {
		t.Errorf("Expected 'user_b', got '%s'", results[1].Name)
	}
	if results[2].Name != "user_c" {
		t.Errorf("Expected 'user_c', got '%s'", results[2].Name)
	}
}

func TestTursoIntegrationOrderByDescending(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedOrderLimitOffsetTestData(t, db)

	results := make([]models.User, 0)
	err := db.Query().Model(&models.User{}).OrderBy("name", "desc").Get(&results)
	if err != nil {
		t.Errorf("OrderBy descending failed: %v", err)
	}
	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}
	if results[0].Name != "user_e" {
		t.Errorf("Expected 'user_e', got '%s'", results[0].Name)
	}
	if results[1].Name != "user_d" {
		t.Errorf("Expected 'user_d', got '%s'", results[1].Name)
	}
	if results[2].Name != "user_c" {
		t.Errorf("Expected 'user_c', got '%s'", results[2].Name)
	}
}

func TestTursoIntegrationOrderByDescMethod(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedOrderLimitOffsetTestData(t, db)

	results := make([]models.User, 0)
	err := db.Query().Model(&models.User{}).OrderByDesc("name").Get(&results)
	if err != nil {
		t.Errorf("OrderByDesc method failed: %v", err)
	}
	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}
	if results[0].Name != "user_e" {
		t.Errorf("Expected 'user_e', got '%s'", results[0].Name)
	}
}

func TestTursoIntegrationMultipleOrderByClauses(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedOrderLimitOffsetTestData(t, db)
	query := db.Query()

	if err := query.Model(&models.User{}).Create(&models.User{Name: "user_a", Avatar: "avatar0"}); err != nil {
		t.Fatalf("Failed to create additional user: %v", err)
	}

	results := make([]models.User, 0)
	err := db.Query().Model(&models.User{}).OrderBy("name", "asc").OrderBy("avatar", "asc").Get(&results)
	if err != nil {
		t.Errorf("Multiple OrderBy clauses failed: %v", err)
	}

	userAEntries := make([]models.User, 0)
	for _, u := range results {
		if u.Name == "user_a" {
			userAEntries = append(userAEntries, u)
		}
	}
	if len(userAEntries) != 2 {
		t.Errorf("Expected 2 user_a entries, got %d", len(userAEntries))
	}
	if userAEntries[0].Avatar != "avatar0" {
		t.Errorf("Expected 'avatar0', got '%s'", userAEntries[0].Avatar)
	}
	if userAEntries[1].Avatar != "avatar1" {
		t.Errorf("Expected 'avatar1', got '%s'", userAEntries[1].Avatar)
	}
}

func TestTursoIntegrationOrderByWithExpressions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedOrderLimitOffsetTestData(t, db)

	var results []models.User
	err := db.Query().Model(&models.User{}).Order("LENGTH(name) DESC").OrderBy("name", "asc").Get(&results)
	if err != nil {
		t.Errorf("OrderBy with expressions failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("Expected non-empty results")
	}
}

func TestTursoIntegrationLimitClause(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedOrderLimitOffsetTestData(t, db)

	var results []models.User
	err := db.Query().Model(&models.User{}).Limit(2).Get(&results)
	if err != nil {
		t.Errorf("Limit clause failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestTursoIntegrationLimitWithOrderBy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedOrderLimitOffsetTestData(t, db)
	query := db.Query()

	if err := query.Model(&models.User{}).Create(&models.User{Name: "user_a", Avatar: "avatar0"}); err != nil {
		t.Fatalf("Failed to create additional user: %v", err)
	}

	results := make([]models.User, 0)
	err := db.Query().Model(&models.User{}).OrderBy("name", "asc").Limit(2).Get(&results)
	if err != nil {
		t.Errorf("Limit with OrderBy failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	if results[0].Name != "user_a" {
		t.Errorf("Expected 'user_a', got '%s'", results[0].Name)
	}
	if results[1].Name != "user_a" {
		t.Errorf("Expected 'user_a', got '%s'", results[1].Name)
	}
}

func TestTursoIntegrationLimitEdgeCaseZero(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedOrderLimitOffsetTestData(t, db)

	var results []models.User
	err := db.Query().Model(&models.User{}).Limit(0).Get(&results)
	if err != nil {
		t.Errorf("Limit edge case (zero) failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestTursoIntegrationLimitEdgeCaseNegative(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedOrderLimitOffsetTestData(t, db)

	var results []models.User
	err := db.Query().Model(&models.User{}).Limit(-1).Get(&results)
	if err != nil {
		t.Errorf("Limit edge case (negative) failed: %v", err)
	}
	if len(results) < 5 {
		t.Errorf("Expected at least 5 results, got %d", len(results))
	}
}

func TestTursoIntegrationOffsetClause(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedOrderLimitOffsetTestData(t, db)
	query := db.Query()

	if err := query.Model(&models.User{}).Create(&models.User{Name: "user_a", Avatar: "avatar0"}); err != nil {
		t.Fatalf("Failed to create additional user: %v", err)
	}

	var results []models.User
	err := db.Query().Model(&models.User{}).OrderBy("name", "asc").Offset(2).Get(&results)
	if err != nil {
		t.Errorf("Offset clause failed: %v", err)
	}
	if len(results) != 4 {
		t.Errorf("Expected 4 results, got %d", len(results))
	}
}

func TestTursoIntegrationOffsetWithLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedOrderLimitOffsetTestData(t, db)
	query := db.Query()

	if err := query.Model(&models.User{}).Create(&models.User{Name: "user_a", Avatar: "avatar0"}); err != nil {
		t.Fatalf("Failed to create additional user: %v", err)
	}

	results := make([]models.User, 0)
	err := db.Query().Model(&models.User{}).OrderBy("name", "asc").Offset(2).Limit(2).Get(&results)
	if err != nil {
		t.Errorf("Offset with Limit failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	if results[0].Name != "user_b" {
		t.Errorf("Expected 'user_b', got '%s'", results[0].Name)
	}
	if results[1].Name != "user_c" {
		t.Errorf("Expected 'user_c', got '%s'", results[1].Name)
	}
}

func TestTursoIntegrationInRandomOrder(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedOrderLimitOffsetTestData(t, db)
	query := db.Query()

	if err := query.Model(&models.User{}).Create(&models.User{Name: "user_a", Avatar: "avatar0"}); err != nil {
		t.Fatalf("Failed to create additional user: %v", err)
	}

	var results1 []models.User
	err := db.Query().Model(&models.User{}).InRandomOrder().Get(&results1)
	if err != nil {
		t.Errorf("InRandomOrder method failed: %v", err)
	}

	var results2 []models.User
	err = db.Query().Model(&models.User{}).InRandomOrder().Get(&results2)
	if err != nil {
		t.Errorf("InRandomOrder method (second) failed: %v", err)
	}

	if len(results1) != 6 {
		t.Errorf("Expected 6 results, got %d", len(results1))
	}
	if len(results2) != 6 {
		t.Errorf("Expected 6 results, got %d", len(results2))
	}
}

func TestTursoIntegrationRandomOrderingWithLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTursoTest(t)
	seedOrderLimitOffsetTestData(t, db)
	query := db.Query()

	if err := query.Model(&models.User{}).Create(&models.User{Name: "user_a", Avatar: "avatar0"}); err != nil {
		t.Fatalf("Failed to create additional user: %v", err)
	}

	var results []models.User
	err := db.Query().Model(&models.User{}).InRandomOrder().Limit(3).Get(&results)
	if err != nil {
		t.Errorf("Random ordering with Limit failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
}
