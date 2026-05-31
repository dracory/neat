package query

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// isProduction returns true when the query is not in debug mode.
// Safe to call when dbConfig is nil (defaults to production/non-debug).
func (q *Query) isProduction() bool {
	if q.dbConfig == nil {
		return true
	}
	return !q.dbConfig.Debug
}

// sanitizeError removes SQL details from error messages in production mode.
// Call with q.isProduction() to prevent SQL detail leakage to callers.
// Integrated into query execution error return paths (Scan, Exec, Create, Update, Delete, Restore, ForceDelete).
// See security review Finding #9.
func sanitizeError(err error, isProduction bool) error {
	if err == nil || !isProduction {
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
