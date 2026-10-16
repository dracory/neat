// Package uid provides functions for generating short, time-ordered, URL-safe
// unique identifiers and normalizing/validating ID strings.
//
// GenerateShortID and SafeShortID produce fixed 14-character lowercase short
// IDs using a vowel-free Base 16 alphabet ("0123456789bcdfgh"). The underlying
// composite value packs a microsecond Unix timestamp and a 4-bit sequence
// counter, so generated IDs sort chronologically. State updates use a lock-free
// atomic CAS loop — no mutex, no sleeping, and backwards clock steps are
// tolerated.
//
// Note: uniqueness is guaranteed per-process only. Two processes generating
// IDs in the same microsecond can collide, so multi-host deployments require
// external coordination.
package uid
