package migrator

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/dracory/neat/contracts/config"
	contractsorm "github.com/dracory/neat/contracts/database/orm"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/contracts/migration"
)

// Repository manages migration records in the database.
type Repository struct {
	config config.Config
	orm    contractsorm.Orm
	schema contractsschema.Schema
	table  string
}

// NewRepository creates a new Repository instance.
func NewRepository(config config.Config, orm contractsorm.Orm, schema contractsschema.Schema) *Repository {
	table := config.GetString("database.migrations.table", "migrations")

	// Validate table name to prevent SQL injection
	// Security: Ensure table name is a simple identifier without SQL injection vectors
	if !isValidMigrationTableName(table) {
		panic(fmt.Sprintf("invalid migration table name: '%s' - must contain only letters, numbers, and underscores, and cannot be an SQL keyword", table))
	}

	return &Repository{
		config: config,
		orm:    orm,
		schema: schema,
		table:  table,
	}
}

func (r *Repository) CreateRepository() error {
	// Check if table already exists
	if r.RepositoryExists() {
		// Upgrade schema if needed
		return r.upgradeRepositorySchema()
	}

	// Create migrations table using schema builder
	return r.schema.Create(r.table, func(table contractsschema.Blueprint) {
		table.ID()
		table.String("migration", 255)
		table.Integer("batch")
		table.Text("description").Nullable()
		table.DateTime("started_at").Nullable()
		table.DateTime("completed_at").Nullable()
		table.Timestamps()
	})
}

// getSchema returns a schema instance
func (r *Repository) getSchema() contractsschema.Schema {
	return r.schema
}

func (r *Repository) Delete(migrationName string) error {
	query := r.orm.Query()
	if query == nil {
		return fmt.Errorf("query not initialized")
	}

	deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE migration = ?", r.table)
	_, err := query.Exec(deleteSQL, migrationName)
	return err
}

func (r *Repository) DeleteRepository() error {
	query := r.orm.Query()
	if query == nil {
		return fmt.Errorf("query not initialized")
	}

	dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s", r.table)
	_, err := query.Exec(dropSQL)
	return err
}

func (r *Repository) GetLast() ([]migration.File, error) {
	query := r.orm.Query()
	if query == nil {
		return nil, fmt.Errorf("query not initialized")
	}

	// Get the last batch number
	var lastBatch sql.NullInt64
	batchSQL := fmt.Sprintf("SELECT MAX(batch) FROM %s", r.table)
	err := query.Raw(batchSQL).Scan(&lastBatch)
	if err != nil {
		return nil, err
	}

	if !lastBatch.Valid {
		return []migration.File{}, nil
	}

	return r.GetMigrationsByBatch(int(lastBatch.Int64))
}

func (r *Repository) GetMigrations() ([]migration.File, error) {
	query := r.orm.Query()
	if query == nil {
		return nil, fmt.Errorf("query not initialized")
	}

	var files []migration.File
	selectSQL := fmt.Sprintf("SELECT id, migration, batch FROM %s ORDER BY id ASC", r.table)
	err := query.Raw(selectSQL).Scan(&files)
	return files, err
}

func (r *Repository) GetMigrationsByBatch(batch int) ([]migration.File, error) {
	query := r.orm.Query()
	if query == nil {
		return nil, fmt.Errorf("query not initialized")
	}

	var files []migration.File
	selectSQL := fmt.Sprintf("SELECT id, migration, batch FROM %s WHERE batch = ? ORDER BY id ASC", r.table)
	err := query.Raw(selectSQL, batch).Scan(&files)
	return files, err
}

func (r *Repository) GetMigrationsByStep(steps int) ([]migration.File, error) {
	query := r.orm.Query()
	if query == nil {
		return nil, fmt.Errorf("query not initialized")
	}

	var files []migration.File
	selectSQL := fmt.Sprintf("SELECT id, migration, batch FROM %s ORDER BY id DESC LIMIT ?", r.table)
	err := query.Raw(selectSQL, steps).Scan(&files)
	return files, err
}

func (r *Repository) GetNextBatchNumber() (int, error) {
	query := r.orm.Query()
	if query == nil {
		return 0, fmt.Errorf("query not initialized")
	}

	var lastBatch sql.NullInt64
	batchSQL := fmt.Sprintf("SELECT MAX(batch) FROM %s", r.table)
	err := query.Raw(batchSQL).Scan(&lastBatch)
	if err != nil {
		return 0, err
	}

	if !lastBatch.Valid {
		return 1, nil
	}

	return int(lastBatch.Int64) + 1, nil
}

func (r *Repository) GetRan() ([]string, error) {
	query := r.orm.Query()
	if query == nil {
		return nil, fmt.Errorf("query not initialized")
	}

	var migrations []string
	selectSQL := fmt.Sprintf("SELECT migration FROM %s ORDER BY id ASC", r.table)
	err := query.Raw(selectSQL).Scan(&migrations)
	return migrations, err
}

func (r *Repository) Log(migrationName string, batch int, description string, startedAt, completedAt time.Time) error {
	query := r.orm.Query()
	if query == nil {
		return fmt.Errorf("query not initialized")
	}

	insertSQL := fmt.Sprintf("INSERT INTO %s (migration, batch, description, started_at, completed_at, created_at) VALUES (?, ?, ?, ?, ?, ?)", r.table)
	_, err := query.Exec(insertSQL, migrationName, batch, description, startedAt, completedAt, time.Now())
	return err
}

