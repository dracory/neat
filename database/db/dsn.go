package db

import (
	"fmt"
	"net/url"
	"strings"
)

const maxDSNLength = 4096

// ParseDSN parses a DSN string and extracts connection information.
func ParseDSN(dsn string) (ConnectionConfig, error) {
	if dsn == "" {
		return ConnectionConfig{}, fmt.Errorf("DSN cannot be empty")
	}
	if len(dsn) > maxDSNLength {
		return ConnectionConfig{}, fmt.Errorf("DSN exceeds maximum length of %d characters", maxDSNLength)
	}
	config := ConnectionConfig{Dsn: dsn}

	// Try to detect driver from DSN format
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		return parsePostgresDSN(dsn)
	}
	if strings.HasPrefix(dsn, "mysql://") {
		return parseMySQLDSN(dsn)
	}
	if strings.HasPrefix(dsn, "sqlite://") || strings.HasPrefix(dsn, "file:") {
		return parseSQLiteDSN(dsn)
	}
	if strings.HasPrefix(dsn, "sqlserver://") {
		return parseSQLServerDSN(dsn)
	}
	if strings.HasPrefix(dsn, "libsql://") {
		return parseTursoDSN(dsn)
	}

	// If no recognized prefix, return as-is
	return config, nil
}

// parsePostgresDSN parses a PostgreSQL DSN string.
func parsePostgresDSN(dsn string) (ConnectionConfig, error) {
	config := ConnectionConfig{Driver: "postgres", Dsn: dsn}

	// Parse URL
	u, err := url.Parse(dsn)
	if err != nil {
		return config, err
	}

	config.Host = u.Hostname()
	config.Database = strings.TrimPrefix(u.Path, "/")
	config.Username = u.User.Username()
	if password, ok := u.User.Password(); ok {
		config.Password = password
	}

	// Parse query parameters
	query := u.Query()
	if schema := query.Get("search_path"); schema != "" {
		config.Schema = schema
	}
	if sslmode := query.Get("sslmode"); sslmode != "" {
		config.SSLMode = sslmode
	}
	if timezone := query.Get("TimeZone"); timezone != "" {
		config.Timezone = timezone
	}

	return config, nil
}

// parseMySQLDSN parses a MySQL DSN string.
func parseMySQLDSN(dsn string) (ConnectionConfig, error) {
	config := ConnectionConfig{Driver: "mysql", Dsn: dsn}

	// Parse URL
	u, err := url.Parse(dsn)
	if err != nil {
		return config, err
	}

	config.Host = u.Hostname()
	config.Database = strings.TrimPrefix(u.Path, "/")
	config.Username = u.User.Username()
	if password, ok := u.User.Password(); ok {
		config.Password = password
	}

	// Parse query parameters
	query := u.Query()
	if charset := query.Get("charset"); charset != "" {
		config.Charset = charset
	}
	if loc := query.Get("loc"); loc != "" {
		config.Loc = loc
	}

	return config, nil
}

// parseSQLiteDSN parses a SQLite DSN string.
func parseSQLiteDSN(dsn string) (ConnectionConfig, error) {
	config := ConnectionConfig{Driver: "sqlite", Dsn: dsn}

	// Extract database path
	if strings.HasPrefix(dsn, "sqlite://") {
		config.Database = strings.TrimPrefix(dsn, "sqlite://")
	} else if strings.HasPrefix(dsn, "file:") {
		config.Database = strings.TrimPrefix(dsn, "file:")
	} else {
		config.Database = dsn
	}

	return config, nil
}

// parseSQLServerDSN parses a SQL Server DSN string.
func parseSQLServerDSN(dsn string) (ConnectionConfig, error) {
	config := ConnectionConfig{Driver: "sqlserver", Dsn: dsn}

	// Parse URL
	u, err := url.Parse(dsn)
	if err != nil {
		return config, err
	}

	config.Host = u.Hostname()
	config.Username = u.User.Username()
	if password, ok := u.User.Password(); ok {
		config.Password = password
	}

	// Parse query parameters
	query := u.Query()
	if database := query.Get("database"); database != "" {
		config.Database = database
	}

	return config, nil
}

// parseTursoDSN parses a Turso DSN string.
func parseTursoDSN(dsn string) (ConnectionConfig, error) {
	config := ConnectionConfig{Driver: "turso", Dsn: dsn}

	// Parse URL
	u, err := url.Parse(dsn)
	if err != nil {
		return config, err
	}

	config.Database = u.Hostname()

	// Parse query parameters
	query := u.Query()
	if authToken := query.Get("authToken"); authToken != "" {
		config.Password = authToken
	}

	return config, nil
}

// BuildDSN builds a DSN string from a ConnectionConfig.
func BuildDSN(config ConnectionConfig) (string, error) {
	builder := NewConfigBuilder(config)
	return builder.BuildDSN()
}
