package query

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// InsertGetId inserts a record and returns the ID.
func (q *Query) InsertGetId(values any) (uint, error) {
	// Build INSERT query
	builder := NewBuilder(q)
	insertSQL, args := builder.BuildInsert(values)
	if insertSQL == "" {
		return 0, fmt.Errorf("failed to build INSERT query")
	}

	// Postgres: use RETURNING id to get inserted ID
	// SQL Server: uses OUTPUT clause (already added in BuildInsert)
	// Oracle: go-ora driver has issues with RETURNING clause and LastInsertId
	// For Oracle, we'll use a separate SELECT query after INSERT
	if q.isPostgres() {
		insertSQL = insertSQL + " RETURNING id"
	}

	start := time.Now()
	var lastID int64

	if q.isPostgres() || q.isSQLServer() {
		// For PostgreSQL with RETURNING or SQL Server with OUTPUT, use QueryRow instead of Exec
		var row *sql.Row
		if q.tx != nil {
			row = q.tx.QueryRowContext(q.ctx, insertSQL, args...)
		} else {
			dbConn, err := q.DB()
			if err != nil {
				return 0, err
			}
			row = dbConn.QueryRowContext(q.ctx, insertSQL, args...)
		}
		if err := row.Scan(&lastID); err != nil {
			return 0, fmt.Errorf("failed to get inserted ID: %w", err)
		}
	} else if q.isOracle() {
		// Oracle: Use RETURNING clause with go-ora driver
		// The driver supports RETURNING clause for INSERT statements
		insertSQL = insertSQL + " RETURNING id INTO :id"

		// We need to use a different approach for Oracle with RETURNING
		// Use Exec with a named parameter for the returned value
		var err error
		var result sql.Result

		// First, try the standard approach with RETURNING clause
		// If that fails, fall back to sequence-based approach
		if q.tx != nil {
			result, err = q.tx.ExecContext(q.ctx, insertSQL, args...)
		} else {
			dbConn, err2 := q.DB()
			if err2 != nil {
				return 0, err2
			}
			result, err = dbConn.ExecContext(q.ctx, insertSQL, args...)
		}

		if err != nil {
			// If RETURNING fails, try sequence-based approach
			// This is a fallback for cases where RETURNING doesn't work
			// Execute the INSERT without RETURNING
			insertSQLWithoutReturning := strings.Replace(insertSQL, " RETURNING id INTO :id", "", 1)
			if q.tx != nil {
				_, err = q.tx.ExecContext(q.ctx, insertSQLWithoutReturning, args...)
			} else {
				dbConn, err2 := q.DB()
				if err2 != nil {
					return 0, err2
				}
				_, err = dbConn.ExecContext(q.ctx, insertSQLWithoutReturning, args...)
			}
			if err != nil {
				return 0, fmt.Errorf("failed to execute INSERT query: %w", err)
			}

			// Get table name and query the sequence current value
			tableName := q.table
			if tableName == "" {
				// Try to extract table name from the INSERT SQL
				parts := strings.Fields(insertSQLWithoutReturning)
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

				if q.tx != nil {
					err = q.tx.QueryRowContext(q.ctx, sequenceQuery).Scan(&lastID)
				} else {
					dbConn, err2 := q.DB()
					if err2 != nil {
						return 0, err2
					}
					err = dbConn.QueryRowContext(q.ctx, sequenceQuery).Scan(&lastID)
				}

				if err != nil {
					// If sequence doesn't exist or CURRVAL fails, fall back to MAX(id)
					// This is less safe but works as a last resort
					maxIDQuery := fmt.Sprintf("SELECT MAX(id) FROM %s", tableName)
					if q.tx != nil {
						err = q.tx.QueryRowContext(q.ctx, maxIDQuery).Scan(&lastID)
					} else {
						dbConn, err2 := q.DB()
						if err2 != nil {
							return 0, err2
						}
						err = dbConn.QueryRowContext(q.ctx, maxIDQuery).Scan(&lastID)
					}
					if err != nil {
						return 0, fmt.Errorf("failed to get last inserted ID: %w", err)
					}
				}
			} else {
				return 0, fmt.Errorf("could not determine table name for ID retrieval")
			}
		} else {
			// RETURNING succeeded, get the last insert ID
			lastID, err = result.LastInsertId()
			if err != nil {
				return 0, fmt.Errorf("failed to get last insert ID: %w", err)
			}
		}
	} else {
		var err error
		var result sql.Result
		if q.tx != nil {
			result, err = q.tx.ExecContext(q.ctx, insertSQL, args...)
		} else {
			dbConn, err2 := q.DB()
			if err2 != nil {
				return 0, err2
			}
			result, err = dbConn.ExecContext(q.ctx, insertSQL, args...)
		}
		if err != nil {
			return 0, fmt.Errorf("failed to execute INSERT query: %w", err)
		}
		lastID, err = result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failed to get last insert ID: %w", err)
		}
	}

	q.logQuery(insertSQL, args, start)

	// Write the ID back to the struct if it's a pointer-to-struct
	setModelPrimaryKey(values, lastID)

	return uint(lastID), nil
}
