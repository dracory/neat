package query

import (
	contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// WithSoftDeleted includes soft-deleted records in the query results.
func (q *Query) WithSoftDeleted() contractsorm.Query {
	newQuery := q.Clone().(*Query)
	newQuery.includeSoftDeleted = true
	newQuery.onlySoftDeleted = false
	newQuery.excludeSoftDeleted = false
	return newQuery
}

// WithTrashed includes soft-deleted records in the query results.
//
// Deprecated: Use WithSoftDeleted() instead.
func (q *Query) WithTrashed() contractsorm.Query {
	return q.WithSoftDeleted()
}

// OnlySoftDeleted returns only soft-deleted records.
func (q *Query) OnlySoftDeleted() contractsorm.Query {
	newQuery := q.Clone().(*Query)
	newQuery.includeSoftDeleted = false
	newQuery.onlySoftDeleted = true
	newQuery.excludeSoftDeleted = false
	return newQuery
}

// OnlyTrashed returns only soft-deleted records.
//
// Deprecated: Use OnlySoftDeleted() instead.
func (q *Query) OnlyTrashed() contractsorm.Query {
	return q.OnlySoftDeleted()
}

// WithoutSoftDeleted excludes soft-deleted records from the query results (default behavior).
func (q *Query) WithoutSoftDeleted() contractsorm.Query {
	newQuery := q.Clone().(*Query)
	newQuery.includeSoftDeleted = false
	newQuery.onlySoftDeleted = false
	newQuery.excludeSoftDeleted = true
	return newQuery
}

// WithoutTrashed excludes soft-deleted records from the query results (default behavior).
//
// Deprecated: Use WithoutSoftDeleted() instead.
func (q *Query) WithoutTrashed() contractsorm.Query {
	return q.WithoutSoftDeleted()
}
