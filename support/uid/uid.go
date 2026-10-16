package uid

import (
	realuid "github.com/dracory/uid"
)

// GenerateShortID creates a new 14-character lowercase safe short ID.
// It delegates to github.com/dracory/uid.SafeShortID().
func GenerateShortID() string {
	return realuid.SafeShortID()
}

// SafeShortID creates a new 14-character lowercase safe short ID.
// It delegates to github.com/dracory/uid.SafeShortID().
func SafeShortID() string {
	return realuid.SafeShortID()
}

// NormalizeID normalizes an ID to lowercase for consistent lookups.
func NormalizeID(id string) string {
	return realuid.NormalizeID(id)
}

// IsShortID checks whether the given string matches short ID lengths (11, 14, or 21 chars).
func IsShortID(id string) bool {
	length := len(id)
	return length == 11 || length == 14 || length == 21
}
