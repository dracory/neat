package query

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
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
		// For Oracle bulk inserts, BuildInsert returns empty to signal we should handle it differently
		// Fall back to iterating and inserting one by one
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
			// Disable events for recursive calls to avoid duplicate Creating events
			originalWithoutEvents := q.withoutEvents
			q.withoutEvents = true
			defer func() { q.withoutEvents = originalWithoutEvents }()

			for i := 0; i < v.Len(); i++ {
				elem := v.Index(i).Interface()
				if err := q.Create(elem); err != nil {
					return fmt.Errorf("failed to create element %d: %w", i, err)
				}
			}
			return nil
		}
		return fmt.Errorf("failed to build INSERT query")
	}

	// Postgres: use RETURNING id to get inserted ID
	// SQL Server: uses OUTPUT clause (already added in BuildInsert)
	// Oracle: go-ora driver has issues with RETURNING clause and LastInsertId
	// For Oracle, we'll use a separate SELECT query after INSERT
	if q.isPostgres() {
		sqlStr = sqlStr + " RETURNING id"
	}

	// Execute query
	var result sql.Result
	var err error
	start := time.Now()
	if q.tx != nil {
		if q.isPostgres() || q.isSQLServer() {
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
		if q.isPostgres() || q.isSQLServer() {
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

	// Populate last insert ID back into the model's primary key field
	if q.isOracle() {
		// Oracle: go-ora driver has issues with RETURNING clause and LastInsertId
		// Use sequence-based approach or MAX(id) fallback instead
		var lastID int64

		// Get table name and query the sequence current value
		tableName := q.table
		if tableName == "" {
			// Try to extract table name from the INSERT SQL
			parts := strings.Fields(sqlStr)
			for i, part := range parts {
				if strings.ToUpper(part) == "INTO" && i+1 < len(parts) {
					tableName = strings.Trim(parts[i+1], `"`)
					break
				}
			}
		}

		if tableName != "" {
			// Try to get the sequence name (Oracle convention: TABLENAME_SEQ)
			sequenceName := strings.ToUpper(tableName) + "_SEQ"
			sequenceQuery := fmt.Sprintf("SELECT %s.CURRVAL FROM dual", sequenceName)

			var seqErr error
			if q.tx != nil {
				seqErr = q.tx.QueryRowContext(q.ctx, sequenceQuery).Scan(&lastID)
			} else {
				var dbConn *sql.DB
				dbConn, seqErr = q.DB()
				if seqErr == nil {
					seqErr = dbConn.QueryRowContext(q.ctx, sequenceQuery).Scan(&lastID)
				}
			}

			if seqErr != nil {
				// If sequence doesn't exist or CURRVAL fails, fall back to MAX(id)
				// This is less safe but works as a last resort
				maxIDQuery := fmt.Sprintf("SELECT MAX(id) FROM %s", tableName)
				if q.tx != nil {
					seqErr = q.tx.QueryRowContext(q.ctx, maxIDQuery).Scan(&lastID)
				} else {
					var dbConn *sql.DB
					dbConn, seqErr = q.DB()
					if seqErr == nil {
						seqErr = dbConn.QueryRowContext(q.ctx, maxIDQuery).Scan(&lastID)
					}
				}
				if seqErr != nil {
					// If both approaches fail, silently continue (ID won't be populated)
					// This matches the behavior of other databases when LastInsertId fails
				}
			}

			// Set the ID if we successfully retrieved it
			if lastID > 0 {
				setModelPrimaryKey(value, lastID)
			}
		}
	} else if !q.isPostgres() && !q.isSQLServer() && result != nil {
		// For other databases, use LastInsertId
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
				isMySQL := q.isMySQL()
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
