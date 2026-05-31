package query

import (
	"fmt"
	"strings"
)

// sanitizeError removes SQL details from error messages in production mode.
// Call with isProduction=!dbConfig.Debug to prevent SQL detail leakage to callers.
// TODO: Wire this into query execution error return paths (query.go Scan/Exec methods)
// by passing !q.dbConfig.Debug as isProduction. See security review Finding #9.
func sanitizeError(err error, isProduction bool) error {
	if err == nil || !isProduction {
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
