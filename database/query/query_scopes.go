package query

import (
	"reflect"

	contractsorm "github.com/dracory/neat/contracts/database/orm"
)

// applyScopes applies registered global and per-query scope functions and returns the modified query.
func (q *Query) applyScopes() *Query {
	var allScopes []func(contractsorm.Query) contractsorm.Query

	if !q.withoutGlobalScopes && len(q.globalScopes) > 0 {
		allScopes = append(allScopes, q.globalScopes...)
	}

	if len(q.scopes) > 0 {
		allScopes = append(allScopes, q.scopes...)
	}

	if len(allScopes) == 0 {
		return q
	}

	var result contractsorm.Query = q
	for _, fn := range allScopes {
		if fn == nil {
			continue
		}
		if len(q.disabledScopes) > 0 {
			ptr := reflect.ValueOf(fn).Pointer()
			if q.disabledScopes[ptr] {
				continue
			}
		}
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

// WithoutScope removes specific scope function(s) from being applied to the query.
func (q *Query) WithoutScope(funcs ...func(contractsorm.Query) contractsorm.Query) contractsorm.Query {
	newQ := q.Clone().(*Query)
	if newQ.disabledScopes == nil {
		newQ.disabledScopes = make(map[uintptr]bool)
	}
	for _, fn := range funcs {
		if fn != nil {
			ptr := reflect.ValueOf(fn).Pointer()
			newQ.disabledScopes[ptr] = true
		}
	}
	return newQ
}

// WithoutScopes removes all per-query scopes from being applied to the query.
func (q *Query) WithoutScopes() contractsorm.Query {
	newQ := q.Clone().(*Query)
	newQ.scopes = nil
	return newQ
}

// WithoutGlobalScopes disables global scopes defined on the model for this query.
func (q *Query) WithoutGlobalScopes() contractsorm.Query {
	newQ := q.Clone().(*Query)
	newQ.withoutGlobalScopes = true
	return newQ
}
