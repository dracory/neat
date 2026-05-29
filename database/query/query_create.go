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
	isPostgres := q.driver != nil && q.driver.Dialect() == "postgres"
	if isPostgres {
		sqlStr = sqlStr + " RETURNING id"
	}

	// Execute query
	var result sql.Result
	var err error
	start := time.Now()
	if q.tx != nil {
		if isPostgres {
			// For PostgreSQL with RETURNING, use Query instead of Exec
			var rows *sql.Rows
			rows, err = q.tx.QueryContext(q.ctx, sqlStr, args...)
			if err != nil {
				return fmt.Errorf("failed to execute INSERT query: %w", err)
			}
			defer rows.Close()

			if rows.Next() {
				var lastID int64
				if err := rows.Scan(&lastID); err != nil {
					return fmt.Errorf("failed to scan returned ID: %w", err)
				}
				setModelPrimaryKey(value, lastID)
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
		if isPostgres {
			// For PostgreSQL with RETURNING, use Query instead of Exec
			var rows *sql.Rows
			rows, err = dbConn.QueryContext(q.ctx, sqlStr, args...)
			if err != nil {
				return fmt.Errorf("failed to execute INSERT query: %w", err)
			}
			defer rows.Close()

			if rows.Next() {
				var lastID int64
				if err := rows.Scan(&lastID); err != nil {
					return fmt.Errorf("failed to scan returned ID: %w", err)
				}
				setModelPrimaryKey(value, lastID)
			}
		} else {
			result, err = dbConn.ExecContext(q.ctx, sqlStr, args...)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to execute INSERT query: %w", err)
	}
	q.logQuery(sqlStr, args, start)

	// Populate last insert ID back into the model's primary key field (for non-PostgreSQL)
	if !isPostgres {
		if lastID, err := result.LastInsertId(); err == nil && lastID > 0 {
			// Handle bulk insert by setting IDs for each element
			v := reflect.ValueOf(value)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}
			if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
				// For bulk inserts, set sequential IDs starting from lastID - len + 1
				startID := lastID - int64(v.Len()) + 1
				for i := 0; i < v.Len(); i++ {
					elem := v.Index(i)
					if elem.Kind() == reflect.Ptr {
						elem = elem.Elem()
					}
					if elem.CanAddr() {
						setModelPrimaryKey(elem.Addr().Interface(), startID+int64(i))
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
