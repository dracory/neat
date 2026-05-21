package observer

import (
	"context"
	"testing"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/contracts/log"
)

// MockObserver is a simple mock observer for testing
type MockObserver struct {
	CreatingCalled  bool
	CreatedCalled   bool
	UpdatingCalled  bool
	UpdatedCalled   bool
	DeletingCalled  bool
	DeletedCalled   bool
	SavingCalled    bool
	SavedCalled     bool
	RetrievedCalled bool
}

func (m *MockObserver) Creating(event contractsorm.Event) error {
	m.CreatingCalled = true
	return nil
}

func (m *MockObserver) Created(event contractsorm.Event) error {
	m.CreatedCalled = true
	return nil
}

func (m *MockObserver) Updating(event contractsorm.Event) error {
	m.UpdatingCalled = true
	return nil
}

func (m *MockObserver) Updated(event contractsorm.Event) error {
	m.UpdatedCalled = true
	return nil
}

func (m *MockObserver) Deleting(event contractsorm.Event) error {
	m.DeletingCalled = true
	return nil
}

func (m *MockObserver) Deleted(event contractsorm.Event) error {
	m.DeletedCalled = true
	return nil
}

func (m *MockObserver) Saving(event contractsorm.Event) error {
	m.SavingCalled = true
	return nil
}

func (m *MockObserver) Saved(event contractsorm.Event) error {
	m.SavedCalled = true
	return nil
}

func (m *MockObserver) Retrieved(event contractsorm.Event) error {
	m.RetrievedCalled = true
	return nil
}

func (m *MockObserver) ForceDeleting(event contractsorm.Event) error {
	return nil
}

func (m *MockObserver) ForceDeleted(event contractsorm.Event) error {
	return nil
}

func (m *MockObserver) Restoring(event contractsorm.Event) error {
	return nil
}

func (m *MockObserver) Restored(event contractsorm.Event) error {
	return nil
}

// TestModel is a simple model for testing
type TestModel struct {
	ID   uint
	Name string
}

func TestEventCreation(t *testing.T) {
	ctx := context.Background()
	model := &TestModel{ID: 1, Name: "Test"}
	original := map[string]any{"id": uint(0)}
	attributes := map[string]any{"id": uint(1), "name": "Test"}
	dirty := map[string]bool{"id": true, "name": true}

	event := NewEvent(ctx, model, original, attributes, dirty, nil, contractsorm.EventCreating)

	if event.Context() != ctx {
		t.Error("Event context not set correctly")
	}

	if event.Model() != model {
		t.Error("Event model not set correctly")
	}

	if event.EventType() != contractsorm.EventCreating {
		t.Error("Event type not set correctly")
	}

	if event.GetAttribute("name") != "Test" {
		t.Error("GetAttribute not working correctly")
	}

	if event.GetOriginal("id", 0) != uint(0) {
		t.Error("GetOriginal not working correctly")
	}

	if !event.IsDirty("name") {
		t.Error("IsDirty not working correctly")
	}

	if event.IsClean("name") {
		t.Error("IsClean not working correctly")
	}

	if !event.IsClean("nonexistent") {
		t.Error("IsClean should return true for non-existent columns")
	}

	event.SetAttribute("new_field", "value")
	if event.GetAttribute("new_field") != "value" {
		t.Error("SetAttribute not working correctly")
	}
}

func TestExtractModelAttributes(t *testing.T) {
	model := &TestModel{ID: 1, Name: "Test"}
	attrs := ExtractModelAttributes(model)

	if len(attrs) == 0 {
		t.Error("ExtractModelAttributes returned empty map")
	}

	if attrs["Name"] != "Test" {
		t.Error("ExtractModelAttributes not extracting Name field correctly")
	}

	if attrs["ID"] != uint(1) {
		t.Error("ExtractModelAttributes not extracting ID field correctly")
	}
}

func TestDispatcherDispatchCreating(t *testing.T) {
	log := &log.NoopLogger{}
	dispatcher := NewDispatcher(log)

	model := &TestModel{ID: 1, Name: "Test"}
	observer := &MockObserver{}

	observers := []contractsorm.ModelToObserver{
		{Model: &TestModel{}, Observer: observer},
	}

	err := dispatcher.DispatchCreating(context.Background(), model, observers, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("DispatchCreating failed: %v", err)
	}

	if !observer.CreatingCalled {
		t.Error("Creating event was not called")
	}
}

