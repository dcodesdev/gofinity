package main

import (
	"errors"
	"fmt"
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
	// TODO
	return ""
}

// Unwrap returns the sentinel underneath, which is what makes
// errors.Is(err, ErrNotNumeric) work on a *ParseError.
func (e *ParseError) Unwrap() error {
	// TODO
	return nil
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
	// TODO
	return "", 0, nil
}

// ParseAll parses every line, numbering them from 1 and skipping empty ones. It
// does not stop at the first failure: it collects every error and returns them
// joined with errors.Join, and a nil map. On success it returns the map and a
// nil error.
func ParseAll(lines []string) (map[string]int, error) {
	// TODO
	return nil, nil
}

// LineOf returns the line number of the first *ParseError anywhere in err, or 0
// when there is none. Use errors.As - err may be a join, a wrap, or the
// *ParseError itself.
func LineOf(err error) int {
	// TODO
	return 0
}

// Lines returns the line numbers of every *ParseError in err, in the order they
// were joined. errors.As finds only the first, so this one walks the tree
// itself: an error may implement `Unwrap() error` (one child) or
// `Unwrap() []error` (several, which is what errors.Join returns).
func Lines(err error) []int {
	// TODO
	return nil
}

// Explain maps err onto one word, checking the sentinels in this order:
//
//	nil             -> "ok"
//	ErrMalformed    -> "malformed"
//	ErrMissingValue -> "missing-value"
//	ErrNotNumeric   -> "not-numeric"
//	anything else   -> "unknown"
func Explain(err error) string {
	// TODO
	return ""
}

func main() {
	cfg, err := ParseAll([]string{"port=8080", "oops", "host="})
	fmt.Println(cfg, err)
	fmt.Println(Lines(err), Explain(err))
}
