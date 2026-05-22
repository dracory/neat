package neat

import (
	"fmt"
	"strings"
	"time"

	contractsdb "github.com/dracory/neat/contracts/database"
	"github.com/dracory/neat/database"
	"github.com/dracory/neat/database/db"
)

// DBConfig holds the database configuration for the standalone module.
type DBConfig struct {
	// Default connection name
	Default string
	// Connection configurations
	Connections map[string]ConnectionConfig
	// Migration configuration
	Migrations MigrationConfig
	// Pool configuration
	Pool PoolConfig
	// Debug mode
	Debug bool
	// Slow query threshold in milliseconds
	SlowThreshold int
}

// ConnectionConfig holds configuration for a single database connection.
type ConnectionConfig struct {
	Driver       string // "postgres", "mysql", "sqlite", "sqlserver", "turso"
	Dsn          string
	Host         string
	Port         int
	Database     string
	Username     string
	Password     string
	Charset      string
	Schema       string // postgres only
	SSLMode      string // postgres only
	Loc          string // mysql only
	Timezone     string // postgres only
	Prefix       string
	Singular     bool
	NoLowerCase  bool
	NameReplacer any
	Read         []contractsdb.Config
	Write        []contractsdb.Config
}

// MigrationConfig holds migration configuration.
type MigrationConfig struct {
	Driver string // "sql" or "orm"
	Table  string // default: "migrations"
}

// PoolConfig holds connection pool configuration.
type PoolConfig struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

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
	case "database.migrations.driver":
		return c.Migrations.Driver
	case "database.migrations.table":
		if c.Migrations.Table != "" {
			return c.Migrations.Table
		}
		return "migrations"
	case "app.env":
		return "local" // default for standalone module
	case "app.debug":
		if c.Debug {
			return "true"
		}
		return "false"
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
	case "database.slow_threshold":
		if c.SlowThreshold > 0 {
			return c.SlowThreshold
		}
		if len(defaultValue) > 0 {
			return defaultValue[0].(int)
		}
		return 200
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
			return int(c.Pool.ConnMaxIdleTime.Seconds())
		}
		if len(defaultValue) > 0 {
			return defaultValue[0].(int)
		}
		return 3600
	case "database.pool.conn_max_lifetime":
		if c.Pool.ConnMaxLifetime > 0 {
			return int(c.Pool.ConnMaxLifetime.Seconds())
		}
		if len(defaultValue) > 0 {
			return defaultValue[0].(int)
		}
		return 3600
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

// New creates a new Database instance from a DBConfig.
func New(cfg DBConfig, opts ...database.Option) (*database.Database, error) {
	// Convert neat.DBConfig to database.db.DBConfig
	dbConfig := db.DBConfig{
		Default:     cfg.Default,
		Connections: make(map[string]db.ConnectionConfig),
		Pool: db.PoolConfig{
			MaxIdleConns:    cfg.Pool.MaxIdleConns,
			MaxOpenConns:    cfg.Pool.MaxOpenConns,
			ConnMaxLifetime: int(cfg.Pool.ConnMaxLifetime.Seconds()),
			ConnMaxIdleTime: int(cfg.Pool.ConnMaxIdleTime.Seconds()),
		},
		Debug:         cfg.Debug,
		SlowThreshold: cfg.SlowThreshold,
	}

	for name, conn := range cfg.Connections {
		dbConn := db.ConnectionConfig{
			Driver:       conn.Driver,
			Dsn:          conn.Dsn,
			Host:         conn.Host,
			Port:         conn.Port,
			Database:     conn.Database,
			Username:     conn.Username,
			Password:     conn.Password,
			Charset:      conn.Charset,
			Schema:       conn.Schema,
			SSLMode:      conn.SSLMode,
			Loc:          conn.Loc,
			Timezone:     conn.Timezone,
			Prefix:       conn.Prefix,
			Singular:     conn.Singular,
			NoLowerCase:  conn.NoLowerCase,
			NameReplacer: conn.NameReplacer,
		}
		for _, r := range conn.Read {
			dbConn.Read = append(dbConn.Read, db.ReplicaConfig{
				Host: r.Host, Port: r.Port, Database: r.Database,
				Username: r.Username, Password: r.Password,
			})
		}
		for _, w := range conn.Write {
			dbConn.Write = append(dbConn.Write, db.ReplicaConfig{
				Host: w.Host, Port: w.Port, Database: w.Database,
				Username: w.Username, Password: w.Password,
			})
		}
		dbConfig.Connections[name] = dbConn
	}

	return database.New(dbConfig, opts...)
}

// NewFromDSN creates a new Database instance from a DSN string.
func NewFromDSN(dsn string, opts ...database.Option) (*database.Database, error) {
	return database.NewFromDSN(dsn, opts...)
}
