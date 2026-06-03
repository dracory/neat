package driver

import "fmt"

// PlaceholderFunc is a function that generates placeholders for a given index.
type PlaceholderFunc func(n int) string

// PlaceholderFuncs contains placeholder functions for different dialects.
var PlaceholderFuncs = map[string]PlaceholderFunc{
	"mysql":     mysqlPlaceholder,
	"oracle":    oraclePlaceholder,
	"postgres":  postgresPlaceholder,
	"sqlite":    sqlitePlaceholder,
	"sqlserver": sqlserverPlaceholder,
	"turso":     sqlitePlaceholder,
}

// mysqlPlaceholder returns MySQL-style placeholders (?).
func mysqlPlaceholder(n int) string {
	return "?"
}

// oraclePlaceholder returns Oracle-style placeholders (:1, :2, :3).
func oraclePlaceholder(n int) string {
	return fmt.Sprintf(":%d", n)
}

// postgresPlaceholder returns PostgreSQL-style placeholders ($1, $2, $3).
func postgresPlaceholder(n int) string {
	return fmt.Sprintf("$%d", n)
}

// sqlitePlaceholder returns SQLite-style placeholders (?).
func sqlitePlaceholder(n int) string {
	return "?"
}

// sqlserverPlaceholder returns SQL Server-style placeholders (@p1, @p2, @p3).
func sqlserverPlaceholder(n int) string {
	return fmt.Sprintf("@p%d", n)
}

// GetPlaceholderFunc returns the placeholder function for the given dialect.
func GetPlaceholderFunc(dialect string) PlaceholderFunc {
	if fn, ok := PlaceholderFuncs[dialect]; ok {
		return fn
	}
	// Default to MySQL-style placeholder
	return mysqlPlaceholder
}
