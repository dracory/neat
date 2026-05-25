package query

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"
)

// Count returns the number of records matching the query.
func (q *Query) Count(count *int64) error {
	if err := q.validateAggregate("*", count); err != nil {
		return err
	}

	// Use a clone to avoid mutating the query state
	clone := q.Clone().(*Query)
	clone.aggregate = "COUNT"
	clone.aggregateCol = "*"
	clone.distinct = q.distinct
	clone.distinctCols = q.distinctCols

	// Build SELECT query
	builder := NewBuilder(clone)
	sql, args := builder.BuildSelect()

	// Execute query
	start := time.Now()
	var err error
	if q.tx != nil {
		err = q.tx.QueryRowContext(q.ctx, sql, args...).Scan(count)
	} else {
		databaseConn, err := q.ReadDB()
		if err != nil {
			return err
		}
		err = databaseConn.QueryRowContext(q.ctx, sql, args...).Scan(count)
	}

	if err != nil {
		return fmt.Errorf("failed to execute COUNT query: %w", err)
	}

	// Log query if enabled
	q.logQuery(sql, args, start)

	return nil
}

// Sum returns the sum of the specified column.
func (q *Query) Sum(column string, dest any) error {
	if err := q.validateAggregate(column, dest); err != nil {
		return err
	}

	// Set aggregate
	q.aggregate = "SUM"
	q.aggregateCol = column

	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	start := time.Now()
	var err error
	if q.tx != nil {
		err = q.tx.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	} else {
		databaseConn, err := q.ReadDB()
		if err != nil {
			return err
		}
		err = databaseConn.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	}

	if err != nil {
		return fmt.Errorf("failed to execute SUM query: %w", err)
	}

	// Log query if enabled
	q.logQuery(sql, args, start)

	return nil
}

// Avg returns the average of the specified column.
func (q *Query) Avg(column string, dest any) error {
	if err := q.validateAggregate(column, dest); err != nil {
		return err
	}

	// Set aggregate
	q.aggregate = "AVG"
	q.aggregateCol = column

	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	start := time.Now()
	var err error
	if q.tx != nil {
		err = q.tx.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	} else {
		databaseConn, err := q.ReadDB()
		if err != nil {
			return err
		}
		err = databaseConn.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	}

	if err != nil {
		return fmt.Errorf("failed to execute AVG query: %w", err)
	}

	// Log query if enabled
	q.logQuery(sql, args, start)

	return nil
}

// Min returns the minimum value of the specified column.
func (q *Query) Min(column string, dest any) error {
	if err := q.validateAggregate(column, dest); err != nil {
		return err
	}

	// Set aggregate
	q.aggregate = "MIN"
	q.aggregateCol = column

	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	start := time.Now()
	var err error
	if q.tx != nil {
		err = q.tx.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	} else {
		databaseConn, err := q.ReadDB()
		if err != nil {
			return err
		}
		err = databaseConn.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	}

	if err != nil {
		return fmt.Errorf("failed to execute MIN query: %w", err)
	}

	// Log query if enabled
	q.logQuery(sql, args, start)

	return nil
}

// Max returns the maximum value of the specified column.
func (q *Query) Max(column string, dest any) error {
	if err := q.validateAggregate(column, dest); err != nil {
		return err
	}

	// Set aggregate
	q.aggregate = "MAX"
	q.aggregateCol = column

	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	start := time.Now()
	var err error
	if q.tx != nil {
		err = q.tx.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	} else {
		databaseConn, err := q.ReadDB()
		if err != nil {
			return err
		}
		err = databaseConn.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	}

	if err != nil {
		return fmt.Errorf("failed to execute MAX query: %w", err)
	}

	// Log query if enabled
	q.logQuery(sql, args, start)

	return nil
}

// Exists checks if any records match the query.
func (q *Query) Exists(exists *bool) error {
	// Set aggregate for EXISTS check
	q.aggregate = "COUNT"
	q.aggregateCol = "1"
	limit := 1
	q.limit = &limit

	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	start := time.Now()
	var count int64
	var err error
	if q.tx != nil {
		err = q.tx.QueryRowContext(q.ctx, sql, args...).Scan(&count)
	} else {
		databaseConn, err := q.ReadDB()
		if err != nil {
			return err
		}
		err = databaseConn.QueryRowContext(q.ctx, sql, args...).Scan(&count)
	}

	if err != nil {
		return fmt.Errorf("failed to execute EXISTS query: %w", err)
	}

	*exists = count > 0

	// Log query if enabled
	q.logQuery(sql, args, start)

	return nil
}

// Pluck retrieves a single column's values from the query results.
func (q *Query) Pluck(column string, dest any) error {
	// Set select to only the specified column
	q.selects = []selectClause{{expr: column}}

	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	var err error
	if q.tx != nil {
		rows, err := q.tx.QueryContext(q.ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to execute PLUCK query: %w", err)
		}
		defer rows.Close()

		return q.pluckRows(rows, dest)
	}

	databaseConn, err := q.ReadDB()
	if err != nil {
		return err
	}

	rows, err := databaseConn.QueryContext(q.ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to execute PLUCK query: %w", err)
	}
	defer rows.Close()

	return q.pluckRows(rows, dest)
}

// Value retrieves a single column's value from the first result.
func (q *Query) Value(column string, dest any) error {
	// Set select to only the specified column
	q.selects = []selectClause{{expr: column}}

	// Set limit to 1
	limit := 1
	q.limit = &limit

	// Build SELECT query
	builder := NewBuilder(q)
	sql, args := builder.BuildSelect()

	// Execute query
	start := time.Now()
	var err error
	if q.tx != nil {
		err = q.tx.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	} else {
		databaseConn, err := q.ReadDB()
		if err != nil {
			return err
		}
		err = databaseConn.QueryRowContext(q.ctx, sql, args...).Scan(dest)
	}

	if err != nil {
		return fmt.Errorf("failed to execute VALUE query: %w", err)
	}

	// Log query if enabled
	q.logQuery(sql, args, start)

	return nil
}

// pluckRows scans a single column from database rows into the destination.
func (q *Query) pluckRows(rows *sql.Rows, dest any) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}

	destValue = destValue.Elem()

	// Handle slice destination
	if destValue.Kind() == reflect.Slice {
		elemType := destValue.Type().Elem()
		columns, _ := rows.Columns()

		for rows.Next() {
			// Create new element
			elemPtr := reflect.New(elemType)
			elem := elemPtr.Elem()

			if elem.Kind() == reflect.Map {
				// Scan into a temporary []any then build the map
				values := make([]any, len(columns))
				ptrs := make([]any, len(columns))
				for i := range values {
					ptrs[i] = &values[i]
				}
				if err := rows.Scan(ptrs...); err != nil {
					return fmt.Errorf("failed to scan row: %w", err)
				}
				m := reflect.MakeMap(elemType)
				keyType := elemType.Key()
				for i, col := range columns {
					m.SetMapIndex(reflect.ValueOf(col).Convert(keyType), reflect.ValueOf(values[i]))
				}
				elem.Set(m)
			} else {
				if err := rows.Scan(elem.Addr().Interface()); err != nil {
					return fmt.Errorf("failed to scan row: %w", err)
				}
			}

			// Append to slice
			destValue.Set(reflect.Append(destValue, elem))
		}

		return rows.Err()
	}

	return fmt.Errorf("unsupported destination type for PLUCK: %T", dest)
}
