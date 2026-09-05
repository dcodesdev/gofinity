package main

import (
	"fmt"
	"testing"
)

// TB is the part of *testing.T these helpers use. Taking an interface rather
// than *testing.T is what makes a test helper testable: the tests in this
// challenge pass a recorder and read back what the helper reported.
//
// *testing.T satisfies it, so `Equal(t, got, want)` still works in a real test.
type TB interface {
	Helper()
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	Cleanup(func())
	Name() string
}

// Suite is a TB that can also start a subtest. *testing.T cannot satisfy this
// - its Run takes a func(*testing.T) - which is what GoSuite below is for.
type Suite interface {
	TB
	// Run starts a subtest named name and reports whether it passed.
	Run(name string, f func(s Suite)) bool
}

// Case is one row of a table.
type Case struct {
	Name string
	In   int
	Want int
}

// Equal reports a failure when got and want differ, and nothing at all when
// they match. The message must show both values.
//
// It is a helper, so it marks itself as one: the failure belongs to the line
// that called Equal, not to the line inside it.
func Equal[V comparable](t TB, got, want V) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

// Must returns v when err is nil, and stops the test with Fatalf when it is
// not. The message must contain err's text.
//
// Use it for the setup a test cannot continue without: the parsed fixture, the
// opened file. It must not return when err is non-nil - Fatalf does not return
// either.
func Must[V any](t TB, v V, err error) V {
	t.Helper()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	return v
}

// WithResource opens a resource, arranges for closeFn to run when the test
// that called it finishes, and returns the resource.
//
// The point is that the caller writes one line and cannot forget the teardown,
// and that the teardown still runs when the test fails or calls Fatalf - which
// a `defer` in the caller would too, but a `defer` in a *helper* would not: it
// would run the moment the helper returned.
func WithResource(t TB, open func() string, closeFn func(string)) string {
	t.Helper()
	resource := open()
	t.Cleanup(func() { closeFn(resource) })
	return resource
}

// RunTable runs each case as its own subtest named after the case, comparing
// fn's result with Want using Equal.
//
// Each case gets its own subtest, so a failure names the row, one row's
// Fatalf does not abandon the others, and `go test -run 'TestX/below'` runs
// exactly one of them.
//
// An empty table is a failure on s itself: a table-driven test with no rows
// passes without testing anything, and that is the one thing a test must never
// do quietly.
func RunTable(s Suite, cases []Case, fn func(int) int) {
	s.Helper()
	if len(cases) == 0 {
		s.Errorf("the table is empty: %s would pass without testing anything", s.Name())
		return
	}
	for _, c := range cases {
		s.Run(c.Name, func(sub Suite) {
			Equal(sub, fn(c.In), c.Want)
		})
	}
}

// GoSuite adapts *testing.T to Suite: it is the whole adapter, and it is four
// lines, because *testing.T already has every other method Suite needs.
type GoSuite struct {
	*testing.T
}

// Run starts a real subtest and hands the *subtest's* T to f, wrapped in a
// GoSuite of its own. Handing over the parent instead would make every failure
// belong to the parent test and every s.Name() the same.
func (g GoSuite) Run(name string, f func(s Suite)) bool {
	return g.T.Run(name, func(t *testing.T) {
		f(GoSuite{t})
	})
}

func main() {
	double := func(n int) int { return n * 2 }
	for _, c := range []Case{{"zero", 0, 0}, {"three", 3, 6}} {
		fmt.Printf("%s: double(%d) = %d, want %d\n", c.Name, c.In, double(c.In), c.Want)
	}
}
