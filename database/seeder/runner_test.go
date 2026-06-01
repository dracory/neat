package seeder

import (
	"errors"
	"testing"

	"github.com/dracory/neat/contracts/database/seeder"
	contractsseeder "github.com/dracory/neat/contracts/database/seeder"
)

// Mock seeder for testing with error support
type errorMockSeeder struct {
	signature string
	runCalled bool
	runError  error
}

func (m *errorMockSeeder) Signature() string {
	return m.signature
}

func (m *errorMockSeeder) Run() error {
	m.runCalled = true
	return m.runError
}

func TestNewRunner(t *testing.T) {
	runner := NewRunner()
	if runner == nil {
		t.Fatal("Expected runner to be created")
	}
	if len(runner.GetSeeders()) != 0 {
		t.Errorf("Expected empty seeders list, got %d", len(runner.GetSeeders()))
	}
}

func TestRunnerRegister(t *testing.T) {
	runner := NewRunner()

	seeder1 := &errorMockSeeder{signature: "seeder_1"}
	seeder2 := &errorMockSeeder{signature: "seeder_2"}

	runner.Register([]contractsseeder.Seeder{seeder1, seeder2})

	seeders := runner.GetSeeders()
	if len(seeders) != 2 {
		t.Errorf("Expected 2 seeders, got %d", len(seeders))
	}
}

func TestRunnerGetSeeder(t *testing.T) {
	runner := NewRunner()

	seeder1 := &errorMockSeeder{signature: "test_seeder"}
	runner.Register([]seeder.Seeder{seeder1})

	retrieved := runner.GetSeeder("test_seeder")
	if retrieved == nil {
		t.Fatal("Expected seeder to be found")
	}
	if retrieved.Signature() != "test_seeder" {
		t.Errorf("Expected signature 'test_seeder', got '%s'", retrieved.Signature())
	}

	nonExistent := runner.GetSeeder("non_existent")
	if nonExistent != nil {
		t.Error("Expected nil for non-existent seeder")
	}
}

func TestCall(t *testing.T) {
	runner := NewRunner()

	seeder1 := &errorMockSeeder{signature: "seeder_1"}
	seeder2 := &errorMockSeeder{signature: "seeder_2"}

	runner.Register([]contractsseeder.Seeder{seeder1, seeder2})

	err := runner.Call([]contractsseeder.Seeder{seeder1, seeder2})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !seeder1.runCalled {
		t.Error("Expected seeder1.Run() to be called")
	}
	if !seeder2.runCalled {
		t.Error("Expected seeder2.Run() to be called")
	}
}

func TestCallWithError(t *testing.T) {
	runner := NewRunner()

	seeder1 := &errorMockSeeder{signature: "seeder_1"}
	seeder2 := &errorMockSeeder{signature: "seeder_2", runError: errors.New("seeder error")}

	runner.Register([]contractsseeder.Seeder{seeder1, seeder2})

	err := runner.Call([]contractsseeder.Seeder{seeder1, seeder2})
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !seeder1.runCalled {
		t.Error("Expected seeder1.Run() to be called before error")
	}
}

func TestCallOnce(t *testing.T) {
	runner := NewRunner()

	seeder1 := &errorMockSeeder{signature: "seeder_1"}
	seeder2 := &errorMockSeeder{signature: "seeder_2"}

	runner.Register([]contractsseeder.Seeder{seeder1, seeder2})

	// First call
	err := runner.CallOnce([]contractsseeder.Seeder{seeder1, seeder2})
	if err != nil {
		t.Errorf("Expected no error on first call, got %v", err)
	}

	if !seeder1.runCalled {
		t.Error("Expected seeder1.Run() to be called on first CallOnce")
	}
	if !seeder2.runCalled {
		t.Error("Expected seeder2.Run() to be called on first CallOnce")
	}

	// Reset runCalled flags
	seeder1.runCalled = false
	seeder2.runCalled = false

	// Second call - should skip
	err = runner.CallOnce([]seeder.Seeder{seeder1, seeder2})
	if err != nil {
		t.Errorf("Expected no error on second call, got %v", err)
	}

	if seeder1.runCalled {
		t.Error("Expected seeder1.Run() to NOT be called on second CallOnce")
	}
	if seeder2.runCalled {
		t.Error("Expected seeder2.Run() to NOT be called on second CallOnce")
	}
}

func TestCallOnceWithError(t *testing.T) {
	runner := NewRunner()

	seeder1 := &errorMockSeeder{signature: "seeder_1"}
	seeder2 := &errorMockSeeder{signature: "seeder_2", runError: errors.New("seeder error")}

	runner.Register([]contractsseeder.Seeder{seeder1, seeder2})

	err := runner.CallOnce([]contractsseeder.Seeder{seeder1, seeder2})
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Verify that seeder1 was marked as called even though seeder2 failed
	seeder1.runCalled = false
	err = runner.CallOnce([]contractsseeder.Seeder{seeder1})
	if err != nil {
		t.Errorf("Expected no error on retry of seeder1, got %v", err)
	}
	if seeder1.runCalled {
		t.Error("Expected seeder1 to be skipped on second CallOnce")
	}

	// Verify that seeder2 (which failed) is NOT marked and will re-run
	seeder2.runCalled = false
	err = runner.CallOnce([]contractsseeder.Seeder{seeder2})
	// seeder2 should fail again since it still has runError
	if err == nil {
		t.Error("Expected error on retry of seeder2, got nil")
	}
	// seeder2 should have been called again (it was not marked due to previous error)
	if !seeder2.runCalled {
		t.Error("Expected seeder2 to run again on second CallOnce (it was not marked due to error)")
	}
}

func TestResetCallOnce(t *testing.T) {
	runner := NewRunner()

	seeder1 := &errorMockSeeder{signature: "seeder_1"}

	runner.Register([]contractsseeder.Seeder{seeder1})

	// First call
	_ = runner.CallOnce([]contractsseeder.Seeder{seeder1})

	// Reset
	runner.ResetCallOnce()

	// Reset runCalled flag
	seeder1.runCalled = false

	// Second call after reset - should run again
	err := runner.CallOnce([]contractsseeder.Seeder{seeder1})
	if err != nil {
		t.Errorf("Expected no error after reset, got %v", err)
	}
	if !seeder1.runCalled {
		t.Error("Expected seeder.Run() to be called after ResetCallOnce")
	}
}

func TestRunnerRegisterMultipleTimes(t *testing.T) {
	runner := NewRunner()

	seeder1 := &errorMockSeeder{signature: "seeder_1"}
	seeder2 := &errorMockSeeder{signature: "seeder_2"}

	runner.Register([]contractsseeder.Seeder{seeder1})
	runner.Register([]contractsseeder.Seeder{seeder2})

	seeders := runner.GetSeeders()
	if len(seeders) != 2 {
		t.Errorf("Expected 2 seeders after multiple Register calls, got %d", len(seeders))
	}
}
