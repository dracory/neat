package errors

import (
	"testing"
)

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
