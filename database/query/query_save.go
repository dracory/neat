package query

import (
	"fmt"

	"github.com/dracory/neat/database/observer"
)

// Save saves the model to the database (INSERT if no primary key, UPDATE otherwise).
func (q *Query) Save(value any) error {
	// Fire Saving event
	if !q.withoutEvents {
		attributes := observer.ExtractModelAttributes(value)
		if err := q.dispatcher.DispatchSaving(q.ctx, value, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return fmt.Errorf("saving event error: %w", err)
		}
	}

	idVal, _ := getPrimaryKeyValueAny(value)
	var saveErr error
	if !isPrimaryKeyZero(value) {
		// UPDATE: set WHERE id = <id> on a clone, then call Update with the value
		clone := q.Clone().(*Query)
		clone.wheres = append(clone.wheres, whereClause{_type: "and", query: "id = ?", args: []any{idVal}})
		_, saveErr = clone.Update(value)
	} else {
		saveErr = q.Create(value)
	}

	if saveErr != nil {
		return saveErr
	}

	// Fire Saved event
	if !q.withoutEvents {
		attributes := observer.ExtractModelAttributes(value)
		if err := q.dispatcher.DispatchSaved(q.ctx, value, q.modelToObserver, nil, attributes, nil, q); err != nil {
			return fmt.Errorf("saved event error: %w", err)
		}
	}
	return nil
}

// SaveQuietly saves the model without firing events.
func (q *Query) SaveQuietly(value any) error {
	return q.WithoutEvents().(*Query).Save(value)
}
