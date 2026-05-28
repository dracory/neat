package query_test

import (
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/database/query"
)

// --- Observer registration tests ---

type TestObserver struct {
	CreatedCalled  bool
	UpdatedCalled  bool
	DeletedCalled  bool
	CreatingCalled bool
	UpdatingCalled bool
	DeletingCalled bool
	SavingCalled   bool
	SavedCalled    bool
}

func (o *TestObserver) Created(event contractsorm.Event) error {
	o.CreatedCalled = true
	return nil
}

func (o *TestObserver) Updated(event contractsorm.Event) error {
	o.UpdatedCalled = true
	return nil
}

func (o *TestObserver) Deleted(event contractsorm.Event) error {
	o.DeletedCalled = true
	return nil
}

func (o *TestObserver) Creating(event contractsorm.Event) error {
	o.CreatingCalled = true
	return nil
}

func (o *TestObserver) Updating(event contractsorm.Event) error {
	o.UpdatingCalled = true
	return nil
}

func (o *TestObserver) Deleting(event contractsorm.Event) error {
	o.DeletingCalled = true
	return nil
}

func (o *TestObserver) ForceDeleted(event contractsorm.Event) error {
	return nil
}

func (o *TestObserver) ForceDeleting(event contractsorm.Event) error {
	return nil
}

func (o *TestObserver) Restored(event contractsorm.Event) error {
	return nil
}

func (o *TestObserver) Retrieved(event contractsorm.Event) error {
	return nil
}

type TestModel struct {
	ID   uint
	Name string
}

func (TestModel) TableName() string {
	return "observer_test"
}

func TestObserveRegistersObserver(t *testing.T) {
	w := openSQLiteQuery(t)
	observer := &TestObserver{}

	w.Q.Observe(&TestModel{}, observer)

	wrapped := query.WrapQuery(w.Q)
	observers := wrapped.GetModelToObserver()

	if len(observers) != 1 {
		t.Errorf("expected 1 observer, got %d", len(observers))
	}

	if observers[0].Observer != observer {
		t.Error("observer not registered correctly")
	}
}

func TestObserveMultipleObservers(t *testing.T) {
	w := openSQLiteQuery(t)
	observer1 := &TestObserver{}
	observer2 := &TestObserver{}

	w.Q.Observe(&TestModel{}, observer1)
	w.Q.Observe(&TestModel{}, observer2)

	wrapped := query.WrapQuery(w.Q)
	observers := wrapped.GetModelToObserver()

	if len(observers) != 2 {
		t.Errorf("expected 2 observers, got %d", len(observers))
	}
}

func TestObserveWithDifferentModels(t *testing.T) {
	w := openSQLiteQuery(t)
	observer1 := &TestObserver{}
	observer2 := &TestObserver{}

	type AnotherModel struct {
		ID uint
	}

	w.Q.Observe(&TestModel{}, observer1)
	w.Q.Observe(&AnotherModel{}, observer2)

	wrapped := query.WrapQuery(w.Q)
	observers := wrapped.GetModelToObserver()

	if len(observers) != 2 {
		t.Errorf("expected 2 observers, got %d", len(observers))
	}
}

func TestWithoutEventsDisablesEvents(t *testing.T) {
	w := openSQLiteQuery(t)

	newQ := w.Q.WithoutEvents()
	wrapped := query.WrapQuery(newQ.(*query.Query))

	if !wrapped.GetWithoutEvents() {
		t.Error("WithoutEvents should set withoutEvents flag")
	}
}

func TestWithoutEventsReturnsNewQuery(t *testing.T) {
	w := openSQLiteQuery(t)

	newQ := w.Q.WithoutEvents()
	if newQ == w.Q {
		t.Error("WithoutEvents should return a new Query instance")
	}

	wrapped := query.WrapQuery(w.Q)
	if wrapped.GetWithoutEvents() {
		t.Error("original query should not have withoutEvents flag set")
	}
}

func TestObserverDispatchDuringCreate(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE observer_test (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")

	observer := &TestObserver{}
	w.Q.Observe(&TestModel{}, observer)

	model := &TestModel{Name: "test"}
	err := w.Q.Model(model).Create(model)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if !observer.CreatingCalled {
		t.Error("Creating observer was not called during Create")
	}
	if !observer.CreatedCalled {
		t.Error("Created observer was not called during Create")
	}
}

func TestObserverDispatchDuringUpdate(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE observer_test (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")
	execSQL(t, w, "INSERT INTO observer_test (name) VALUES ('original')")

	observer := &TestObserver{}
	w.Q.Observe(&TestModel{}, observer)

	model := &TestModel{Name: "updated"}
	_, err := w.Q.Model(model).Where("id", 1).Update(model)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if !observer.UpdatingCalled {
		t.Error("Updating observer was not called during Update")
	}
	if !observer.UpdatedCalled {
		t.Error("Updated observer was not called during Update")
	}
}

func TestObserverDispatchDuringDelete(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE observer_test (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")
	execSQL(t, w, "INSERT INTO observer_test (name) VALUES ('test')")

	observer := &TestObserver{}
	w.Q.Observe(&TestModel{}, observer)

	model := &TestModel{}
	_, err := w.Q.Model(model).Where("id", 1).Delete()
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if !observer.DeletingCalled {
		t.Error("Deleting observer was not called during Delete")
	}
	if !observer.DeletedCalled {
		t.Error("Deleted observer was not called during Delete")
	}
}

func TestObserverDispatchWithoutEvents(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE observer_test (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")

	observer := &TestObserver{}
	w.Q.Observe(&TestModel{}, observer)

	w.Q = w.Q.WithoutEvents().(*query.Query)

	model := &TestModel{Name: "test"}
	err := w.Q.Model(model).Create(model)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if observer.CreatingCalled {
		t.Error("Creating observer should not be called when WithoutEvents is set")
	}
	if observer.CreatedCalled {
		t.Error("Created observer should not be called when WithoutEvents is set")
	}
}

func TestMultipleObserversDispatchDuringCreate(t *testing.T) {
	w := openSQLiteQuery(t)
	execSQL(t, w, "CREATE TABLE observer_test (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")

	observer1 := &TestObserver{}
	observer2 := &TestObserver{}
	w.Q.Observe(&TestModel{}, observer1)
	w.Q.Observe(&TestModel{}, observer2)

	model := &TestModel{Name: "test"}
	err := w.Q.Model(model).Create(model)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if !observer1.CreatedCalled {
		t.Error("Created observer was not called on observer1")
	}
	if !observer2.CreatedCalled {
		t.Error("Created observer was not called on observer2")
	}
}
