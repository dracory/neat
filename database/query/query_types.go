package query

import (
	"context"
	"database/sql"
	"sync"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
	"github.com/dracory/neat/contracts/log"
	"github.com/dracory/neat/database/db"
	"github.com/dracory/neat/database/driver"
	"github.com/dracory/neat/database/observer"
)

// Query implements the Query interface using native database/sql.
type Query struct {
	ctx        context.Context
	db         *sql.DB // primary (write) connection
	readDB     *sql.DB // optional read-replica connection; nil means use db
	writeDB    *sql.DB // optional explicit write connection; nil means use db
	driver     driver.Driver
	connection string
	dbConfig   *db.DBConfig
	log        log.Log
	queryLog   *[]contractsorm.QueryLog
	enableLog  bool

	// Runtime debug state
	debugMu    sync.RWMutex
	debugState bool

	// Query state
	table        string
	tableArgs    []any
	model        any
	selects      []selectClause
	wheres       []whereClause
	joins        []joinClause
	groups       []string
	havings      []havingClause
	orders       []orderClause
	limit        *int
	offset       *int
	distinct     bool
	distinctCols []string
	aggregate    string
	aggregateCol string

	// Transaction state
	inTransaction  bool
	tx             *sql.Tx
	savepointLevel int
	savepointName  string

	// Observer state
	modelToObserver []contractsorm.ModelToObserver
	withoutEvents   bool
	dispatcher      *observer.Dispatcher

	// Transaction lifecycle hooks
	beforeCommit   []func() error
	afterCommit    []func() error
	beforeRollback []func() error
	afterRollback  []func() error

	// Raw SQL state
	rawSQL  string
	rawArgs []any

	// Lock state
	lockForUpdate bool
	sharedLock    bool

	// Scopes
	scopes []func(contractsorm.Query) contractsorm.Query

	// Omit columns
	omitColumns []string

	// Soft delete state
	withTrashed    bool
	onlyTrashed    bool
	withoutTrashed bool

	// Eager loading state
	withRelations       []string
	relationConstraints map[string]func(contractsorm.Query) contractsorm.Query

	// Count and Exists subqueries
	withCountQueries  []countQuery
	withExistsQueries []existsQuery

	// Validation error
	buildError error
}

// countQuery represents a count subquery for eager loading.
type countQuery struct {
	relation   string
	column     string
	constraint func(contractsorm.Query) contractsorm.Query
}

// existsQuery represents an exists subquery for eager loading.
type existsQuery struct {
	relation   string
	constraint func(contractsorm.Query) contractsorm.Query
}

// whereClause represents a WHERE clause in a query.
type whereClause struct {
	_type string // "and", "or"
	query string
	args  []any
}

// joinClause represents a JOIN clause in a query.
type joinClause struct {
	_type string // "join", "left join", "right join", "cross join"
	query string
	args  []any
}

// havingClause represents a HAVING clause in a query.
type havingClause struct {
	query string
	args  []any
}

// selectClause represents a SELECT clause in a query.
type selectClause struct {
	expr string
	args []any
}

// RawExpression represents a raw SQL expression that can be used as a value in maps
type RawExpression struct {
	SQL  string
	Args []any
}

// orderClause represents an ORDER BY clause in a query.
type orderClause struct {
	column    string
	direction string // "asc", "desc"
}

// RawExpr creates a new raw SQL expression for use in Create/Update values.
// WARNING: This function injects SQL directly without parameterization. NEVER pass user input to this function.
// The SQL will be injected directly into the query, creating a SQL injection vulnerability if used with untrusted data.
// Example: db.Table("users").Create(map[string]any{"created_at": RawExpr("NOW()")})
// Safe usage: Only use with hardcoded SQL expressions like "NOW()", "CURRENT_TIMESTAMP", etc.
// Dangerous usage: RawExpr(userInput) - DO NOT DO THIS
func RawExpr(sql string, args ...any) RawExpression {
	// Filter out nil arguments to prevent issues
	filteredArgs := make([]any, 0, len(args))
	for _, arg := range args {
		if arg != nil {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	return RawExpression{SQL: sql, Args: filteredArgs}
}
