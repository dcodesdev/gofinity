package main

import (
	"errors"
	"testing"
)

func TestScoreSuccess(t *testing.T) {
	got, err := Score("alice")
	if err != nil {
		t.Fatalf(`Score("alice") returned %v, want no error`, err)
	}
	if got != 42 {
		t.Errorf(`Score("alice") = %d, want 42`, got)
	}
}

func TestScoreWrapsNotFound(t *testing.T) {
	got, err := Score("nobody")
	if err == nil {
		t.Fatal(`Score("nobody") returned no error, want one`)
	}
	if want := `score "nobody": not found`; err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
	if !errors.Is(err, ErrNotFound) {
		t.Error("errors.Is(err, ErrNotFound) = false - wrap with the w verb, not v")
	}
	if errors.Is(err, ErrInvalid) {
		t.Error("errors.Is(err, ErrInvalid) = true, want false")
	}
	if got != 0 {
		t.Errorf(`Score("nobody") = %d beside an error, want 0`, got)
	}
}

func TestScoreWrapsInvalid(t *testing.T) {
	for _, tc := range []struct {
		name string
		want string
	}{
		{"bob", `score "bob": parse "not-a-number": invalid`},
		{"carol", `score "carol": parse "-1": invalid`},
	} {
		_, err := Score(tc.name)
		if err == nil {
			t.Fatalf("Score(%q) returned no error, want one", tc.name)
		}
		if err.Error() != tc.want {
			t.Errorf("Score(%q) error = %q, want %q", tc.name, err.Error(), tc.want)
		}
		if !errors.Is(err, ErrInvalid) {
			t.Errorf("Score(%q): errors.Is(err, ErrInvalid) = false, want true", tc.name)
		}
		if errors.Is(err, ErrNotFound) {
			t.Errorf("Score(%q): errors.Is(err, ErrNotFound) = true, want false", tc.name)
		}
	}
}

func TestParseScore(t *testing.T) {
	if got, err := parseScore("0"); err != nil || got != 0 {
		t.Errorf(`parseScore("0") = (%d, %v), want (0, nil)`, got, err)
	}
	if got, err := parseScore("1007"); err != nil || got != 1007 {
		t.Errorf(`parseScore("1007") = (%d, %v), want (1007, nil)`, got, err)
	}
	for _, raw := range []string{"", "12x", "-1", " 3"} {
		_, err := parseScore(raw)
		if !errors.Is(err, ErrInvalid) {
			t.Errorf("parseScore(%q) error = %v, want ErrInvalid in the chain", raw, err)
		}
	}
}

func TestReason(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{nil, "ok"},
		{ErrNotFound, "missing"},
		{ErrInvalid, "bad-data"},
		{errors.New("something else"), "unknown"},
	}
	for _, tc := range cases {
		if got := Reason(tc.err); got != tc.want {
			t.Errorf("Reason(%v) = %q, want %q", tc.err, got, tc.want)
		}
	}

	if _, err := Score("nobody"); Reason(err) != "missing" {
		t.Error("Reason of a wrapped ErrNotFound is not \"missing\" - use errors.Is, not ==")
	}
	if _, err := Score("bob"); Reason(err) != "bad-data" {
		t.Error("Reason of a doubly wrapped ErrInvalid is not \"bad-data\"")
	}
}

func TestUnwrapped(t *testing.T) {
	if Unwrapped(nil) != nil {
		t.Error("Unwrapped(nil) is not nil")
	}
	if got := Unwrapped(ErrInvalid); got != ErrInvalid {
		t.Errorf("Unwrapped(ErrInvalid) = %v, want ErrInvalid itself", got)
	}
	_, err := Score("bob")
	if got := Unwrapped(err); got != ErrInvalid {
		t.Errorf("Unwrapped of a doubly wrapped error = %v, want ErrInvalid", got)
	}
	_, err = Score("nobody")
	if got := Unwrapped(err); got != ErrNotFound {
		t.Errorf("Unwrapped = %v, want ErrNotFound", got)
	}
}

func TestDepth(t *testing.T) {
	if got := Depth(nil); got != 0 {
		t.Errorf("Depth(nil) = %d, want 0", got)
	}
	if got := Depth(ErrInvalid); got != 0 {
		t.Errorf("Depth(ErrInvalid) = %d, want 0", got)
	}
	if _, err := Score("nobody"); Depth(err) != 1 {
		t.Errorf(`Depth(Score("nobody")) = %d, want 1`, Depth(err))
	}
	if _, err := Score("bob"); Depth(err) != 2 {
		t.Errorf(`Depth(Score("bob")) = %d, want 2`, Depth(err))
	}
}

func TestTotal(t *testing.T) {
	got, err := Total([]string{"alice", "alice"})
	if err != nil {
		t.Fatalf("Total returned %v, want no error", err)
	}
	if got != 84 {
		t.Errorf("Total = %d, want 84", got)
	}

	if got, err := Total(nil); err != nil || got != 0 {
		t.Errorf("Total(nil) = (%d, %v), want (0, nil)", got, err)
	}

	got, err = Total([]string{"alice", "nobody", "bob"})
	if err == nil {
		t.Fatal("Total returned no error, want one")
	}
	if want := `entry 1: score "nobody": not found`; err.Error() != want {
		t.Errorf("Total error = %q, want %q", err.Error(), want)
	}
	if !errors.Is(err, ErrNotFound) {
		t.Error("Total broke the chain: errors.Is(err, ErrNotFound) = false")
	}
	if Depth(err) != 2 {
		t.Errorf("Depth of the Total error = %d, want 2", Depth(err))
	}
	if got != 0 {
		t.Errorf("Total = %d beside an error, want 0", got)
	}

	_, err = Total([]string{"carol"})
	if want := `entry 0: score "carol": parse "-1": invalid`; err == nil || err.Error() != want {
		t.Errorf("Total error = %v, want %q", err, want)
	}
}
