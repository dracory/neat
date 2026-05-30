package seeder

import (
	"sync"

	"github.com/dracory/neat/contracts/database/seeder"
)

// Global seeder registry with mutex for thread safety
var (
	seederRegistry = make(map[string]seeder.Seeder)
	registryMutex  sync.RWMutex
)

// RegisterSeeder registers a seeder in the global registry
func RegisterSeeder(name string, s seeder.Seeder) {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	seederRegistry[name] = s
}

// GetSeeder retrieves a seeder from the global registry
func GetSeeder(name string) seeder.Seeder {
	registryMutex.RLock()
	defer registryMutex.RUnlock()
	return seederRegistry[name]
}

// GetSeeders retrieves all seeders from the global registry
func GetSeeders() []seeder.Seeder {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	seeders := make([]seeder.Seeder, 0, len(seederRegistry))
	for _, s := range seederRegistry {
		seeders = append(seeders, s)
	}
	return seeders
}

// ClearRegistry clears the global seeder registry (useful for testing)
func ClearRegistry() {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	seederRegistry = make(map[string]seeder.Seeder)
}
