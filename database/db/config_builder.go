package db

import (
	"fmt"
	"net/url"
	"strings"
)

// GetString implements config.Config interface for DBConfig.
func (c *DBConfig) GetString(path string, defaultValue ...any) string {
	switch path {
	case "database.default":
		if c.Default != "" {
			return c.Default
		}
		if len(defaultValue) > 0 {
			return defaultValue[0].(string)
		}
		return ""
	}

	// Handle connection-specific paths
	if strings.HasPrefix(path, "database.connections.") {
		parts := splitPath(path)
		if len(parts) >= 4 {
			connName := parts[2]
			field := parts[3]

			conn, ok := c.Connections[connName]
			if !ok && len(defaultValue) > 0 {
				return defaultValue[0].(string)
			}

			switch field {
			case "driver":
				return conn.Driver
			case "prefix":
				return conn.Prefix
			case "dsn":
				return conn.Dsn
			case "host":
				return conn.Host
			case "port":
				return fmt.Sprintf("%d", conn.Port)
			case "username":
				return conn.Username
			case "password":
				return conn.Password
			case "charset":
				return conn.Charset
			case "schema":
				if conn.Schema != "" {
					return conn.Schema
				}
				if len(defaultValue) > 0 {
					return defaultValue[0].(string)
				}
				return "public"
			case "loc":
				return conn.Loc
			case "sslmode":
				return conn.SSLMode
			case "timezone":
				return conn.Timezone
			case "database":
				return conn.Database
			}
		}
	}

	if len(defaultValue) > 0 {
		return defaultValue[0].(string)
	}
	return ""
}

// GetInt implements config.Config interface for DBConfig.
func (c *DBConfig) GetInt(path string, defaultValue ...any) int {
	switch path {
	case "database.pool.max_idle_conns":
		if c.Pool.MaxIdleConns > 0 {
			return c.Pool.MaxIdleConns
		}
		if len(defaultValue) > 0 {
			return defaultValue[0].(int)
		}
		return 10
	case "database.pool.max_open_conns":
		if c.Pool.MaxOpenConns > 0 {
			return c.Pool.MaxOpenConns
		}
		if len(defaultValue) > 0 {
			return defaultValue[0].(int)
		}
		return 100
	case "database.pool.conn_max_idletime":
		if c.Pool.ConnMaxIdleTime > 0 {
			return c.Pool.ConnMaxIdleTime
		}
		if len(defaultValue) > 0 {
			return defaultValue[0].(int)
		}
		return 3600
	case "database.pool.conn_max_lifetime":
		if c.Pool.ConnMaxLifetime > 0 {
			return c.Pool.ConnMaxLifetime
		}
		if len(defaultValue) > 0 {
			return defaultValue[0].(int)
		}
		return 3600
	case "database.pool.query_timeout":
		if c.Pool.QueryTimeout > 0 {
			return c.Pool.QueryTimeout
		}
		if len(defaultValue) > 0 {
			return defaultValue[0].(int)
		}
		return 30 // Default 30 seconds
	}

	// Handle connection-specific port
	if strings.HasPrefix(path, "database.connections.") {
		parts := splitPath(path)
		if len(parts) >= 4 {
			connName := parts[2]
			field := parts[3]

			conn, ok := c.Connections[connName]
			if ok && field == "port" {
				return conn.Port
			}
		}
	}

	if len(defaultValue) > 0 {
		return defaultValue[0].(int)
	}
	return 0
}

// GetBool implements config.Config interface for DBConfig.
func (c *DBConfig) GetBool(path string, defaultValue ...any) bool {
	switch path {
	case "app.debug":
		return c.Debug
	}

	// Handle connection-specific bool fields
	if strings.HasPrefix(path, "database.connections.") {
		parts := splitPath(path)
		if len(parts) >= 4 {
			connName := parts[2]
			field := parts[3]

			conn, ok := c.Connections[connName]
			if ok {
				switch field {
				case "singular":
					return conn.Singular
				case "no_lower_case":
					return conn.NoLowerCase
				}
			}
		}
	}

	if len(defaultValue) > 0 {
		return defaultValue[0].(bool)
	}
	return false
}

