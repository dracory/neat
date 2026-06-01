package migration

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/dracory/neat/contracts/config"
	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/contracts/migration"
)

// Repository manages migration records in the database.
type Repository struct {
	config config.Config
	orm    contractsorm.Orm
	table  string
}

// NewRepository creates a new Repository instance.
func NewRepository(config config.Config, orm contractsorm.Orm) *Repository {
	table := config.GetString("database.migrations.table", "migrations")
	return &Repository{
		config: config,
		orm:    orm,
		table:  table,
	}
}

func (r *Repository) CreateRepository() error {
	query := r.orm.Query()
	if query == nil {
		return fmt.Errorf("query not initialized")
	}

	// Check if table already exists
	if r.RepositoryExists() {
		return nil
	}

	// Create migrations table
	createSQL := fmt.Sprintf(`
		CREATE TABLE %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			migration VARCHAR(255) NOT NULL,
			batch INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`, r.table)

	// Adjust SQL for different databases
	driver := r.config.GetString(fmt.Sprintf("database.connections.%s.driver", r.orm.Name()))
	switch driver {
	case "postgres":
		createSQL = fmt.Sprintf(`
			CREATE TABLE %s (
				id SERIAL PRIMARY KEY,
				migration VARCHAR(255) NOT NULL,
				batch INTEGER NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.table)
	case "mysql":
		createSQL = fmt.Sprintf(`
			CREATE TABLE %s (
				id INT AUTO_INCREMENT PRIMARY KEY,
				migration VARCHAR(255) NOT NULL,
				batch INT NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.table)
	case "sqlserver":
		createSQL = fmt.Sprintf(`
			CREATE TABLE %s (
				id INT IDENTITY(1,1) PRIMARY KEY,
				migration NVARCHAR(255) NOT NULL,
				batch INT NOT NULL,
				created_at DATETIME DEFAULT GETDATE()
			)
		`, r.table)
	}

	_, err := query.Exec(createSQL)
	return err
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

func (r *Repository) Log(migrationName string, batch int) error {
	query := r.orm.Query()
	if query == nil {
		return fmt.Errorf("query not initialized")
	}

	insertSQL := fmt.Sprintf("INSERT INTO %s (migration, batch, created_at) VALUES (?, ?, ?)", r.table)
	_, err := query.Exec(insertSQL, migrationName, batch, time.Now())
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
