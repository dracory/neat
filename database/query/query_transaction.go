package query

import (
	"database/sql"
	"fmt"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// Transaction runs a callback wrapped in a database transaction.
func (q *Query) Transaction(txFunc func(tx contractsorm.Query) error, opts ...*sql.TxOptions) error {
	txQuery, err := q.Begin(opts...)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	txQ := txQuery.(*Query)

	defer func() {
		if p := recover(); p != nil {
			_ = txQ.Rollback()
			panic(p)
		}
	}()

	if err := txFunc(txQ); err != nil {
		if rbErr := txQ.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %w", err, rbErr)
		}
		return err
	}

	return txQ.Commit()
}

// doCommit runs beforeCommit hooks, commits, then runs afterCommit hooks.
func (q *Query) doCommit() error {
	for _, cb := range q.beforeCommit {
		if err := cb(); err != nil {
			_ = q.tx.Rollback()
			return fmt.Errorf("beforeCommit hook error: %w", err)
		}
	}

	if err := q.tx.Commit(); err != nil {
		return err
	}

	for _, cb := range q.afterCommit {
		if err := cb(); err != nil {
			return fmt.Errorf("afterCommit hook error: %w", err)
		}
	}

	return nil
}

// doRollback runs beforeRollback hooks, rolls back, then runs afterRollback hooks.
func (q *Query) doRollback() error {
	for _, cb := range q.beforeRollback {
		_ = cb()
	}

	err := q.tx.Rollback()
	for _, cb := range q.afterRollback {
		_ = cb()
	}
	return err
}

// Begin starts a new database transaction.
func (q *Query) Begin(opts ...*sql.TxOptions) (contractsorm.Query, error) {
	var txOpts *sql.TxOptions
	if len(opts) > 0 {
		txOpts = opts[0]
	}

	var tx *sql.Tx
	var err error
	if q.tx != nil {
		// Already in a transaction, create a savepoint for nested transaction
		q.savepointLevel++
		savepointName := fmt.Sprintf("neat_sp_%d", q.savepointLevel)
		_, err = q.tx.ExecContext(q.ctx, fmt.Sprintf("SAVEPOINT %s", savepointName))
		if err != nil {
			return nil, fmt.Errorf("failed to create savepoint: %w", err)
		}
		// Create a new query instance with the savepoint
		newQuery := q.Clone().(*Query)
		newQuery.savepointName = savepointName
		newQuery.inTransaction = true
		newQuery.tx = q.tx
		return newQuery, nil
	}

	dbConn := q.writeConn()

	tx, err = dbConn.BeginTx(q.ctx, txOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Create a new query instance with the transaction
	newQuery := q.Clone().(*Query)
	newQuery.tx = tx
	newQuery.inTransaction = true
	newQuery.savepointLevel = 0
	return newQuery, nil
}

// Commit commits the current transaction.
func (q *Query) Commit() error {
	if !q.inTransaction || q.tx == nil {
		return fmt.Errorf("not in a transaction")
	}

	// If this is a nested transaction (savepoint), release it
	if q.savepointName != "" {
		_, err := q.tx.ExecContext(q.ctx, fmt.Sprintf("RELEASE SAVEPOINT %s", q.savepointName))
		if err != nil {
			return fmt.Errorf("failed to release savepoint: %w", err)
		}
		q.savepointName = ""
		q.inTransaction = false // Nested transactions are also "transactions" in our model
		return nil
	}

	err := q.doCommit()
	q.inTransaction = false
	q.tx = nil
	return err
}

// Rollback rolls back the current transaction.
func (q *Query) Rollback() error {
	if !q.inTransaction || q.tx == nil {
		return fmt.Errorf("not in a transaction")
	}

	// If this is a nested transaction (savepoint), rollback to it
	if q.savepointName != "" {
		_, err := q.tx.ExecContext(q.ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", q.savepointName))
		if err != nil {
			return fmt.Errorf("failed to rollback to savepoint: %w", err)
		}
		// Release the savepoint after rollback
		_, err = q.tx.ExecContext(q.ctx, fmt.Sprintf("RELEASE SAVEPOINT %s", q.savepointName))
		if err != nil {
			return fmt.Errorf("failed to release savepoint: %w", err)
		}
		q.savepointName = ""
		q.inTransaction = false
		return nil
	}

	err := q.doRollback()
	q.inTransaction = false
	q.tx = nil
	return err
}

// RollbackTo rolls back to a specific savepoint.
func (q *Query) RollbackTo(level string) error {
	if !q.inTransaction || q.tx == nil {
		return fmt.Errorf("not in a transaction")
	}

	// Execute savepoint rollback (dialect-specific)
	_, err := q.tx.ExecContext(q.ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", level))
	if err != nil {
		return fmt.Errorf("failed to rollback to savepoint: %w", err)
	}
	return nil
}

// SavePoint creates a new savepoint within the transaction.
func (q *Query) SavePoint(name string) error {
	if !q.inTransaction || q.tx == nil {
		return fmt.Errorf("not in a transaction")
	}

	// Execute savepoint creation (dialect-specific)
	_, err := q.tx.ExecContext(q.ctx, fmt.Sprintf("SAVEPOINT %s", name))
	if err != nil {
		return fmt.Errorf("failed to create savepoint: %w", err)
	}
	return nil
}

// BeforeCommit registers a callback to run before the transaction is committed.
func (q *Query) BeforeCommit(callback func() error) {
	q.beforeCommit = append(q.beforeCommit, callback)
}

// AfterCommit registers a callback to run after the transaction is committed.
func (q *Query) AfterCommit(callback func() error) {
	q.afterCommit = append(q.afterCommit, callback)
}

// BeforeRollback registers a callback to run before the transaction is rolled back.
func (q *Query) BeforeRollback(callback func() error) {
	q.beforeRollback = append(q.beforeRollback, callback)
}

// AfterRollback registers a callback to run after the transaction is rolled back.
func (q *Query) AfterRollback(callback func() error) {
	q.afterRollback = append(q.afterRollback, callback)
}
