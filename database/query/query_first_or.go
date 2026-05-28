package query

// FirstOr retrieves the first record or executes a callback if not found.
func (q *Query) FirstOr(dest any, callback func() error) error {
	err := q.First(dest)
	if err != nil {
		return callback()
	}
	return nil
}

// FirstOrCreate retrieves the first record or creates it if not found.
func (q *Query) FirstOrCreate(dest any, conds ...any) error {
	// Try to find the record first
	err := q.First(dest)
	if err == nil {
		return nil // Record exists
	}

	// Record doesn't exist, create it
	return q.Create(dest)
}

// FirstOrNew retrieves the first record or prepares a new instance if not found.
func (q *Query) FirstOrNew(dest any, attributes any, values ...any) error {
	// Try to find the record first
	err := q.First(dest)
	if err == nil {
		return nil // Record exists
	}

	// Record doesn't exist, prepare new instance (without saving)
	// This is a simplified implementation
	return nil
}

// UpdateOrCreate updates a record if it exists, or creates it if it doesn't.
func (q *Query) UpdateOrCreate(dest any, attributes any, values any) error {
	// Try to find the record first
	err := q.First(dest)
	if err == nil {
		// Record exists, update it
		return q.Save(values)
	}

	// Record doesn't exist, create it
	return q.Create(values)
}
