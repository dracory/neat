package query

import (
	"context"
	"errors"
	"slices"
	"strings"
)

// sanitizeError removes SQL details from error messages in production mode.
// Logs full errors when debug is enabled.
// Integrated into query execution error return paths (Scan, Exec, Create, Update, Delete, Restore, ForceDelete).
// See docs/security-review.md Finding #4.
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
		return &redactedError{inner: err}
	}

	return err
}

// redactedError presents a generic message via Error() while keeping the
// original error reachable through Unwrap() so errors.Is/errors.As still
// work. Unlike fmt.Errorf("...: %w", err), the wrapped error's message text
// is NOT included in Error() output — that is what actually prevents SQL
// detail leakage.
type redactedError struct {
	inner error
}

func (e *redactedError) Error() string { return "database operation failed" }
func (e *redactedError) Unwrap() error { return e.inner }
