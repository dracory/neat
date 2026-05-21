package db

import (
	"fmt"
	"net/url"
	"strings"
)

// ConfigBuilder builds DSN strings from connection configuration.
type ConfigBuilder struct {
	config ConnectionConfig
}

// NewConfigBuilder creates a new ConfigBuilder.
func NewConfigBuilder(config ConnectionConfig) *ConfigBuilder {
	return &ConfigBuilder{config: config}
}

// BuildDSN builds a DSN string from the connection configuration.
func (b *ConfigBuilder) BuildDSN() (string, error) {
	// If DSN is already provided, use it
	if b.config.Dsn != "" {
		return b.config.Dsn, nil
	}

	switch b.config.Driver {
	case "mysql":
		return b.buildMySQLDSN()
	case "postgres":
		return b.buildPostgresDSN()
	case "sqlite":
		return b.buildSQLiteDSN()
	case "sqlserver":
		return b.buildSQLServerDSN()
	case "turso":
		return b.buildTursoDSN()
	default:
		return "", fmt.Errorf("unsupported driver: %s", b.config.Driver)
	}
}

// buildMySQLDSN builds a MySQL DSN string.
func (b *ConfigBuilder) buildMySQLDSN() (string, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		b.config.Username,
		b.config.Password,
		b.config.Host,
		b.config.Port,
		b.config.Database,
	)

	params := []string{}
	if b.config.Charset != "" {
		params = append(params, "charset="+b.config.Charset)
	}
	if b.config.Loc != "" {
		params = append(params, "loc="+url.QueryEscape(b.config.Loc))
	}

	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}

	return dsn, nil
}

// buildPostgresDSN builds a PostgreSQL DSN string.
func (b *ConfigBuilder) buildPostgresDSN() (string, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		b.config.Host,
		b.config.Port,
		b.config.Username,
		b.config.Password,
		b.config.Database,
	)

	if b.config.Schema != "" {
		dsn += " search_path=" + b.config.Schema
	}
	if b.config.SSLMode != "" {
		dsn += " sslmode=" + b.config.SSLMode
	}
	if b.config.Timezone != "" {
		dsn += " TimeZone=" + b.config.Timezone
	}

	return dsn, nil
}

// buildSQLiteDSN builds a SQLite DSN string.
func (b *ConfigBuilder) buildSQLiteDSN() (string, error) {
	// SQLite DSN is just the file path
	if b.config.Database == "" {
		return ":memory:", nil
	}
	return b.config.Database, nil
}

// buildSQLServerDSN builds a SQL Server DSN string.
func (b *ConfigBuilder) buildSQLServerDSN() (string, error) {
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		b.config.Username,
		b.config.Password,
		b.config.Host,
		b.config.Port,
		b.config.Database,
	)
	return dsn, nil
}

// buildTursoDSN builds a Turso DSN string.
func (b *ConfigBuilder) buildTursoDSN() (string, error) {
	// Turso uses libsql:// or file:// format
	if strings.HasPrefix(b.config.Database, "libsql://") || strings.HasPrefix(b.config.Database, "file://") {
		return b.config.Database, nil
	}

	// If auth token is provided, add it
	if b.config.Password != "" {
		return fmt.Sprintf("libsql://%s?authToken=%s", b.config.Database, b.config.Password), nil
	}

	return fmt.Sprintf("libsql://%s", b.config.Database), nil
}

// ConnectionConfig represents a database connection configuration.
type ConnectionConfig struct {
	Driver       string
	Dsn          string
	Host         string
	Port         int
	Database     string
	Username     string
	Password     string
	Charset      string
	Schema       string
	SSLMode      string
	Loc          string
	Timezone     string
	Prefix       string
	Singular     bool
	NoLowerCase  bool
	NameReplacer any
}

// PoolConfig represents connection pool configuration.
type PoolConfig struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int // seconds
	ConnMaxIdleTime int // seconds
}

// DBConfig represents the database configuration.
type DBConfig struct {
	Default     string
	Connections map[string]ConnectionConfig
	Pool        PoolConfig
	Debug       bool
}
