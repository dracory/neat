package neat

import "github.com/dracory/neat/support/uid"

// GenerateID creates a new 14-character lowercase safe short ID using a vowel-free Base 16 alphabet.
// It delegates to support/uid.GenerateShortID().
//
// Example:
//
//	id := neat.GenerateID()
func GenerateID() string {
	return uid.GenerateShortID()
}
