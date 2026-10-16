// Package uid provides functions for generating short, time-ordered, URL-safe
// unique identifiers and normalizing/validating ID strings.
//
// It delegates short ID generation (GenerateShortID, SafeShortID) to
// github.com/dracory/uid.SafeShortID(), which produces fixed 14-character
// lowercase short IDs using a vowel-free Base 16 alphabet ("0123456789bcdfrt").
package uid
