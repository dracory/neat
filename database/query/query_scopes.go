package query

import contractsorm "github.com/dracory/neat/contracts/database/orm"

// applyScopes applies registered scope functions and returns the modified query.
func (q *Query) applyScopes() *Query {
	if len(q.scopes) == 0 {
		return q
	}
	var result contractsorm.Query = q
	for _, fn := range q.scopes {
		result = fn(result)
	}
	if r, ok := result.(*Query); ok {
		return r
	}
	return q
}

// Scopes registers scope functions to be applied to the query.
func (q *Query) Scopes(funcs ...func(contractsorm.Query) contractsorm.Query) contractsorm.Query {
	newQ := q.Clone().(*Query)
	newQ.scopes = append(newQ.scopes, funcs...)
	return newQ
}
