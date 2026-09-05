package main

import "fmt"

// TB is the part of *testing.T a table runner actually uses. A helper that
// takes this interface instead of *testing.T can be tested like any other
// function: the tests in this challenge pass it a recorder and read back what
// it was told.
//
// *testing.T satisfies it, so a real test still calls RunCases(t, ...).
type TB interface {
	// Helper marks the caller as a helper, so a failure is reported at the
	// line of the code that called it rather than inside it.
	Helper()
	// Errorf records a failure and lets the test carry on.
	Errorf(format string, args ...any)
}

// Case is one row of the table: a name, the inputs, and the expected result.
// The name is what a reader sees when the row fails, so it names the situation
// ("below the range"), not the numbers.
type Case struct {
	Name string
	N    int
	Lo   int
	Hi   int
	Want int
}

// Clamp returns n limited to the closed range [lo, hi]: lo when n is below it,
// hi when n is above it, and n itself otherwise. Callers guarantee lo <= hi.
func Clamp(n, lo, hi int) int {
	if n < lo {
		return lo
	}
	if n > hi {
		return hi
	}
	return n
}

// ClampCases returns the table that exercises Clamp. Each row names a
// situation rather than its numbers, so a failure reads as a sentence.
func ClampCases() []Case {
	return []Case{
		{Name: "below the range", N: -4, Lo: 0, Hi: 10, Want: 0},
		{Name: "above the range", N: 40, Lo: 0, Hi: 10, Want: 10},
		{Name: "strictly inside", N: 7, Lo: 0, Hi: 10, Want: 7},
		{Name: "exactly at lo", N: 0, Lo: 0, Hi: 10, Want: 0},
		{Name: "exactly at hi", N: 10, Lo: 0, Hi: 10, Want: 10},
		{Name: "inside a negative range", N: -5, Lo: -8, Hi: -3, Want: -5},
		{Name: "below a negative range", N: -20, Lo: -8, Hi: -3, Want: -8},
		{Name: "single value range", N: 9, Lo: 4, Hi: 4, Want: 4},
	}
}

// RunCases runs fn over every case and reports the ones that disagree with
// Want. Every row runs: the value of a table is that one run tells you about
// all of them.
func RunCases(t TB, cases []Case, fn func(Case) int) {
	t.Helper()
	for _, c := range cases {
		got := fn(c)
		if got != c.Want {
			t.Errorf("clamp: %s = got %d, want %d", c.Name, got, c.Want)
		}
	}
}

func main() {
	for _, c := range ClampCases() {
		fmt.Printf("%-24s Clamp(%d, %d, %d) = %d\n", c.Name, c.N, c.Lo, c.Hi, c.Want)
	}
}
