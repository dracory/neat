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
// Key Advantages (Pros) of support/uid:
//
//  1. Zero Randomness / No System Entropy Overhead:
//     UUID v7 requires cryptographically secure pseudorandom number generation (crypto/rand)
//     to fill ~62 bits on every generation call. support/uid is fully deterministic based
//     on timestamp and sequence counter, eliminating CSPRNG overhead and avoiding system
//     entropy consumption under high-volume ID generation.
//
//  2. Sub-Millisecond Precision & True Sub-Tick Ordering:
//     UUID v7 records timestamps with millisecond resolution; identifiers created within
//     the same millisecond rely on random bits for relative ordering. support/uid uses
//     microsecond resolution combined with an explicit sequence counter (0–15) that increments
//     on ties, guaranteeing exact creation-order monotonicity for up to 16 IDs per microsecond.
//
//  3. Clean, Compact, URL and Filename-Safe Encoding:
//     Crockford Base32 produces an 11-character lowercase alphanumeric string without hyphens
//     or special characters. It can be embedded directly into URLs, database keys, filenames,
//     or slugs without URL escaping or string normalization. Standard UUID strings are 36
//     characters long and contain hyphens.
//
//  4. Built-in Normalization Helper:
//     The NormalizeID function provides standard whitespace trimming and lowercasing for
//     case-insensitive lookups, addressing common string ID handling requirements out of the box.
//
//  5. Minimal Dependency Footprint:
//     The package depends only on standard library primitives (strings, sync, time), making
//     it concise, fast, and easy to audit end-to-end.
//
// Key Trade-offs (Cons):
//
//  1. Single-Process / Single-Instance Scope:
//     Uniqueness depends on in-memory state (lastTimestamp and counter) protected by a mutex.
//     It is designed for single-process architectures. Unlike UUID v7, which uses ~62 bits of
//     entropy to achieve global collision resistance across distributed nodes without coordination,
//     support/uid requires external coordination if generated across multiple independent processes.
//
//  2. Burst Throttling under Extreme Concurrency:
//     If more than 16 IDs are generated within a single microsecond, the generator pauses execution
//     via time.Sleep(1 * time.Millisecond) to allow the clock tick to advance and prevent counter wrap-around.
package uid
