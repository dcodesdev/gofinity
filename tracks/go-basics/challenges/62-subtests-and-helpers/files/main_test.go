package main

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// fake is a Suite that records what it was told instead of failing anything.
// Fatalf ends the calling goroutine, exactly as *testing.T's does, so a helper
// that "returns after Fatalf" is caught here rather than in production.
type fake struct {
	name     string
	errors   []string
	fatals   []string
	helpers  int
	cleanups []func()
	subs     []*fake
}

func (f *fake) Helper() { f.helpers++ }

func (f *fake) Errorf(format string, args ...any) {
	f.errors = append(f.errors, fmt.Sprintf(format, args...))
}

func (f *fake) Fatalf(format string, args ...any) {
	f.fatals = append(f.fatals, fmt.Sprintf(format, args...))
	runtime.Goexit()
}

func (f *fake) Cleanup(fn func()) { f.cleanups = append(f.cleanups, fn) }

func (f *fake) Name() string { return f.name }

func (f *fake) Run(name string, body func(s Suite)) bool {
	sub := &fake{name: f.name + "/" + name}
	f.subs = append(f.subs, sub)
	done := make(chan struct{})
	go func() {
		defer close(done)
		body(sub)
	}()
	<-done
	sub.runCleanups()
	return len(sub.errors) == 0 && len(sub.fatals) == 0
}

// runCleanups runs them last registered first, like testing does.
func (f *fake) runCleanups() {
	for i := len(f.cleanups) - 1; i >= 0; i-- {
		f.cleanups[i]()
	}
	f.cleanups = nil
}

func (f *fake) failures() []string { return append(append([]string{}, f.errors...), f.fatals...) }

// inGoroutine runs body and reports whether it ran to the end. A helper that
// called Fatalf must not have.
func inGoroutine(body func()) (finished bool) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		body()
		finished = true
	}()
	<-done
	return finished
}

func TestEqualSaysNothingWhenValuesMatch(t *testing.T) {
	f := &fake{name: "match"}
	Equal(f, 7, 7)
	Equal(f, "same", "same")
	if got := f.failures(); len(got) != 0 {
		t.Errorf("equal values reported %v", got)
	}
}

func TestEqualReportsBothValues(t *testing.T) {
	f := &fake{name: "mismatch"}
	Equal(f, 41, 42)
	if len(f.fatals) != 0 {
		t.Errorf("Equal used Fatalf %v, want Errorf so the rest of the test still runs", f.fatals)
	}
	if len(f.errors) != 1 {
		t.Fatalf("got %d failures, want exactly 1: %v", len(f.errors), f.errors)
	}
	msg := f.errors[0]
	for _, want := range []string{"41", "42"} {
		if !strings.Contains(msg, want) {
			t.Errorf("failure %q does not mention %s - a message needs both the got and the want", msg, want)
		}
	}
	if f.helpers == 0 {
		t.Error("Equal never called t.Helper(), so every failure points at the same line inside Equal")
	}
}

func TestEqualWorksForStrings(t *testing.T) {
	f := &fake{name: "strings"}
	Equal(f, "alpha", "omega")
	if len(f.errors) != 1 {
		t.Fatalf("got %d failures, want 1: %v", len(f.errors), f.errors)
	}
	for _, want := range []string{"alpha", "omega"} {
		if !strings.Contains(f.errors[0], want) {
			t.Errorf("failure %q does not mention %q", f.errors[0], want)
		}
	}
}

func TestMustReturnsTheValue(t *testing.T) {
	f := &fake{name: "must-ok"}
	var got int
	finished := inGoroutine(func() { got = Must(f, 12, nil) })
	if !finished {
		t.Fatal("Must stopped the test even though err was nil")
	}
	if got != 12 {
		t.Errorf("Must returned %d, want 12", got)
	}
	if fails := f.failures(); len(fails) != 0 {
		t.Errorf("a nil error reported %v", fails)
	}
}

func TestMustStopsOnAnError(t *testing.T) {
	f := &fake{name: "must-fail"}
	err := errors.New("fixture would not open")
	finished := inGoroutine(func() { _ = Must(f, "unused", err) })
	if finished {
		t.Error("Must returned after reporting the error - Fatalf does not return, and neither may a helper that calls it")
	}
	if len(f.errors) != 0 {
		t.Errorf("Must used Errorf %v, want Fatalf: the test cannot continue without the value", f.errors)
	}
	if len(f.fatals) != 1 {
		t.Fatalf("got %d fatals, want 1: %v", len(f.fatals), f.fatals)
	}
	if !strings.Contains(f.fatals[0], "fixture would not open") {
		t.Errorf("failure %q does not contain the error's own text", f.fatals[0])
	}
	if f.helpers == 0 {
		t.Error("Must never called t.Helper()")
	}
}

