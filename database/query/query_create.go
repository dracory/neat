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
			// Clone the query to avoid race condition with shared query instance
			clonedQuery := q.WithoutEvents().(*Query)

			for i := 0; i < v.Len(); i++ {
				elem := v.Index(i).Interface()
				if err := clonedQuery.Create(elem); err != nil {
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
	ctx, cancel := q.timeoutContext()
	defer cancel()
	var result sql.Result
	var err error
	start := time.Now()
	if q.tx != nil {
		if q.isPostgres() || q.isSQLServer() {
			// For PostgreSQL with RETURNING or SQL Server with OUTPUT, use Query instead of Exec
			var rows *sql.Rows
			rows, err = q.tx.QueryContext(ctx, sqlStr, args...)
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
				if err := rows.Err(); err != nil {
					return fmt.Errorf("error during rows iteration: %w", err)
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
				if err := rows.Err(); err != nil {
					return fmt.Errorf("error during rows iteration: %w", err)
				}
			}
		} else {
			result, err = q.tx.ExecContext(ctx, sqlStr, args...)
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
			rows, err = dbConn.QueryContext(ctx, sqlStr, args...)
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
				if err := rows.Err(); err != nil {
					return fmt.Errorf("error during rows iteration: %w", err)
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
				if err := rows.Err(); err != nil {
					return fmt.Errorf("error during rows iteration: %w", err)
				}
			}
		} else {
			result, err = dbConn.ExecContext(ctx, sqlStr, args...)
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
			// Validate table name is a simple identifier to prevent SQL injection
			if !isSimpleIdentifier(tableName) {
				return fmt.Errorf("invalid table name: %s", tableName)
			}
			// Try multiple sequence naming conventions
			// Common patterns: TABLENAME_SEQ, SEQ_TABLENAME, TABLENAME_ID_SEQ
			sequencePatterns := []string{
				strings.ToUpper(tableName) + "_SEQ",
				"SEQ_" + strings.ToUpper(tableName),
				strings.ToUpper(tableName) + "_ID_SEQ",
			}

			var seqErr error
			for _, sequenceName := range sequencePatterns {
				if !isSimpleIdentifier(sequenceName) {
					return fmt.Errorf("invalid sequence name: %s", sequenceName)
				}
				sequenceQuery := fmt.Sprintf("SELECT %s.CURRVAL FROM dual", sequenceName)

				if q.tx != nil {
					seqErr = q.tx.QueryRowContext(ctx, sequenceQuery).Scan(&lastID)
				} else {
					var dbConn *sql.DB
					dbConn, seqErr = q.DB()
					if seqErr == nil {
						seqErr = dbConn.QueryRowContext(ctx, sequenceQuery).Scan(&lastID)
					}
				}

				if seqErr == nil && lastID > 0 {
					break // Successfully retrieved ID from sequence
				}
			}

			if seqErr != nil || lastID == 0 {
				// All sequence patterns failed (e.g. table uses GENERATED BY DEFAULT AS IDENTITY
				// rather than an explicit named sequence). Fall back to MAX(id).
				// When inside a caller-supplied transaction the INSERT and this SELECT share
				// the same session, which is the safest available option on Oracle without
				// RETURNING support. Outside a transaction the race window is acknowledged;
				// callers that need strict correctness should wrap Create in a transaction.
				maxIDQuery := fmt.Sprintf("SELECT MAX(id) FROM %s", tableName)
				if q.tx != nil {
					if maxErr := q.tx.QueryRowContext(ctx, maxIDQuery).Scan(&lastID); maxErr != nil {
						_ = maxErr // If MAX(id) also fails, leave ID unpopulated
					}
				} else {
					var dbConn *sql.DB
					dbConn, seqErr = q.DB()
					if seqErr == nil {
						if maxErr := dbConn.QueryRowContext(ctx, maxIDQuery).Scan(&lastID); maxErr != nil {
							_ = maxErr // If MAX(id) also fails, leave ID unpopulated
						}
					}
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
				// Skip bulk insert ID population for MySQL and SQLite
				// The calculation assumes consecutive IDs, which fails with triggers, custom sequences, or ID gaps
				// Users should use InsertGetId() for bulk inserts if they need IDs
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
