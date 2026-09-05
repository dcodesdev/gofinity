package main

import (
	"fmt"
	"strings"
	"testing"
)

// recorder is a TB that remembers what it was told instead of failing a test.
// It is the whole reason RunCases takes an interface.
type recorder struct {
	helpers  int
	messages []string
}

func (r *recorder) Helper() { r.helpers++ }

func (r *recorder) Errorf(format string, args ...any) {
	r.messages = append(r.messages, fmt.Sprintf(format, args...))
}

// refClamp is the reference the whole file is graded against.
func refClamp(n, lo, hi int) int {
	if n < lo {
		return lo
	}
	if n > hi {
		return hi
	}
	return n
}

func TestClamp(t *testing.T) {
	cases := []struct {
		name            string
		n, lo, hi, want int
	}{
		{"below", -4, 0, 10, 0},
		{"above", 40, 0, 10, 10},
		{"inside", 7, 0, 10, 7},
		{"at lo", 0, 0, 10, 0},
		{"at hi", 10, 0, 10, 10},
		{"negative range", -1, -8, -3, -3},
		{"below negative range", -40, -8, -3, -8},
		{"single value range", 9, 4, 4, 4},
		{"zero width at zero", -2, 0, 0, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Clamp(c.n, c.lo, c.hi); got != c.want {
				t.Errorf("Clamp(%d, %d, %d) = %d, want %d", c.n, c.lo, c.hi, got, c.want)
			}
		})
	}
}

func TestClampCasesAreCorrect(t *testing.T) {
	cases := ClampCases()
	if len(cases) < 7 {
		t.Fatalf("ClampCases() returned %d cases, want at least 7 - one situation is uncovered", len(cases))
	}
	seen := map[string]bool{}
	for i, c := range cases {
		if strings.TrimSpace(c.Name) == "" {
			t.Errorf("case %d has an empty name", i)
		}
		if seen[c.Name] {
			t.Errorf("case %d repeats the name %q", i, c.Name)
		}
		seen[c.Name] = true
		if c.Lo > c.Hi {
			t.Errorf("case %q has lo %d above hi %d, which Clamp does not promise anything about", c.Name, c.Lo, c.Hi)
			continue
		}
		if want := refClamp(c.N, c.Lo, c.Hi); c.Want != want {
			t.Errorf("case %q expects Clamp(%d, %d, %d) = %d, but the answer is %d", c.Name, c.N, c.Lo, c.Hi, c.Want, want)
		}
	}
}

func TestClampCasesCoverTheSituations(t *testing.T) {
	cases := ClampCases()
	covered := map[string]bool{}
	for _, c := range cases {
		if c.Lo > c.Hi {
			continue
		}
		switch {
		case c.N < c.Lo:
			covered["below lo"] = true
		case c.N > c.Hi:
			covered["above hi"] = true
		case c.N == c.Lo && c.N == c.Hi:
			covered["at lo"] = true
			covered["at hi"] = true
		case c.N == c.Lo:
			covered["at lo"] = true
		case c.N == c.Hi:
			covered["at hi"] = true
		default:
			covered["strictly inside"] = true
		}
		if c.Hi < 0 {
			covered["negative range"] = true
		}
		if c.Lo == c.Hi {
			covered["lo == hi"] = true
		}
	}
	for _, want := range []string{
		"below lo", "above hi", "strictly inside", "at lo", "at hi", "negative range", "lo == hi",
	} {
		if !covered[want] {
			t.Errorf("no case covers %q", want)
		}
	}
}

func TestRunCasesReportsNothingWhenEveryCasePasses(t *testing.T) {
	r := &recorder{}
	cases := []Case{
		{Name: "below", N: -4, Lo: 0, Hi: 10, Want: 0},
		{Name: "inside", N: 3, Lo: 0, Hi: 10, Want: 3},
		{Name: "above", N: 99, Lo: 0, Hi: 10, Want: 10},
	}
	RunCases(r, cases, func(c Case) int { return refClamp(c.N, c.Lo, c.Hi) })
	if len(r.messages) != 0 {
		t.Errorf("passing cases reported %d failures: %v", len(r.messages), r.messages)
	}
}

func TestRunCasesReportsEveryFailingCase(t *testing.T) {
	r := &recorder{}
	cases := []Case{
		{Name: "alpha", N: 0, Lo: 0, Hi: 10, Want: 1001},
		{Name: "beta", N: 0, Lo: 0, Hi: 10, Want: 0},
		{Name: "gamma", N: 0, Lo: 0, Hi: 10, Want: 3003},
		{Name: "delta", N: 0, Lo: 0, Hi: 10, Want: 4004},
	}
	calls := []string{}
	RunCases(r, cases, func(c Case) int {
		calls = append(calls, c.Name)
		if c.Want == 0 {
			return 0
		}
		return c.Want + 7000
	})

	if len(calls) != len(cases) {
		t.Fatalf("fn was called %d times (%v), want once per case in order - a table that stops at the first failure hides the rest", len(calls), calls)
	}
	for i, c := range cases {
		if calls[i] != c.Name {
			t.Fatalf("fn call %d was for %q, want %q", i, calls[i], c.Name)
		}
	}
	if len(r.messages) != 3 {
		t.Fatalf("got %d failures, want 3 (alpha, gamma, delta): %v", len(r.messages), r.messages)
	}
	joined := strings.Join(r.messages, "\n")
	for _, name := range []string{"alpha", "gamma", "delta"} {
		if !strings.Contains(joined, name) {
			t.Errorf("no failure message names the case %q:\n%s", name, joined)
		}
	}
	if strings.Contains(joined, "beta") {
		t.Errorf("the passing case beta was reported as a failure:\n%s", joined)
	}
	for _, value := range []string{"1001", "8001", "3003", "10003", "4004", "11004"} {
		if !strings.Contains(joined, value) {
			t.Errorf("the failures do not mention %s - a message needs both the got and the want:\n%s", value, joined)
		}
	}
}

func TestRunCasesMarksItselfAsAHelper(t *testing.T) {
	r := &recorder{}
	RunCases(r, []Case{{Name: "wrong", N: 5, Lo: 0, Hi: 10, Want: 9}}, func(c Case) int { return 5 })
	if r.helpers == 0 {
		t.Error("RunCases never called t.Helper(), so failures point at the runner instead of the test that called it")
	}
}

func TestRunCasesHandlesAnEmptyTable(t *testing.T) {
	r := &recorder{}
	RunCases(r, nil, func(c Case) int { return 0 })
	if len(r.messages) != 0 {
		t.Errorf("an empty table reported %v", r.messages)
	}
}
