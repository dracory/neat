package uid

import (
	"strings"
	"sync"
	"time"
)

// crockfordAlphabet is the Crockford Base32 alphabet (omits I, L, O, U to avoid ambiguity).
const crockfordAlphabet = "0123456789abcdefghjkmnpqrstvwxyz"

var (
	idMutex       sync.Mutex
	lastTimestamp int64
	counter       int64
)

// GenerateShortID creates a new 11-character lowercase short ID.
// It encodes the current microsecond timestamp in Crockford Base32.
// Thread-safe via mutex to prevent duplicate IDs under concurrency.
func GenerateShortID() string {
	idMutex.Lock()
	defer idMutex.Unlock()

	ts := time.Now().UnixMicro()
	if ts == lastTimestamp {
		counter++
	} else {
		lastTimestamp = ts
		counter = 0
	}

	// Pack timestamp and sub-microsecond counter into a single int64.
	// The counter provides up to 16 unique IDs within the same microsecond.
	composite := (ts << 4) | counter
	return strings.ToLower(encodeCrockford(composite))
}

// encodeCrockford encodes a positive int64 into Crockford Base32.
func encodeCrockford(n int64) string {
	if n == 0 {
		return "0"
	}
	var buf [13]byte // max length for int64 in base32
	i := len(buf) - 1
	for n > 0 {
		buf[i] = crockfordAlphabet[n&0x1f]
		n >>= 5
		i--
	}
	return string(buf[i+1:])
}

// NormalizeID normalizes an ID to lowercase for consistent lookups.
func NormalizeID(id string) string {
	return strings.ToLower(strings.TrimSpace(id))
}

// IsShortID checks whether the given string matches short ID lengths (11 or 21 chars).
func IsShortID(id string) bool {
	length := len(id)
	return length == 11 || length == 21
}
