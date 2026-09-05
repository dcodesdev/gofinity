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
	// TODO
	return 0
}

// ClampCases returns the table that exercises Clamp.
//
// Every case's Want must be the correct answer - a table with a wrong
// expectation is worse than no table at all. Names must be unique and
// non-empty, and between them the rows must cover all seven situations:
//
//	n below lo, n above hi, n strictly inside the range,
//	n exactly lo, n exactly hi, a range that is entirely negative,
//	and a range where lo == hi.
//
// One row may cover more than one of those.
func ClampCases() []Case {
	// TODO
	return nil
}

// RunCases runs fn over every case and reports the ones that disagree with
// Want.
//
// It runs every case: a table that stops at the first failure tells you about
// one row when it could have told you about four. Passing rows report nothing
// at all.
//
// A failure message must name the case and show both values, because "false is
// not true" is not a bug report:
//
//	clamp: Name = got 7, want 5
//
// Call t.Helper() before reporting, so the failure is blamed on the caller's
// line rather than on this function.
func RunCases(t TB, cases []Case, fn func(Case) int) {
	// TODO
}

func main() {
	for _, c := range ClampCases() {
		fmt.Printf("%-24s Clamp(%d, %d, %d) = %d\n", c.Name, c.N, c.Lo, c.Hi, c.Want)
	}
}
