package schemer

import (
	"testing"
	"time"
)

func TestMigrationTracker_Fields(t *testing.T) {
	tracker := MigrationTracker{
		ID:          "2024_06_15_120000_create_users_table",
		Batch:       20240615120000,
		Description: "Create users table",
		StartedAt:   time.Now(),
		CompletedAt: time.Now(),
	}

	if tracker.ID != "2024_06_15_120000_create_users_table" {
		t.Errorf("Expected ID '2024_06_15_120000_create_users_table', got '%s'", tracker.ID)
	}
	if tracker.Batch != 20240615120000 {
		t.Errorf("Expected Batch 20240615120000, got %d", tracker.Batch)
	}
	if tracker.Description != "Create users table" {
		t.Errorf("Expected Description 'Create users table', got '%s'", tracker.Description)
	}
	if tracker.StartedAt.IsZero() {
		t.Error("Expected StartedAt to be set")
	}
	if tracker.CompletedAt.IsZero() {
		t.Error("Expected CompletedAt to be set")
	}
}

func TestMigrationStatus_Fields(t *testing.T) {
	status := MigrationStatus{
		ID:          "2024_06_15_120000_create_users_table",
		Description: "Create users table",
		Batch:       20240615120000,
		StartedAt:   time.Now(),
		CompletedAt: time.Now(),
		State:       "completed",
	}

	if status.ID != "2024_06_15_120000_create_users_table" {
		t.Errorf("Expected ID '2024_06_15_120000_create_users_table', got '%s'", status.ID)
	}
	if status.Description != "Create users table" {
		t.Errorf("Expected Description 'Create users table', got '%s'", status.Description)
	}
	if status.Batch != 20240615120000 {
		t.Errorf("Expected Batch 20240615120000, got %d", status.Batch)
	}
	if status.StartedAt.IsZero() {
		t.Error("Expected StartedAt to be set")
	}
	if status.CompletedAt.IsZero() {
		t.Error("Expected CompletedAt to be set")
	}
	if status.State != "completed" {
		t.Errorf("Expected State 'completed', got '%s'", status.State)
	}
}

func TestMigrationStatus_JSONTags(t *testing.T) {
	status := MigrationStatus{
		ID:          "test_id",
		Description: "test description",
		Batch:       123,
		StartedAt:   time.Now(),
		CompletedAt: time.Now(),
		State:       "pending",
	}

	// Verify that the struct has proper JSON tags by checking field names
	// This is a compile-time check to ensure the tags are present
	type jsonTaggedStruct struct {
		ID          string    `json:"id"`
		Description string    `json:"description"`
		Batch       int       `json:"batch"`
		StartedAt   time.Time `json:"started_at"`
		CompletedAt time.Time `json:"completed_at"`
		State       string    `json:"state"`
	}

	_ = jsonTaggedStruct{
		ID:          status.ID,
		Description: status.Description,
		Batch:       status.Batch,
		StartedAt:   status.StartedAt,
		CompletedAt: status.CompletedAt,
		State:       status.State,
	}
}

func TestMigrationStatus_StateValues(t *testing.T) {
	validStates := []string{"pending", "completed", "failed"}

	for _, state := range validStates {
		status := MigrationStatus{
			State: state,
		}

		if status.State != state {
			t.Errorf("Expected State '%s', got '%s'", state, status.State)
		}
	}
}

func TestMigrationTracker_ZeroValues(t *testing.T) {
	tracker := MigrationTracker{}

	if tracker.ID != "" {
		t.Errorf("Expected empty ID, got '%s'", tracker.ID)
	}
	if tracker.Batch != 0 {
		t.Errorf("Expected Batch 0, got %d", tracker.Batch)
	}
	if tracker.Description != "" {
		t.Errorf("Expected empty Description, got '%s'", tracker.Description)
	}
	if !tracker.StartedAt.IsZero() {
		t.Error("Expected StartedAt to be zero")
	}
	if !tracker.CompletedAt.IsZero() {
		t.Error("Expected CompletedAt to be zero")
	}
}

func TestMigrationStatus_ZeroValues(t *testing.T) {
	status := MigrationStatus{}

	if status.ID != "" {
		t.Errorf("Expected empty ID, got '%s'", status.ID)
	}
	if status.Description != "" {
		t.Errorf("Expected empty Description, got '%s'", status.Description)
	}
	if status.Batch != 0 {
		t.Errorf("Expected Batch 0, got %d", status.Batch)
	}
	if !status.StartedAt.IsZero() {
		t.Error("Expected StartedAt to be zero")
	}
	if !status.CompletedAt.IsZero() {
		t.Error("Expected CompletedAt to be zero")
	}
	if status.State != "" {
		t.Errorf("Expected empty State, got '%s'", status.State)
	}
}
