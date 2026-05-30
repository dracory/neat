package seeder

import (
	"fmt"
	"sync"

	"github.com/dracory/neat/contracts/database/seeder"
)

// Runner implements the seeder.Facade interface
type Runner struct {
	seeders    []seeder.Seeder
	seedersMu  sync.RWMutex
	callOnce   map[string]bool
	callOnceMu sync.Mutex
}

// NewRunner creates a new seeder runner
func NewRunner() *Runner {
	return &Runner{
		seeders:  make([]seeder.Seeder, 0),
		callOnce: make(map[string]bool),
	}
}

// Register registers seeders
func (r *Runner) Register(seeders []seeder.Seeder) {
	r.seedersMu.Lock()
	defer r.seedersMu.Unlock()
	r.seeders = append(r.seeders, seeders...)
}

// GetSeeder gets a seeder instance from the seeders
func (r *Runner) GetSeeder(name string) seeder.Seeder {
	r.seedersMu.RLock()
	defer r.seedersMu.RUnlock()
	for _, s := range r.seeders {
		if s.Signature() == name {
			return s
		}
	}
	return nil
}

// GetSeeders gets all the seeders
func (r *Runner) GetSeeders() []seeder.Seeder {
	r.seedersMu.RLock()
	defer r.seedersMu.RUnlock()
	return r.seeders
}

// Call executes the specified seeder(s)
func (r *Runner) Call(seeders []seeder.Seeder) error {
	for _, s := range seeders {
		if err := s.Run(); err != nil {
			return fmt.Errorf("failed to run seeder %s: %w", s.Signature(), err)
		}
	}
	return nil
}

// CallOnce executes the specified seeder(s) only once
func (r *Runner) CallOnce(seeders []seeder.Seeder) error {
	r.callOnceMu.Lock()
	defer r.callOnceMu.Unlock()

	for _, s := range seeders {
		signature := s.Signature()
		if r.callOnce[signature] {
			continue
		}

		if err := s.Run(); err != nil {
			return fmt.Errorf("failed to run seeder %s: %w", signature, err)
		}

		r.callOnce[signature] = true
	}
	return nil
}

// ResetCallOnce clears the CallOnce tracking (useful for testing)
func (r *Runner) ResetCallOnce() {
	r.callOnceMu.Lock()
	defer r.callOnceMu.Unlock()
	r.callOnce = make(map[string]bool)
}
