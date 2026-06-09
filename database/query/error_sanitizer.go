package query

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// sanitizeError removes SQL details from error messages in production mode.
// Logs full errors when debug is enabled.
// Integrated into query execution error return paths (Scan, Exec, Create, Update, Delete, Restore, ForceDelete).
// See security review Finding #9.
func (q *Query) sanitizeError(err error) error {
	if err == nil {
		return err
	}

	if q.IsDebug() {
		// Log full error when debug is enabled
		if q.log != nil {
			q.log.Errorf("Database error: %v", err)
		}
		return err
	}

	// Never suppress context errors — callers depend on errors.Is(err, context.Canceled) etc.
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}

	// In production, return generic error messages
	errMsg := err.Error()

	// Check if error contains SQL details
	if strings.Contains(strings.ToLower(errMsg), "sql") ||
		strings.Contains(strings.ToLower(errMsg), "query") ||
		strings.Contains(strings.ToLower(errMsg), "syntax") {
		return fmt.Errorf("database operation failed")
	}

	return err
}