func TestDispatcherDispatchCreated(t *testing.T) {
	log := &log.NoopLogger{}
	dispatcher := NewDispatcher(log)

	model := &TestModel{ID: 1, Name: "Test"}
	observer := &MockObserver{}

	observers := []contractsorm.ModelToObserver{
		{Model: &TestModel{}, Observer: observer},
	}

	err := dispatcher.DispatchCreated(context.Background(), model, observers, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("DispatchCreated failed: %v", err)
	}

	if !observer.CreatedCalled {
		t.Error("Created event was not called")
	}
}

func TestDispatcherDispatchUpdating(t *testing.T) {
	log := &log.NoopLogger{}
	dispatcher := NewDispatcher(log)

	model := &TestModel{ID: 1, Name: "Test"}
	observer := &MockObserver{}

	observers := []contractsorm.ModelToObserver{
		{Model: &TestModel{}, Observer: observer},
	}

	err := dispatcher.DispatchUpdating(context.Background(), model, observers, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("DispatchUpdating failed: %v", err)
	}

	if !observer.UpdatingCalled {
		t.Error("Updating event was not called")
	}
}

func TestDispatcherDispatchUpdated(t *testing.T) {
	log := &log.NoopLogger{}
	dispatcher := NewDispatcher(log)

	model := &TestModel{ID: 1, Name: "Test"}
	observer := &MockObserver{}

	observers := []contractsorm.ModelToObserver{
		{Model: &TestModel{}, Observer: observer},
	}

	err := dispatcher.DispatchUpdated(context.Background(), model, observers, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("DispatchUpdated failed: %v", err)
	}

	if !observer.UpdatedCalled {
		t.Error("Updated event was not called")
	}
}

func TestDispatcherDispatchDeleting(t *testing.T) {
	log := &log.NoopLogger{}
	dispatcher := NewDispatcher(log)

	model := &TestModel{ID: 1, Name: "Test"}
	observer := &MockObserver{}

	observers := []contractsorm.ModelToObserver{
		{Model: &TestModel{}, Observer: observer},
	}

	err := dispatcher.DispatchDeleting(context.Background(), model, observers, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("DispatchDeleting failed: %v", err)
	}

	if !observer.DeletingCalled {
		t.Error("Deleting event was not called")
	}
}

func TestDispatcherDispatchDeleted(t *testing.T) {
	log := &log.NoopLogger{}
	dispatcher := NewDispatcher(log)

	model := &TestModel{ID: 1, Name: "Test"}
	observer := &MockObserver{}

	observers := []contractsorm.ModelToObserver{
		{Model: &TestModel{}, Observer: observer},
	}

	err := dispatcher.DispatchDeleted(context.Background(), model, observers, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("DispatchDeleted failed: %v", err)
	}

	if !observer.DeletedCalled {
		t.Error("Deleted event was not called")
	}
}

func TestDispatcherMultipleObservers(t *testing.T) {
	log := &log.NoopLogger{}
	dispatcher := NewDispatcher(log)

	model := &TestModel{ID: 1, Name: "Test"}
	observer1 := &MockObserver{}
	observer2 := &MockObserver{}

	observers := []contractsorm.ModelToObserver{
		{Model: &TestModel{}, Observer: observer1},
		{Model: &TestModel{}, Observer: observer2},
	}

	err := dispatcher.DispatchCreated(context.Background(), model, observers, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("DispatchCreated failed: %v", err)
	}

	if !observer1.CreatedCalled {
		t.Error("Created event was not called on observer1")
	}

	if !observer2.CreatedCalled {
		t.Error("Created event was not called on observer2")
	}
}

func TestDispatcherNoObservers(t *testing.T) {
	log := &log.NoopLogger{}
	dispatcher := NewDispatcher(log)

	model := &TestModel{ID: 1, Name: "Test"}
	observers := []contractsorm.ModelToObserver{}

	err := dispatcher.DispatchCreated(context.Background(), model, observers, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("DispatchCreated should not fail with no observers: %v", err)
	}
}