func TestWithResourceClosesAtTheEndOfTheTest(t *testing.T) {
	f := &fake{name: "resource"}
	closed := []string{}
	got := WithResource(f, func() string { return "conn-1" }, func(v string) { closed = append(closed, v) })
	if got != "conn-1" {
		t.Errorf("WithResource returned %q, want the value open produced", got)
	}
	if len(closed) != 0 {
		t.Fatalf("the resource was closed before the test finished: %v", closed)
	}
	if len(f.cleanups) != 1 {
		t.Fatalf("registered %d cleanups, want exactly 1", len(f.cleanups))
	}
	f.runCleanups()
	if len(closed) != 1 || closed[0] != "conn-1" {
		t.Errorf("cleanup closed %v, want [conn-1]", closed)
	}
	if f.helpers == 0 {
		t.Error("WithResource never called t.Helper()")
	}
}

func TestWithResourceCleanupsRunLastFirst(t *testing.T) {
	f := &fake{name: "nested"}
	closed := []string{}
	record := func(v string) { closed = append(closed, v) }
	WithResource(f, func() string { return "outer" }, record)
	WithResource(f, func() string { return "inner" }, record)
	f.runCleanups()
	if len(closed) != 2 || closed[0] != "inner" || closed[1] != "outer" {
		t.Errorf("closed %v, want [inner outer] - cleanups unwind in reverse, like defer", closed)
	}
}

func TestRunTableRunsOneSubtestPerCase(t *testing.T) {
	f := &fake{name: "Table"}
	cases := []Case{
		{Name: "zero", In: 0, Want: 0},
		{Name: "small", In: 3, Want: 6},
		{Name: "large", In: 50, Want: 100},
	}
	RunTable(f, cases, func(n int) int { return n * 2 })

	if len(f.subs) != len(cases) {
		t.Fatalf("started %d subtests, want %d - each case is its own subtest", len(f.subs), len(cases))
	}
	for i, c := range cases {
		if want := "Table/" + c.Name; f.subs[i].name != want {
			t.Errorf("subtest %d is named %q, want %q", i, f.subs[i].name, want)
		}
	}
	if fails := f.failures(); len(fails) != 0 {
		t.Errorf("a passing table reported %v", fails)
	}
}

func TestRunTableReportsFailuresOnTheSubtest(t *testing.T) {
	f := &fake{name: "Table"}
	cases := []Case{
		{Name: "ok", In: 2, Want: 4},
		{Name: "broken", In: 5, Want: 999},
		{Name: "also-ok", In: 7, Want: 14},
	}
	RunTable(f, cases, func(n int) int { return n * 2 })

	if len(f.subs) != 3 {
		t.Fatalf("started %d subtests, want 3 - one failing row must not abandon the rest", len(f.subs))
	}
	if fails := f.failures(); len(fails) != 0 {
		t.Errorf("the parent reported %v, want the failure on the subtest that owns it", fails)
	}
	if fails := f.subs[0].failures(); len(fails) != 0 {
		t.Errorf("the passing row %q reported %v", cases[0].Name, fails)
	}
	if fails := f.subs[2].failures(); len(fails) != 0 {
		t.Errorf("the passing row %q reported %v", cases[2].Name, fails)
	}
	broken := f.subs[1].failures()
	if len(broken) != 1 {
		t.Fatalf("the failing row reported %d failures, want 1: %v", len(broken), broken)
	}
	for _, want := range []string{"10", "999"} {
		if !strings.Contains(broken[0], want) {
			t.Errorf("failure %q does not mention %s", broken[0], want)
		}
	}
}

func TestRunTableRejectsAnEmptyTable(t *testing.T) {
	f := &fake{name: "Table"}
	RunTable(f, nil, func(n int) int { return n })
	if len(f.subs) != 0 {
		t.Errorf("an empty table started %d subtests", len(f.subs))
	}
	if len(f.failures()) == 0 {
		t.Error("an empty table passed silently, which is the one thing a test may never do")
	}
}

func TestGoSuiteRunsARealSubtest(t *testing.T) {
	var (
		ran     bool
		gotName string
	)
	ok := GoSuite{t}.Run("child", func(s Suite) {
		ran = true
		gotName = s.Name()
	})
	if !ran {
		t.Fatal("GoSuite.Run never called the body")
	}
	if !ok {
		t.Error("GoSuite.Run returned false for a subtest that passed - it must return what t.Run returned")
	}
	if want := t.Name() + "/child"; gotName != want {
		t.Errorf("the body saw Name() = %q, want %q - Run must hand over the subtest's T, not the parent's", gotName, want)
	}
}

func TestGoSuiteDrivesRunTableForReal(t *testing.T) {
	RunTable(GoSuite{t}, []Case{
		{Name: "zero", In: 0, Want: 0},
		{Name: "one", In: 1, Want: 2},
		{Name: "big", In: 21, Want: 42},
	}, func(n int) int { return n * 2 })
}
