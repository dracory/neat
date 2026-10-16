package uid

import (
	"strings"
	"sync/atomic"
	"time"
)

// Package-level state for the short ID generators.
var (
	// shortIDAtomicState packs (microsecond timestamp << 4) | 4-bit counter.
	shortIDAtomicState atomic.Uint64

	// nowMicro is the clock used by the generators. It is a variable so tests
	// can inject controlled timestamps (e.g. backwards clock steps).
	// Tests that override it MUST NOT use t.Parallel() and must restore it
	// (and shortIDAtomicState) afterwards.
	nowMicro = func() int64 { return time.Now().UnixMicro() }
)

const (
	// safeShortAlphabet is a vowel-free Base 16 alphabet. Every letter of
	// standard hex is shifted to the next sequential consonant
	// ('a'->'b', 'b'->'c', 'c'->'d', 'd'->'f', 'e'->'g', 'f'->'h') to avoid
	// generating accidental English words; this greatly reduces but does not
	// eliminate them (leet-speak like "b00b" is still reachable). The alphabet
	// is in ascending byte order, so generated IDs sort chronologically.
	safeShortAlphabet = "0123456789bcdfgh"
	// SafeShortEncodedLength is the base16 digit count of the 55-bit composite.
	SafeShortEncodedLength = 14

	crockfordLowerAlphabet = "0123456789abcdefghjkmnpqrstvwxyz"
	// CrockfordEncodedLength is the base32 digit count of the 55-bit composite.
	CrockfordEncodedLength = 11

	// nanoidAlphabet is the URL-safe alphabet used by 21-character NanoID-style IDs.
	nanoidAlphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-"
	// NanoIDEncodedLength is the character count of a NanoID-style ID.
	NanoIDEncodedLength = 21
)

// GenerateShortID creates a new 14-character lowercase safe short ID.
// It encodes the current microsecond timestamp using the vowel-free
// Base 16 alphabet. Thread-safe via a lock-free CAS loop; survives
// backwards clock steps.
//
// Business Logic:
//   - Uses time.Now().UnixMicro() as the uniqueness base.
//   - Maintains a package-level atomic (timestamp << 4) | counter state.
//   - Counter increments while the timestamp does not advance; when it would
//     exceed 15, the timestamp is bumped by one microsecond instead of
//     sleeping, so callers never stall and IDs never regress.
//   - Packs timestamp and counter into a single int64: (ts << 4) | counter.
//   - Encodes the composite value with the vowel-free Base 16 alphabet.
//
// Notes:
//   - Uniqueness is per-process only: two processes generating IDs in the
//     same microsecond can collide. The IDs are also predictable and reveal
//     their creation time, so they are suitable for sortable database keys
//     but NOT for security tokens or unguessable public identifiers.
//
// Returns:
//   - A 14-character lowercase string (e.g. "05fh88gd139f8h").
func GenerateShortID() string {
	return encodeHexFixed(nextShortIDValue(), safeShortAlphabet, SafeShortEncodedLength)
}

// SafeShortID creates a new 14-character lowercase safe short ID.
// It is an alias for GenerateShortID.
func SafeShortID() string {
	return GenerateShortID()
}

// nextShortIDValue atomically generates a unique 55-bit composite value
// combining a microsecond timestamp (shifted left 4 bits) and a 4-bit
// rolling counter (0-15). Uses a CAS loop for thread safety. If the clock
// moves backwards the timestamp is held at its last value and the counter
// keeps incrementing; on counter overflow the timestamp is bumped forward
// by one microsecond rather than sleeping.
func nextShortIDValue() int64 {
	for {
		now := nowMicro()
		oldState := shortIDAtomicState.Load()
		oldTS := int64(oldState >> 4)
		oldCnt := int64(oldState & 0x0F)

		var newTS, newCnt int64
		if now > oldTS {
			newTS = now
			newCnt = 0
		} else if oldCnt < 0x0F {
			newTS = oldTS
			newCnt = oldCnt + 1
		} else {
			newTS = oldTS + 1
			newCnt = 0
		}

		newState := uint64(newTS)<<4 | uint64(newCnt)
		if shortIDAtomicState.CompareAndSwap(oldState, newState) {
			return int64(newState)
		}
	}
}

// encodeHexFixed encodes val into exactly width lowercase hex-style digits,
// most-significant digit first. It only works for 16-character alphabets
// (it shifts by 4 bits per digit) and width must be <= 14.
func encodeHexFixed(val int64, alphabet string, width int) string {
	var buf [SafeShortEncodedLength]byte
	for i := width - 1; i >= 0; i-- {
		buf[i] = alphabet[val&0x0f]
		val >>= 4
	}
	return string(buf[:width])
}

// NormalizeID normalizes an ID to lowercase for consistent lookups.
func NormalizeID(id string) string {
	return strings.ToLower(strings.TrimSpace(id))
}

// IsShortID checks whether the given string looks like a generated short ID:
// the correct length AND only characters from the alphabet of the matching
// scheme — 11 chars use the Crockford Base32 alphabet, 14 chars use the
// vowel-free Base 16 safe alphabet, and 21 chars use the NanoID alphabet.
//
// The check is case-sensitive: the Crockford and safe alphabets are lowercase
// only, so uppercase input is rejected. Callers handling user-supplied input
// should pass it through NormalizeID first.
func IsShortID(id string) bool {
	length := len(id)
	var alphabet string
	switch length {
	case CrockfordEncodedLength:
		alphabet = crockfordLowerAlphabet
	case SafeShortEncodedLength:
		alphabet = safeShortAlphabet
	case NanoIDEncodedLength:
		alphabet = nanoidAlphabet
	default:
		return false
	}

	for i := 0; i < length; i++ {
		if strings.IndexByte(alphabet, id[i]) < 0 {
			return false
		}
	}
	return true
}
