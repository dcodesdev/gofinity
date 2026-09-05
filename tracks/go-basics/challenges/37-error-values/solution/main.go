package main

import (
	"errors"
	"fmt"
)

// ErrEmpty is a sentinel: one error value, created once at package level, that
// callers can compare against. Sentinels are exported, named Err..., and built
// with errors.New - never inside a function, or every call makes a new one that
// compares equal to nothing.
var ErrEmpty = errors.New("empty input")

// ErrNotFound is the second sentinel this file needs. Declare it with the
// message "not found".
var ErrNotFound = errors.New("not found")

// Divide returns a / b. Division by zero is not a panic here, it is an error:
// return 0 and an error built with fmt.Errorf saying "divide 7 by zero" for
// Divide(7, 0). The value beside a non-nil error is the zero value, and callers
// must not read it.
func Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("divide %d by zero", a)
	}
	return a / b, nil
}

// First returns the first element of s. An empty slice is ErrEmpty itself -
// return the sentinel, not a copy of its message, or a caller comparing with
// errors.Is has nothing to match.
func First(s []int) (int, error) {
	if len(s) == 0 {
		return 0, ErrEmpty
	}
	return s[0], nil
}

// Lookup returns the value stored under key. A missing key returns ErrNotFound.
func Lookup(m map[string]int, key string) (int, error) {
	// The comma-ok form, so a stored 0 is a hit rather than a miss.
	v, ok := m[key]
	if !ok {
		return 0, ErrNotFound
	}
	return v, nil
}

// SumQuotients divides total by each divisor in turn and adds the results,
// stopping at the first error and returning it. On success it returns the sum
// and a nil error. An empty divisors slice is 0 and no error, not ErrEmpty:
// there is nothing wrong with adding nothing.
func SumQuotients(total int, divisors []int) (int, error) {
	sum := 0
	for _, d := range divisors {
		q, err := Divide(total, d)
		if err != nil {
			// Nothing partial escapes: the zero value goes back with the error.
			return 0, err
		}
		sum += q
	}
	return sum, nil
}

// Describe turns a result and an error into one line for a human:
//
//	err == nil          -> "ok: 12"
//	err != nil          -> "failed: divide 7 by zero"
//
// It never reads the value when the error is non-nil.
func Describe(value int, err error) string {
	if err != nil {
		return "failed: " + err.Error()
	}
	return fmt.Sprintf("ok: %d", value)
}

func main() {
	fmt.Println(Describe(Divide(84, 7)))
	fmt.Println(Describe(Divide(7, 0)))
}