func (r *Repository) RepositoryExists() bool {
	// Use the same DB connection for consistency
	databaseConn, err := r.orm.DB()
	if err != nil {
		return false
	}

	// Check if table exists by trying to query it
	var count int

	// For SQLite, use different approach
	driver := r.config.GetString(fmt.Sprintf("database.connections.%s.driver", r.orm.Name()))
	if driver == "sqlite" || driver == "turso" {
		countSQL := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
		err := databaseConn.QueryRow(countSQL, r.table).Scan(&count)
		return err == nil && count > 0
	}

	// For other databases, use information_schema
	countSQL := "SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ?"
	err = databaseConn.QueryRow(countSQL, r.table).Scan(&count)
	return err == nil && count > 0
}

// upgradeRepositorySchema adds new columns to existing migration table if they don't exist
func (r *Repository) upgradeRepositorySchema() error {
	// Use schema builder to add columns
	// Check if columns exist first to avoid errors
	if !r.columnExists("description") {
		if err := r.schema.Table(r.table, func(table contractsschema.Blueprint) {
			table.Text("description").Nullable()
		}); err != nil {
			// Ignore error if column already exists (race condition)
		}
	}

	if !r.columnExists("started_at") {
		if err := r.schema.Table(r.table, func(table contractsschema.Blueprint) {
			table.DateTime("started_at").Nullable()
		}); err != nil {
			// Ignore error if column already exists
		}
	}

	if !r.columnExists("completed_at") {
		if err := r.schema.Table(r.table, func(table contractsschema.Blueprint) {
			table.DateTime("completed_at").Nullable()
		}); err != nil {
			// Ignore error if column already exists
		}
	}

	return nil
}

// columnExists checks if a column exists in the migration table
func (r *Repository) columnExists(columnName string) bool {
	databaseConn, err := r.orm.DB()
	if err != nil {
		return false
	}

	driver := r.config.GetString(fmt.Sprintf("database.connections.%s.driver", r.orm.Name()))

	switch driver {
	case "sqlite", "turso":
		// SQLite: use PRAGMA table_info
		rows, err := databaseConn.Query(fmt.Sprintf("PRAGMA table_info(%s)", r.table))
		if err != nil {
			return false
		}
		defer rows.Close()

		for rows.Next() {
			var cid int
			var name string
			var ctype string
			var notnull int
			var dfltValue interface{}
			var pk int
			if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
				continue
			}
			if name == columnName {
				return true
			}
		}
		return false

	case "postgres":
		var count int
		query := `SELECT COUNT(*) FROM information_schema.columns 
		          WHERE table_name = $1 AND column_name = $2`
		err := databaseConn.QueryRow(query, r.table, columnName).Scan(&count)
		return err == nil && count > 0

	case "mysql":
		var count int
		query := `SELECT COUNT(*) FROM information_schema.columns 
		          WHERE table_name = ? AND column_name = ?`
		err := databaseConn.QueryRow(query, r.table, columnName).Scan(&count)
		return err == nil && count > 0

	case "sqlserver":
		var count int
		query := `SELECT COUNT(*) FROM information_schema.columns 
		          WHERE table_name = @p1 AND column_name = @p2`
		err := databaseConn.QueryRow(query, r.table, columnName).Scan(&count)
		return err == nil && count > 0

	default:
		return false
	}
}

// GetHistory returns the migration execution history from the database
func (r *Repository) GetHistory() ([]migration.File, error) {
	query := r.orm.Query()
	if query == nil {
		return nil, fmt.Errorf("query not initialized")
	}

	var files []migration.File
	selectSQL := fmt.Sprintf("SELECT id, migration, batch, description, started_at, completed_at FROM %s ORDER BY id ASC", r.table)
	err := query.Raw(selectSQL).Scan(&files)
	return files, err
}

// isValidMigrationTableName validates that a table name is safe to use in SQL queries.
// It checks for:
// - Only alphanumeric characters and underscores
// - Does not start with a number
// - Not empty
// - Not an SQL keyword
// This prevents SQL injection attacks through malicious table names.
func isValidMigrationTableName(tableName string) bool {
	if tableName == "" {
		return false
	}

	// Must start with a letter or underscore (not a number)
	first := tableName[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}

	// Must contain only alphanumeric characters and underscores
	for _, char := range tableName {
		isLetter := (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
		isDigit := char >= '0' && char <= '9'
		isUnderscore := char == '_'
		if !isLetter && !isDigit && !isUnderscore {
			return false
		}
	}

	// Reject SQL keywords to prevent injection attempts
	upperTableName := strings.ToUpper(tableName)
	sqlKeywords := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE",
		"ALTER", "TRUNCATE", "REPLACE", "MERGE", "UNION", "EXCEPT",
		"INTERSECT", "WHERE", "FROM", "JOIN", "INNER", "OUTER",
		"LEFT", "RIGHT", "FULL", "CROSS", "ON", "USING", "AND",
		"OR", "NOT", "IN", "EXISTS", "BETWEEN", "LIKE", "IS",
		"NULL", "TRUE", "FALSE", "CASE", "WHEN", "THEN", "ELSE",
		"END", "GROUP", "HAVING", "ORDER", "BY", "LIMIT", "OFFSET",
		"DISTINCT", "ALL", "AS", "TABLE", "VIEW", "INDEX", "TRIGGER",
		"PROCEDURE", "FUNCTION", "DATABASE", "SCHEMA", "GRANT", "REVOKE",
		"EXEC", "EXECUTE",
	}

	for _, keyword := range sqlKeywords {
		if upperTableName == keyword {
			return false
		}
	}

	// Reasonable length limit (most databases support 64 chars, we allow up to 128)
	if len(tableName) > 128 {
		return false
	}

	return true
}
