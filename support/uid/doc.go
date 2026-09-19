// Package uid provides functions for generating short, time-ordered, URL-safe
// unique identifiers and normalizing/validating ID strings.
//
// # Architecture & Mechanism
//
// The primary generator, GenerateShortID, packs a microsecond Unix timestamp
// and a 4-bit sequence counter into a single 64-bit integer:
//
//	composite = (ts << 4) | counter
//
// The composite value is then encoded into Crockford Base32 (using the 32-character
// alphabet "0123456789abcdefghjkmnpqrstvwxyz" which excludes ambiguous characters like
// I, L, O, and U) and lowercased to yield an 11-character string (e.g. "sa4rc789wxg").
// State updates (timestamp tracking and counter incrementing) are guarded by a package-level
// mutex.
//
// # Comparison: support/uid vs Go 1.27's Standard Library uuid (UUID v7)
//
// Go 1.27 introduces a standard library uuid package generating 128-bit UUID v7
// identifiers (combining a 48-bit millisecond timestamp with ~62 bits of CSPRNG randomness,
// formatted as standard 36-character hyphenated strings).
//
// While UUID v7 is the right choice when generating IDs across independent, uncoordinated
// machines or microservices, support/uid's design offers significant advantages across three
// primary pillars:
//
// 1. Zero Randomness / No System Entropy Overhead:
//
//   - UUID v7 requires cryptographically secure pseudorandom number generation (crypto/rand)
//     to fill ~62 bits on every generation call.
//   - support/uid is fully deterministic based on timestamp and sequence counter, eliminating
//     CSPRNG calls and system entropy consumption under high-volume ID generation.
//
// 2. Sub-Millisecond Precision & Sub-Tick Ordering:
//
//   - UUID v7 encodes time down to millisecond resolution; IDs generated within the same
//     millisecond are sorted by random bits rather than exact creation sequence.
//   - support/uid uses microsecond timestamps paired with an explicit counter (0–15) that
//     increments on ties, guaranteeing strict creation-order monotonicity for up to 16 IDs
//     per microsecond.
//
// 3. Human-Optimized Ergonomics & URL/Filename Safety:
//
//   - Compact Length: An 11-character ID is dramatically easier to read over the phone,
//     retype from a support ticket, or eyeball in log files compared to a 36-character UUID string.
//   - No Visual Ambiguity: Crockford Base32 removes ambiguous characters (I, L, O, U),
//     eliminating 1/I/l and 0/O confusion that causes human transcription errors.
//   - Case-Insensitive & Normalization-Friendly: Designed to be case-insensitive. The
//     NormalizeID helper ensures that habituated capitalization or uppercase pipeline steps
//     do not break database lookups.
//   - Hyphen-Free: Free of hyphens, eliminating copy-paste failure modes caused by misplaced
//     or missing dashes in standard UUIDs (e.g., xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx).
//
// 4. Minimal Dependency Footprint:
//
//   - The package depends exclusively on standard library primitives (strings, sync, time),
//     making it lightweight, fast, and trivial to audit end-to-end.
//
// # Comparison: support/uid vs xid (github.com/rs/xid)
//
// xid is a popular 12-byte identifier composed of a 4-byte timestamp (seconds), a 3-byte
// machine identifier, a 2-byte process ID, and a 3-byte counter, encoded as a 20-character
// base32hex string.
//
// Where support/uid Wins:
//
//   - Compactness: 11 characters vs xid's 20 characters. support/uid is significantly shorter.
//   - Simplicity: ~80 lines of pure standard library Go code with zero platform-specific files
//     (unlike xid which uses OS-specific hostid implementations like hostid_linux.go, hostid_darwin.go).
//
// Where xid Wins:
//
//   - Uncoordinated Multi-Host Safety: xid bakes the machine ID and process ID directly into
//     every ID via hostid resolution and os.Getpid(), allowing multiple replicas or pods to
//     generate IDs safely out-of-the-box without central coordination.
//   - Lock-Free Concurrency: xid uses atomic counter operations instead of a sync.Mutex,
//     scaling better across multiple CPU cores under extremely heavy concurrent generation.
//   - Guaranteed Per-Host Throughput: xid guarantees uniqueness for up to 16,777,216 IDs per second
//     per host/process via its 24-bit counter.
//
// # Summary & Trade-offs (Cons):
//
// 1. Single-Process / Single-Instance Scope:
//
//   - Uniqueness relies on in-memory state (lastTimestamp and counter) guarded by a mutex.
//   - Unlike UUID v7 (entropy) or xid (host ID + PID), support/uid requires external
//     coordination if generated across multiple independent processes or pods.
//
// 2. Burst Throttling under Extreme Concurrency:
//
//   - If more than 16 IDs are requested within a single microsecond, the generator pauses
//     execution via time.Sleep(1 * time.Millisecond) to advance the clock tick and prevent
//     counter overflow/duplicates.
package uid
