package main_test

import (
	"sync"
	"testing"

	"github.com/dracory/neat"
	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/contracts/database/schema"
	mainpkg "github.com/dracory/neat/examples/observers"
)

func TestRunExample(t *testing.T) {
	err := mainpkg.RunExample("sqlite://:memory:")
	if err != nil {
		t.Fatalf("RunExample failed: %v", err)
	}
}

// trackingObserver records every event it receives.
type trackingObserver struct {
	mu               sync.Mutex
	CreatingCount    int
	CreatedCount     int
	UpdatingCount    int
	UpdatedCount     int
	DeletingCount    int
	DeletedCount     int
	ForceDeleteCount int
}

func (o *trackingObserver) Creating(_ contractsorm.Event) error {
	o.mu.Lock()
	o.CreatingCount++
	o.mu.Unlock()
	return nil
}
func (o *trackingObserver) Created(_ contractsorm.Event) error {
	o.mu.Lock()
	o.CreatedCount++
	o.mu.Unlock()
	return nil
}
func (o *trackingObserver) Updating(_ contractsorm.Event) error {
	o.mu.Lock()
	o.UpdatingCount++
	o.mu.Unlock()
	return nil
}
func (o *trackingObserver) Updated(_ contractsorm.Event) error {
	o.mu.Lock()
	o.UpdatedCount++
	o.mu.Unlock()
	return nil
}
func (o *trackingObserver) Deleting(_ contractsorm.Event) error {
	o.mu.Lock()
	o.DeletingCount++
	o.mu.Unlock()
	return nil
}
func (o *trackingObserver) Deleted(_ contractsorm.Event) error {
	o.mu.Lock()
	o.DeletedCount++
	o.mu.Unlock()
	return nil
}
func (o *trackingObserver) ForceDeleted(_ contractsorm.Event) error {
	o.mu.Lock()
	o.ForceDeleteCount++
	o.mu.Unlock()
	return nil
}

// testUser is the model used for assertions.
type testUser struct {
	ID    uint   `gorm:"column:id;primaryKey"`
	Name  string `gorm:"column:name"`
	Email string `gorm:"column:email"`
}

func (testUser) TableName() string { return "users" }

func setupObserverDB(t *testing.T) *neat.Database {
	t.Helper()
	db, err := neat.NewFromDSN("sqlite://:memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	if err = db.Schema().Create("users", func(bp schema.Blueprint) {
		bp.ID()
		bp.String("name")
		bp.String("email")
	}); err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	return db
}

func TestObserver_Creating_Created_FiredOnCreate(t *testing.T) {
	db := setupObserverDB(t)
	defer func() { _ = db.Close() }()

	obs := &trackingObserver{}
	db.Observe(&testUser{}, obs)

	user := &testUser{Name: "Alice", Email: "alice@example.com"}
	if err := db.Query().Model(&testUser{}).Create(user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if obs.CreatingCount != 1 {
		t.Errorf("expected CreatingCount=1, got %d", obs.CreatingCount)
	}
	if obs.CreatedCount != 1 {
		t.Errorf("expected CreatedCount=1, got %d", obs.CreatedCount)
	}
}

func TestObserver_Updating_Updated_FiredOnUpdate(t *testing.T) {
	db := setupObserverDB(t)
	defer func() { _ = db.Close() }()

	obs := &trackingObserver{}
	db.Observe(&testUser{}, obs)

	user := &testUser{Name: "Bob", Email: "bob@example.com"}
	if err := db.Query().Model(&testUser{}).Create(user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	user.Name = "Robert"
	if _, err := db.Query().Model(&testUser{}).Where("id = ?", user.ID).Update(user); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if obs.UpdatingCount != 1 {
		t.Errorf("expected UpdatingCount=1, got %d", obs.UpdatingCount)
	}
	if obs.UpdatedCount != 1 {
		t.Errorf("expected UpdatedCount=1, got %d", obs.UpdatedCount)
	}
}

func TestObserver_Deleting_Deleted_FiredOnDelete(t *testing.T) {
	db := setupObserverDB(t)
	defer func() { _ = db.Close() }()

	obs := &trackingObserver{}
	db.Observe(&testUser{}, obs)

	user := &testUser{Name: "Carol", Email: "carol@example.com"}
	if err := db.Query().Model(&testUser{}).Create(user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if _, err := db.Query().Model(&testUser{}).Delete(user); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if obs.DeletingCount != 1 {
		t.Errorf("expected DeletingCount=1, got %d", obs.DeletingCount)
	}
	if obs.DeletedCount != 1 {
		t.Errorf("expected DeletedCount=1, got %d", obs.DeletedCount)
	}
}

func TestObserver_NotFiredWithWithoutEvents(t *testing.T) {
	db := setupObserverDB(t)
	defer func() { _ = db.Close() }()

	obs := &trackingObserver{}
	db.Observe(&testUser{}, obs)

	user := &testUser{Name: "Dave", Email: "dave@example.com"}
	if err := db.Query().WithoutEvents().Model(&testUser{}).Create(user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if obs.CreatingCount != 0 {
		t.Errorf("expected CreatingCount=0 with WithoutEvents, got %d", obs.CreatingCount)
	}
	if obs.CreatedCount != 0 {
		t.Errorf("expected CreatedCount=0 with WithoutEvents, got %d", obs.CreatedCount)
	}
}

func TestObserver_MultipleObservers_BothFired(t *testing.T) {
	db := setupObserverDB(t)
	defer func() { _ = db.Close() }()

	obs1 := &trackingObserver{}
	obs2 := &trackingObserver{}
	db.Observe(&testUser{}, obs1)
	db.Observe(&testUser{}, obs2)

	user := &testUser{Name: "Eve", Email: "eve@example.com"}
	if err := db.Query().Model(&testUser{}).Create(user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if obs1.CreatedCount != 1 {
		t.Errorf("obs1: expected CreatedCount=1, got %d", obs1.CreatedCount)
	}
	if obs2.CreatedCount != 1 {
		t.Errorf("obs2: expected CreatedCount=1, got %d", obs2.CreatedCount)
	}
}

func TestObserver_OnlyFiredForRegisteredModel(t *testing.T) {
	db := setupObserverDB(t)
	defer func() { _ = db.Close() }()

	if err := db.Schema().Create("posts", func(bp schema.Blueprint) {
		bp.ID()
		bp.String("title")
	}); err != nil {
		t.Fatalf("failed to create posts: %v", err)
	}

	type post struct {
		ID    uint   `gorm:"column:id;primaryKey"`
		Title string `gorm:"column:title"`
	}

	obs := &trackingObserver{}
	db.Observe(&testUser{}, obs) // only watching User, not post

	p := &post{Title: "Hello"}
	if err := db.Query().Model(&post{}).Create(p); err != nil {
		t.Fatalf("Create post failed: %v", err)
	}

	if obs.CreatedCount != 0 {
		t.Errorf("expected CreatedCount=0 for unregistered model, got %d", obs.CreatedCount)
	}
}
