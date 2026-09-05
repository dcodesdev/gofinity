package main

import (
	"errors"
	"fmt"
)

// The two conditions a caller of this package can branch on.
var (
	ErrNotFound = errors.New("not found")
	ErrInvalid  = errors.New("invalid")
)

// store is the fake data layer. It is written for you.
var store = map[string]string{
	"alice": "42",
	"bob":   "not-a-number",
	"carol": "-1",
}

// fetch returns the raw record for name, or ErrNotFound. Written for you: this
// is the bottom of the stack, where the sentinel is produced.
func fetch(name string) (string, error) {
	raw, ok := store[name]
	if !ok {
		return "", ErrNotFound
	}
	return raw, nil
}

// parseScore turns a raw record into a score. A record that is not a number, or
// is negative, is ErrInvalid wrapped with the offending text:
//
//	parseScore("not-a-number") -> `parse "not-a-number": invalid`
//	parseScore("-1")           -> `parse "-1": invalid`
//
// Use fmt.Errorf with %w so errors.Is(err, ErrInvalid) holds, and %q for the
// raw text. Parse the digits by hand - strconv is the next lesson's business,
// and an all-digit string is all this needs.
func parseScore(raw string) (int, error) {
	// TODO
	return 0, nil
}

// Score fetches a record and parses it, adding the name as context at each
// step. Both failures come back as:
//
//	Score("nobody") -> `score "nobody": not found`
//	Score("bob")    -> `score "bob": parse "not-a-number": invalid`
//
// The chain must stay intact: errors.Is(err, ErrNotFound) and
// errors.Is(err, ErrInvalid) both hold for the right inputs.
func Score(name string) (int, error) {
	// TODO
	return 0, nil
}

// Reason maps any error from Score onto a short word for a UI:
//
//	nil        -> "ok"
//	ErrNotFound anywhere in the chain -> "missing"
//	ErrInvalid  anywhere in the chain -> "bad-data"
//	anything else                     -> "unknown"
//
// Compare with errors.Is, never with == : by the time an error reaches here it
// has been wrapped twice and equals neither sentinel.
func Reason(err error) string {
	// TODO
	return ""
}

// Unwrapped walks err all the way down with errors.Unwrap and returns the
// innermost error - the one that does not wrap anything. It returns nil for a
// nil error.
func Unwrapped(err error) error {
	// TODO
	return nil
}

// Depth counts how many wrappers sit on top of the innermost error. A nil error
// is 0, an unwrapped sentinel is 0, and `score "bob": parse "...": invalid` is 2.
func Depth(err error) int {
	// TODO
	return 0
}

// Total sums the scores for every name, in order, and stops at the first
// failure. The error it returns wraps whatever Score gave it, prefixed with the
// position that failed:
//
//	Total([]string{"alice", "nobody"}) -> `entry 1: score "nobody": not found`
//
// Positions are zero-based, matching the index in names.
func Total(names []string) (int, error) {
	// TODO
	return 0, nil
}

func main() {
	_, err := Score("bob")
	fmt.Println(err)          // score "bob": parse "not-a-number": invalid
	fmt.Println(Reason(err))  // bad-data
	fmt.Println(Unwrapped(err)) // invalid
}
