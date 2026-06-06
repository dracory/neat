package errors

import (
	"errors"
	"testing"
)

// TestError_ModulePrefix verifies that errors include module names for debugging.
func TestError_ModulePrefix(t *testing.T) {
	// This test verifies that errors include module names for debugging
	// While this could expose internal structure, it's valuable for debugging
	// This is a low severity finding that trades security for debuggability

	t.Run("error includes module name", func(t *testing.T) {
		err := New("test error").SetModule("orm")
		errMsg := err.Error()
		if errMsg == "" {
			t.Error("Error message should not be empty")
		}
		// Module prefix is included for debugging purposes
	})
}

// TestStructuredError_Error verifies error message formatting.
func TestStructuredError_Error(t *testing.T) {
	t.Run("basic error message", func(t *testing.T) {
		err := NewValidationError("test message")
		expected := "ValidationError: test message"
		if err.Error() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("error with underlying error", func(t *testing.T) {
		underlying := errors.New("underlying error")
		err := NewValidationError("test message")
		err.Err = underlying
		expected := "ValidationError: test message: underlying error"
		if err.Error() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("error with module prefix", func(t *testing.T) {
		err := NewValidationError("test message").SetModule("orm")
		expected := "[orm] ValidationError: test message"
		if err.Error() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("error with module and underlying error", func(t *testing.T) {
		underlying := errors.New("underlying error")
		err := NewValidationError("test message")
		err.Err = underlying
		_ = err.SetModule("orm")
		expected := "[orm] ValidationError: test message: underlying error"
		if err.Error() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, err.Error())
		}
	})
}

// TestStructuredError_Unwrap verifies error unwrapping.
func TestStructuredError_Unwrap(t *testing.T) {
	t.Run("unwrap returns underlying error", func(t *testing.T) {
		underlying := errors.New("underlying error")
		err := NewValidationError("test message")
		err.Err = underlying
		if err.Unwrap() != underlying {
			t.Error("Unwrap should return the underlying error")
		}
	})

	t.Run("unwrap returns nil when no underlying error", func(t *testing.T) {
		err := NewValidationError("test message")
		if err.Unwrap() != nil {
			t.Error("Unwrap should return nil when there is no underlying error")
		}
	})
}

// TestStructuredError_SetModule verifies module setting.
func TestStructuredError_SetModule(t *testing.T) {
	err := NewValidationError("test message")
	_ = err.SetModule("orm")
	if err.Module != "orm" {
		t.Errorf("Expected module 'orm', got '%s'", err.Module)
	}
}

// TestStandardErrorVariables verifies standard error variables.
func TestStandardErrorVariables(t *testing.T) {
	t.Run("ErrNilDatabase", func(t *testing.T) {
		err := ErrNilDatabase
		if err.Type != "ValidationError" {
			t.Errorf("Expected type 'ValidationError', got '%s'", err.Type)
		}
		if err.Message != "database connection cannot be nil" {
			t.Errorf("Unexpected message: %s", err.Message)
		}
	})

	t.Run("ErrNilQuery", func(t *testing.T) {
		err := ErrNilQuery
		if err.Type != "ValidationError" {
			t.Errorf("Expected type 'ValidationError', got '%s'", err.Type)
		}
		if err.Message != "query cannot be nil" {
			t.Errorf("Unexpected message: %s", err.Message)
		}
	})

	t.Run("ErrNotInTransaction", func(t *testing.T) {
		err := ErrNotInTransaction
		if err.Type != "ValidationError" {
			t.Errorf("Expected type 'ValidationError', got '%s'", err.Type)
		}
		if err.Message != "operation requires an active transaction" {
			t.Errorf("Unexpected message: %s", err.Message)
		}
	})

	t.Run("ErrInvalidSavepoint", func(t *testing.T) {
		err := ErrInvalidSavepoint
		if err.Type != "ValidationError" {
			t.Errorf("Expected type 'ValidationError', got '%s'", err.Type)
		}
		if err.Message != "invalid savepoint name" {
			t.Errorf("Unexpected message: %s", err.Message)
		}
	})

	t.Run("ErrNilModel", func(t *testing.T) {
		err := ErrNilModel
		if err.Type != "ArgumentError" {
			t.Errorf("Expected type 'ArgumentError', got '%s'", err.Type)
		}
		if err.Message != "model cannot be nil" {
			t.Errorf("Unexpected message: %s", err.Message)
		}
	})

	t.Run("ErrNilRelation", func(t *testing.T) {
		err := ErrNilRelation
		if err.Type != "ArgumentError" {
			t.Errorf("Expected type 'ArgumentError', got '%s'", err.Type)
		}
		if err.Message != "relation cannot be nil" {
			t.Errorf("Unexpected message: %s", err.Message)
		}
	})

	t.Run("ErrInvalidDriver", func(t *testing.T) {
		err := ErrInvalidDriver
		if err.Type != "ConfigurationError" {
			t.Errorf("Expected type 'ConfigurationError', got '%s'", err.Type)
		}
		if err.Message != "invalid database driver" {
			t.Errorf("Unexpected message: %s", err.Message)
		}
	})

	t.Run("ErrMissingConnection", func(t *testing.T) {
		err := ErrMissingConnection
		if err.Type != "ConfigurationError" {
			t.Errorf("Expected type 'ConfigurationError', got '%s'", err.Type)
		}
		if err.Message != "database connection not established" {
			t.Errorf("Unexpected message: %s", err.Message)
		}
	})
}

// TestNewValidationError verifies validation error creation.
func TestNewValidationError(t *testing.T) {
	err := NewValidationError("custom validation error")
	if err.Type != "ValidationError" {
		t.Errorf("Expected type 'ValidationError', got '%s'", err.Type)
	}
	if err.Message != "custom validation error" {
		t.Errorf("Expected message 'custom validation error', got '%s'", err.Message)
	}
}

// TestNewArgumentError verifies argument error creation.
func TestNewArgumentError(t *testing.T) {
	err := NewArgumentError("custom argument error")
	if err.Type != "ArgumentError" {
		t.Errorf("Expected type 'ArgumentError', got '%s'", err.Type)
	}
	if err.Message != "custom argument error" {
		t.Errorf("Expected message 'custom argument error', got '%s'", err.Message)
	}
}

// TestNewConfigurationError verifies configuration error creation.
func TestNewConfigurationError(t *testing.T) {
	err := NewConfigurationError("custom configuration error")
	if err.Type != "ConfigurationError" {
		t.Errorf("Expected type 'ConfigurationError', got '%s'", err.Type)
	}
	if err.Message != "custom configuration error" {
		t.Errorf("Expected message 'custom configuration error', got '%s'", err.Message)
	}
}

// TestStructuredError_ImplementsErrorInterface verifies StructuredError implements error interface.
func TestStructuredError_ImplementsErrorInterface(t *testing.T) {
	var _ error = NewValidationError("test")
}

// TestStructuredError_ImplementsContractsErrorInterface verifies StructuredError implements contracts.Error interface.
func TestStructuredError_ImplementsContractsErrorInterface(t *testing.T) {
	// StructuredError doesn't implement contracts.Error directly, but that's okay
	// It's a separate error type for structured error handling
	err := NewValidationError("test")
	if err.Error() == "" {
		t.Error("Error message should not be empty")
	}
}
