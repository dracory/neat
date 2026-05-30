package common

import (
	"time"

	"github.com/dracory/neat/contracts/log"
)

// TriggerSlowQueryWarning manually triggers a slow query warning for testing.
// This simulates a query that took the specified duration.
func TriggerSlowQueryWarning(logger log.Log, sqlStr string, bindings []any, duration time.Duration) {
	elapsed := float64(duration.Milliseconds())
	logger.Warningf("[slow query %.1fms] %s %v", elapsed, sqlStr, bindings)
}
