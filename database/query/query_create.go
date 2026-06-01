package query

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"github.com/dracory/neat/database/observer"
)

// Create inserts a new record into the database.
func (q *Query) Create(value any) error {
	if q.buildError != nil {
		return q.buildError
	}
	// Fire Creating event if not disabled
	if !q.withoutEvents {
		attributes := observer.ExtractModelAttributes(value)
		if err := q.dispatcher.DispatchCreating(q.ctx, value, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return fmt.Errorf("creating event error: %w", err)
		}
	}

	// Build INSERT query
	builder := NewBuilder(q)
	sqlStr, args := builder.BuildInsert(value)
	if sqlStr == "" {
		return fmt.Errorf("failed to build INSERT query")
	}

	// Postgres: use RETURNING id to get inserted ID
	// SQL Server: uses OUTPUT clause (already added in BuildInsert)
	isPostgres := q.driver != nil && q.driver.Dialect() == "postgres"
	isSQLServer := q.driver != nil && q.driver.Dialect() == "sqlserver"
	if isPostgres {
		sqlStr = sqlStr + " RETURNING id"
	}

	// Execute query
	var result sql.Result
	var err error
	start := time.Now()
	if q.tx != nil {
		if isPostgres || isSQLServer {
			// For PostgreSQL with RETURNING or SQL Server with OUTPUT, use Query instead of Exec
			var rows *sql.Rows
			rows, err = q.tx.QueryContext(q.ctx, sqlStr, args...)
			if err != nil {
				return sanitizeError(fmt.Errorf("failed to execute INSERT query: %w", err), q.isProduction())
			}
			defer func() { _ = rows.Close() }()

			// Handle bulk insert by setting IDs for each element
			v := reflect.ValueOf(value)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}
			if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
				// Bulk insert: scan all returned IDs
				i := 0
				for rows.Next() {
					var lastID int64
					if err := rows.Scan(&lastID); err != nil {
						return fmt.Errorf("failed to scan returned ID: %w", err)
					}
					if i < v.Len() {
						elem := v.Index(i)
						if elem.Kind() == reflect.Ptr {
							elem = elem.Elem()
						}
						if elem.CanAddr() {
							setModelPrimaryKey(elem.Addr().Interface(), lastID)
						}
					}
					i++
				}
			} else {
				// Single insert: scan one ID
				if rows.Next() {
					var lastID int64
					if err := rows.Scan(&lastID); err != nil {
						return fmt.Errorf("failed to scan returned ID: %w", err)
					}
					setModelPrimaryKey(value, lastID)
				}
			}
		} else {
			result, err = q.tx.ExecContext(q.ctx, sqlStr, args...)
		}
	} else {
		var dbConn *sql.DB
		dbConn, err = q.DB()
		if err != nil {
			return err
		}
		if isPostgres || isSQLServer {
			// For PostgreSQL with RETURNING or SQL Server with OUTPUT, use Query instead of Exec
			var rows *sql.Rows
			rows, err = dbConn.QueryContext(q.ctx, sqlStr, args...)
			if err != nil {
				return sanitizeError(fmt.Errorf("failed to execute INSERT query: %w", err), q.isProduction())
			}
			defer func() { _ = rows.Close() }()

			// Handle bulk insert by setting IDs for each element
			v := reflect.ValueOf(value)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}
			if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
				// Bulk insert: scan all returned IDs
				i := 0
				for rows.Next() {
					var lastID int64
					if err := rows.Scan(&lastID); err != nil {
						return fmt.Errorf("failed to scan returned ID: %w", err)
					}
					if i < v.Len() {
						elem := v.Index(i)
						if elem.Kind() == reflect.Ptr {
							elem = elem.Elem()
						}
						if elem.CanAddr() {
							setModelPrimaryKey(elem.Addr().Interface(), lastID)
						}
					}
					i++
				}
			} else {
				// Single insert: scan one ID
				if rows.Next() {
					var lastID int64
					if err := rows.Scan(&lastID); err != nil {
						return fmt.Errorf("failed to scan returned ID: %w", err)
					}
					setModelPrimaryKey(value, lastID)
				}
			}
		} else {
			result, err = dbConn.ExecContext(q.ctx, sqlStr, args...)
		}
	}

	if err != nil {
		return sanitizeError(fmt.Errorf("failed to execute INSERT query: %w", err), q.isProduction())
	}
	q.logQuery(sqlStr, args, start)

	// Populate last insert ID back into the model's primary key field (for non-PostgreSQL and non-SQLServer)
	if !isPostgres && !isSQLServer && result != nil {
		if lastID, err := result.LastInsertId(); err == nil && lastID > 0 {
			// Handle bulk insert by setting IDs for each element
			v := reflect.ValueOf(value)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}
			if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
				// Different databases return different values from LastInsertId() for bulk inserts:
				// - SQLite: returns the LAST auto-increment value
				// - MySQL: returns the FIRST auto-increment value
				isMySQL := q.driver != nil && q.driver.Dialect() == "mysql"
				for i := 0; i < v.Len(); i++ {
					elem := v.Index(i)
					if elem.Kind() == reflect.Ptr {
						elem = elem.Elem()
					}
					if elem.CanAddr() {
						var id int64
						if isMySQL {
							// MySQL returns first ID, so add offset
							id = lastID + int64(i)
						} else {
							// SQLite returns last ID, so subtract to get first then add offset
							id = lastID - int64(v.Len()) + 1 + int64(i)
						}
						setModelPrimaryKey(elem.Addr().Interface(), id)
					}
				}
			} else {
				setModelPrimaryKey(value, lastID)
			}
		}
	}

	// Fire Created event if not disabled
	if !q.withoutEvents {
		attributes := observer.ExtractModelAttributes(value)
		if err := q.dispatcher.DispatchCreated(q.ctx, value, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return fmt.Errorf("created event error: %w", err)
		}
	}

	return nil
}
