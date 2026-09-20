package neat

import "github.com/dracory/neat/support/uid"

// GenerateID creates a new 11-character lowercase short ID using Crockford Base32.
// It delegates to support/uid.GenerateShortID().
//
// Example:
//
//	id := neat.GenerateID()
func GenerateID() string {
	return uid.GenerateShortID()
}
