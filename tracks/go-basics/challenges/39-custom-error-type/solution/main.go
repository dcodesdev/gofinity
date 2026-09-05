package main

import (
	"errors"
	"fmt"
	"strings"
)

// The three ways a config line can be wrong. These are the conditions; the
// *ParseError below carries the detail.
var (
	ErrMalformed    = errors.New("missing '='")
	ErrMissingValue = errors.New("missing value")
	ErrNotNumeric   = errors.New("not numeric")
)

// ParseError is a custom error type: a struct that satisfies error and carries
// the fields a caller can act on. A message is for a human; Line and Key are
// for code.
type ParseError struct {
	Line int    // 1-based, as a person counts lines in a file
	Key  string // empty when the line was too broken to have one
	Err  error  // one of the sentinels above
}

// Error renders the struct. With a key:
//
//	line 3: key "host": missing value
//
// Without one:
//
//	line 2: missing '='
//
// The receiver is a pointer, so *ParseError is what satisfies error.
func (e *ParseError) Error() string {
	if e.Key == "" {
		return fmt.Sprintf("line %d: %v", e.Line, e.Err)
	}
	return fmt.Sprintf("line %d: key %q: %v", e.Line, e.Key, e.Err)
}

// Unwrap returns the sentinel underneath, which is what makes
// errors.Is(err, ErrNotNumeric) work on a *ParseError.
func (e *ParseError) Unwrap() error {
	return e.Err
}

// ParseLine parses `key=value`, where value is a run of digits. n is the
// 1-based line number, used only for the error.
//
//	"port=8080" -> ("port", 8080, nil)
//	"oops"      -> ParseError{Line: n, Err: ErrMalformed}
//	"=8080"     -> ParseError{Line: n, Err: ErrMalformed}
//	"host="     -> ParseError{Line: n, Key: "host", Err: ErrMissingValue}
//	"n=abc"     -> ParseError{Line: n, Key: "n", Err: ErrNotNumeric}
//
// Return a bare nil for the error on success. A nil *ParseError returned as an
// error is not a nil error.
func ParseLine(n int, line string) (string, int, error) {
	key, value, found := strings.Cut(line, "=")
	if !found || key == "" {
		return "", 0, &ParseError{Line: n, Err: ErrMalformed}
	}
	if value == "" {
		return "", 0, &ParseError{Line: n, Key: key, Err: ErrMissingValue}
	}
	n2 := 0
	for _, r := range value {
		if r < '0' || r > '9' {
			return "", 0, &ParseError{Line: n, Key: key, Err: ErrNotNumeric}
		}
		n2 = n2*10 + int(r-'0')
	}
	// A bare nil, so the caller's `err != nil` is false.
	return key, n2, nil
}

// ParseAll parses every line, numbering them from 1 and skipping empty ones. It
// does not stop at the first failure: it collects every error and returns them
// joined with errors.Join, and a nil map. On success it returns the map and a
// nil error.
func ParseAll(lines []string) (map[string]int, error) {
	cfg := make(map[string]int, len(lines))
	var errs []error
	for i, line := range lines {
		if line == "" {
			continue
		}
		key, value, err := ParseLine(i+1, line)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		cfg[key] = value
	}
	if len(errs) > 0 {
		// errors.Join keeps every branch reachable by errors.Is and errors.As.
		return nil, errors.Join(errs...)
	}
	return cfg, nil
}

// LineOf returns the line number of the first *ParseError anywhere in err, or 0
// when there is none. Use errors.As - err may be a join, a wrap, or the
// *ParseError itself.
func LineOf(err error) int {
	var pe *ParseError
	if errors.As(err, &pe) {
		return pe.Line
	}
	return 0
}

// Lines returns the line numbers of every *ParseError in err, in the order they
// were joined. errors.As finds only the first, so this one walks the tree
// itself: an error may implement `Unwrap() error` (one child) or
// `Unwrap() []error` (several, which is what errors.Join returns).
func Lines(err error) []int {
	var out []int
	var walk func(error)
	walk = func(err error) {
		if err == nil {
			return
		}
		if pe, ok := err.(*ParseError); ok {
			out = append(out, pe.Line)
			return
		}
		switch u := err.(type) {
		case interface{ Unwrap() []error }:
			for _, child := range u.Unwrap() {
				walk(child)
			}
		case interface{ Unwrap() error }:
			walk(u.Unwrap())
		}
	}
	walk(err)
	return out
}

// Explain maps err onto one word, checking the sentinels in this order:
//
//	nil             -> "ok"
//	ErrMalformed    -> "malformed"
//	ErrMissingValue -> "missing-value"
//	ErrNotNumeric   -> "not-numeric"
//	anything else   -> "unknown"
func Explain(err error) string {
	switch {
	case err == nil:
		return "ok"
	case errors.Is(err, ErrMalformed):
		return "malformed"
	case errors.Is(err, ErrMissingValue):
		return "missing-value"
	case errors.Is(err, ErrNotNumeric):
		return "not-numeric"
	default:
		return "unknown"
	}
}

func main() {
	cfg, err := ParseAll([]string{"port=8080", "oops", "host="})
	fmt.Println(cfg, err)
	fmt.Println(Lines(err), Explain(err))
}
