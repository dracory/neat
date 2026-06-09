package query

import (
	"sync"
	"testing"

	"github.com/dracory/neat/database/db"
)

func TestEnableDebug(t *testing.T) {
	q := &Query{
		dbConfig: &db.DBConfig{Debug: false},
	}

	// Initially debug should be false
	if q.IsDebug() {
		t.Error("Expected debug to be false initially")
	}

	// Enable debug
	q.EnableDebug()

	// Debug should now be true
	if !q.IsDebug() {
		t.Error("Expected debug to be true after EnableDebug")
	}

	// dbConfig.Debug should also be true
	if !q.dbConfig.Debug {
		t.Error("Expected dbConfig.Debug to be true after EnableDebug")
	}
}

func TestDisableDebug(t *testing.T) {
	q := &Query{
		dbConfig: &db.DBConfig{Debug: true},
	}

	// Initially debug should be true
	if !q.IsDebug() {
		t.Error("Expected debug to be true initially")
	}

	// Disable debug
	q.DisableDebug()

	// Debug should now be false
	if q.IsDebug() {
		t.Error("Expected debug to be false after DisableDebug")
	}

	// dbConfig.Debug should also be false
	if q.dbConfig.Debug {
		t.Error("Expected dbConfig.Debug to be false after DisableDebug")
	}
}

func TestIsDebug(t *testing.T) {
	tests := []struct {
		name          string
		debugState    bool
		dbConfigDebug bool
		expected      bool
	}{
		{
			name:          "both false",
			debugState:    false,
			dbConfigDebug: false,
			expected:      false,
		},
		{
			name:          "debugState true",
			debugState:    true,
			dbConfigDebug: false,
			expected:      true,
		},
		{
			name:          "dbConfigDebug true",
			debugState:    false,
			dbConfigDebug: true,
			expected:      true,
		},
		{
			name:          "both true",
			debugState:    true,
			dbConfigDebug: true,
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Query{
				debugState: tt.debugState,
				dbConfig:   &db.DBConfig{Debug: tt.dbConfigDebug},
			}

			if q.IsDebug() != tt.expected {
				t.Errorf("IsDebug() = %v, expected %v", q.IsDebug(), tt.expected)
			}
		})
	}
}

func TestIsDebugWithNilDbConfig(t *testing.T) {
	q := &Query{
		debugState: true,
		dbConfig:   nil,
	}

	// Should return true based on debugState
	if !q.IsDebug() {
		t.Error("Expected IsDebug to return true when debugState is true and dbConfig is nil")
	}

	q.debugState = false

	// Should return false when both are false/nil
	if q.IsDebug() {
		t.Error("Expected IsDebug to return false when debugState is false and dbConfig is nil")
	}
}

func TestEnableDebugThreadSafety(t *testing.T) {
	q := &Query{
		dbConfig: &db.DBConfig{Debug: false},
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrently enable debug
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			q.EnableDebug()
		}()
	}

	wg.Wait()

	// Debug should be true
	if !q.IsDebug() {
		t.Error("Expected debug to be true after concurrent EnableDebug calls")
	}
}

func TestDisableDebugThreadSafety(t *testing.T) {
	q := &Query{
		dbConfig: &db.DBConfig{Debug: true},
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrently disable debug
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			q.DisableDebug()
		}()
	}

	wg.Wait()

	// Debug should be false
	if q.IsDebug() {
		t.Error("Expected debug to be false after concurrent DisableDebug calls")
	}
}

func TestIsDebugThreadSafety(t *testing.T) {
	q := &Query{
		dbConfig: &db.DBConfig{Debug: false},
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrently toggle debug and check IsDebug
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			q.EnableDebug()
			_ = q.IsDebug()
			q.DisableDebug()
			_ = q.IsDebug()
		}()
	}

	wg.Wait()

	// Final state should be false (last operation was DisableDebug)
	if q.IsDebug() {
		t.Error("Expected debug to be false after concurrent toggle operations")
	}
}

func TestConcurrentEnableDisable(t *testing.T) {
	q := &Query{
		dbConfig: &db.DBConfig{Debug: false},
	}

	var wg sync.WaitGroup
	numGoroutines := 50

	// Concurrently enable and disable debug
	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			q.EnableDebug()
		}()
		go func() {
			defer wg.Done()
			q.DisableDebug()
		}()
	}

	wg.Wait()

	// The final state is non-deterministic, but the operations should not panic
	_ = q.IsDebug()
}
