package query

import contractsorm "github.com/dracory/neat/contracts/database/orm"

// WithTrashed includes soft-deleted records in the query results.
func (q *Query) WithTrashed() contractsorm.Query {
	newQuery := *q
	newQuery.withTrashed = true
	newQuery.onlyTrashed = false
	newQuery.withoutTrashed = false
	return &newQuery
}

// OnlyTrashed returns only soft-deleted records.
func (q *Query) OnlyTrashed() contractsorm.Query {
	newQuery := *q
	newQuery.withTrashed = false
	newQuery.onlyTrashed = true
	newQuery.withoutTrashed = false
	return &newQuery
}

// WithoutTrashed excludes soft-deleted records from the query results (default behavior).
func (q *Query) WithoutTrashed() contractsorm.Query {
	newQuery := *q
	newQuery.withTrashed = false
	newQuery.onlyTrashed = false
	newQuery.withoutTrashed = true
	return &newQuery
}
