package query

// EnableDebug enables debug mode at runtime, allowing detailed SQL error messages.
// This is thread-safe and can be called while the application is running.
func (q *Query) EnableDebug() {
	q.debugMu.Lock()
	defer q.debugMu.Unlock()
	q.debugState = true
	if q.dbConfig != nil {
		q.dbConfig.Debug = true
	}
}

// DisableDebug disables debug mode at runtime, sanitizing SQL error messages.
// This is thread-safe and can be called while the application is running.
func (q *Query) DisableDebug() {
	q.debugMu.Lock()
	defer q.debugMu.Unlock()
	q.debugState = false
	if q.dbConfig != nil {
		q.dbConfig.Debug = false
	}
}

// IsDebug returns true if debug mode is enabled.
// This checks both the runtime debug state and the dbConfig.Debug setting.
func (q *Query) IsDebug() bool {
	q.debugMu.RLock()
	defer q.debugMu.RUnlock()
	if q.debugState {
		return true
	}
	if q.dbConfig != nil {
		return q.dbConfig.Debug
	}
	return false
}