// Get implements config.Config interface for DBConfig.
func (c *DBConfig) Get(path string, defaultValue ...any) any {
	// Handle connection-specific maps
	if strings.HasPrefix(path, "database.connections.") {
		parts := splitPath(path)
		if len(parts) >= 4 {
			connName := parts[2]
			field := parts[3]

			conn, ok := c.Connections[connName]
			if !ok {
				if len(defaultValue) > 0 {
					return defaultValue[0]
				}
				return nil
			}

			switch field {
			case "read":
				// Return single connection as array for simplicity
				return []any{conn}
			case "write":
				return []any{conn}
			case "name_replacer":
				return conn.NameReplacer
			}
		}
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return nil
}

// Env implements config.Config interface for DBConfig (stub).
func (c *DBConfig) Env(envName string, defaultValue ...any) any {
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return nil
}

// Add implements config.Config interface for DBConfig (stub).
func (c *DBConfig) Add(name string, configuration any) {
	// Not implemented for standalone module
}

// splitPath splits a dot-separated path into its components.
func splitPath(path string) []string {
	var parts []string
	current := ""
	for _, ch := range path {
		if ch == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

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
	case "oracle":
		return b.buildOracleDSN()
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
	charset := b.config.Charset
	if charset == "" {
		charset = "utf8mb4"
	}
	params = append(params, "charset="+charset)
	// parseTime=true is required for the MySQL driver to return time.Time values
	// for DATETIME/TIMESTAMP columns instead of []uint8 byte slices.
	params = append(params, "parseTime=true")
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

// buildOracleDSN builds an Oracle DSN string.
func (b *ConfigBuilder) buildOracleDSN() (string, error) {
	// Oracle connection string format: oracle://user:pass@host:port/service
	return fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
		b.config.Username,
		b.config.Password,
		b.config.Host,
		b.config.Port,
		b.config.Database,
	), nil
}

// ReplicaConfig holds host/port/credentials for a single read or write replica.
type ReplicaConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
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
	Read         []ReplicaConfig // read replicas; if non-empty, SELECTs are routed here
	Write        []ReplicaConfig // write primaries; if non-empty, mutating queries go here
}

// String returns a string representation of ConnectionConfig with password masked.
func (c ConnectionConfig) String() string {
	return fmt.Sprintf("{Driver: %s, Host: %s, Port: %d, Database: %s, Username: %s, Password: ***}",
		c.Driver, c.Host, c.Port, c.Database, c.Username)
}

// Validate checks that the connection configuration has all required fields for the driver.
// If a DSN is provided directly, driver-specific field validation is skipped.
func (c *ConnectionConfig) Validate() error {
	if c.Driver == "" {
		return fmt.Errorf("driver is required")
	}
	if c.Dsn != "" {
		return nil
	}
	switch c.Driver {
	case "sqlite":
		// database path is optional; empty defaults to :memory:
		return nil
	case "mysql":
		if c.Host == "" {
			return fmt.Errorf("host is required for %s driver", c.Driver)
		}
		if c.Database == "" {
			return fmt.Errorf("database name is required for %s driver", c.Driver)
		}
		if c.Username == "" {
			return fmt.Errorf("username is required for %s driver", c.Driver)
		}
	case "postgres":
		if c.Host == "" {
			return fmt.Errorf("host is required for %s driver", c.Driver)
		}
		if c.Database == "" {
			return fmt.Errorf("database name is required for %s driver", c.Driver)
		}
		if c.Username == "" {
			return fmt.Errorf("username is required for %s driver", c.Driver)
		}
	case "sqlserver":
		if c.Host == "" {
			return fmt.Errorf("host is required for %s driver", c.Driver)
		}
		if c.Database == "" {
			return fmt.Errorf("database name is required for %s driver", c.Driver)
		}
	case "turso":
		if c.Database == "" {
			return fmt.Errorf("database URL is required for %s driver", c.Driver)
		}
	case "oracle":
		if c.Host == "" {
			return fmt.Errorf("host is required for %s driver", c.Driver)
		}
		if c.Database == "" {
			return fmt.Errorf("database/service name is required for %s driver", c.Driver)
		}
	default:
		return fmt.Errorf("unsupported driver: %s", c.Driver)
	}
	return nil
}

// PoolConfig represents connection pool configuration.
type PoolConfig struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int // seconds
	ConnMaxIdleTime int // seconds
	QueryTimeout    int // seconds (default: 30)
}

// DBConfig represents the database configuration.
type DBConfig struct {
	Default       string
	Connections   map[string]ConnectionConfig
	Pool          PoolConfig
	Debug         bool
	SlowThreshold int // slow query threshold in milliseconds (0 = disabled)
}

// Validate checks that the database configuration has all required fields.
// It verifies that at least one connection is configured, the default connection
// exists, and each connection's driver-specific requirements are met.
func (c *DBConfig) Validate() error {
	if len(c.Connections) == 0 {
		return fmt.Errorf("at least one database connection is required")
	}
	defaultConn := c.Default
	if defaultConn == "" {
		for name := range c.Connections {
			defaultConn = name
			break
		}
	}
	if _, ok := c.Connections[defaultConn]; !ok {
		return fmt.Errorf("default connection %q not found in connections", defaultConn)
	}
	for name, conn := range c.Connections {
		connCopy := conn
		if err := connCopy.Validate(); err != nil {
			return fmt.Errorf("connection %q: %w", name, err)
		}
	}
	return nil
}
