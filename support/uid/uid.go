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
//
// Business Logic:
//   - Uses time.Now().UnixMicro() as the uniqueness base.
//   - Maintains package-level lastTimestamp and counter across calls.
//   - Counter increments when the timestamp equals the previous call;
//     resets to 0 when the timestamp changes.
//   - The counter is 4 bits (0-15), giving 16 unique IDs per timestamp tick.
//   - On Windows the system timer resolution is ~1 ms, so UnixMicro returns
//     the same value for every call within the same millisecond.
//   - Once the counter exceeds 15 it would overflow and wrap back to 0,
//     producing the exact same composite value and therefore a duplicate ID.
//   - Sleeping 1 ms guarantees the timestamp advances to the next tick.
//   - Packs timestamp and counter into a single int64: (ts << 4) | counter.
//   - The timestamp occupies the high bits; the counter occupies the low 4 bits.
//   - Encodes the composite value using Crockford Base32 alphabet.
//   - Converts the result to lowercase.
//   - Returns an 11-character string (e.g. "sa4rc789wxg").
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

	if counter > 15 {
		time.Sleep(1 * time.Millisecond)
		ts = time.Now().UnixMicro()
		lastTimestamp = ts
		counter = 0
	}

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
