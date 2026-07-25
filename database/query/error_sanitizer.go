package query

import (
	"context"
	"errors"
	"fmt"
	"slices"
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
	lowerMsg := strings.ToLower(errMsg)
	keywords := []string{"sql", "query", "syntax", "column", "table", "constraint", "foreign key", "primary key", "duplicate", "unique", "relation", "schema"}
	if slices.ContainsFunc(keywords, func(k string) bool { return strings.Contains(lowerMsg, k) }) {
		return fmt.Errorf("database operation failed: %w", err)
	}

	return err
}
