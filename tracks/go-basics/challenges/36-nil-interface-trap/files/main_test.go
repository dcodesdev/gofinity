package main

import (
	"errors"
	"testing"
)

// *Failure satisfies error; Failure does not, because Error has a pointer
// receiver. That asymmetry is what makes the trap possible in the first place.
var _ error = (*Failure)(nil)

func TestErrorSurvivesANilReceiver(t *testing.T) {
	var f *Failure
	if got := f.Error(); got != "<no failure>" {
		t.Errorf("(*Failure)(nil).Error() = %q, want %q", got, "<no failure>")
	}
	if got := (&Failure{Msg: "boom"}).Error(); got != "boom" {
		t.Errorf("Error() = %q, want %q", got, "boom")
	}
}

func TestBrokenIsBroken(t *testing.T) {
	// This is the behaviour of the code you were given, not a bug to fix.
	// The interface has a type in it, so it is not nil.
	err := Broken(false)
	if err == nil {
		t.Fatal("Broken(false) == nil - Broken is meant to demonstrate the trap")
	}
	if got := err.Error(); got != "<no failure>" {
		t.Errorf("Broken(false).Error() = %q, want %q", got, "<no failure>")
	}
}

func TestFixed(t *testing.T) {
	if err := Fixed(false); err != nil {
		t.Errorf("Fixed(false) = %v, want a nil error - return a bare nil, not a nil *Failure", err)
	}
	err := Fixed(true)
	if err == nil {
		t.Fatal("Fixed(true) = nil, want an error")
	}
	if got := err.Error(); got != "broken" {
		t.Errorf("Fixed(true).Error() = %q, want %q", got, "broken")
	}
	var f *Failure
	if !errors.As(err, &f) || f == nil {
		t.Errorf("Fixed(true) = %#v, want a non-nil *Failure", err)
	}
}

func TestWrap(t *testing.T) {
	if err := Wrap(nil); err != nil {
		t.Errorf("Wrap(nil) = %v, want a nil error", err)
	}
	var f *Failure
	if err := Wrap(f); err != nil {
		t.Errorf("Wrap of a nil *Failure variable = %v, want a nil error", err)
	}
	real := &Failure{Msg: "boom"}
	if err := Wrap(real); err != error(real) {
		t.Errorf("Wrap(%v) = %v, want the same *Failure back", real, err)
	}
}

func TestHoldsNilPointer(t *testing.T) {
	if HoldsNilPointer(nil) {
		t.Error("HoldsNilPointer(nil) = true, want false - a nil error is not the trap")
	}
	if !HoldsNilPointer(Broken(false)) {
		t.Error("HoldsNilPointer(Broken(false)) = false, want true")
	}
	if HoldsNilPointer(&Failure{Msg: "boom"}) {
		t.Error("HoldsNilPointer on a real failure = true, want false")
	}
	if HoldsNilPointer(errors.New("other")) {
		t.Error("HoldsNilPointer on some other error type = true, want false")
	}
}

func TestCompact(t *testing.T) {
	boom := &Failure{Msg: "boom"}
	other := errors.New("other")
	errs := []error{nil, Broken(false), boom, nil, other}

	got := Compact(errs)
	if len(got) != 2 {
		t.Fatalf("Compact kept %d errors (%v), want 2", len(got), got)
	}
	if got[0] != error(boom) || got[1] != other {
		t.Errorf("Compact = %v, want [%v %v] in that order", got, boom, other)
	}

	if len(Compact(nil)) != 0 {
		t.Error("Compact(nil) kept something, want an empty result")
	}
	if len(Compact([]error{nil, Broken(false)})) != 0 {
		t.Error("Compact of nothing real kept something, want an empty result")
	}

	// The input must be left alone, and the result must not alias it.
	all := []error{boom, other}
	out := Compact(all)
	if len(out) == 2 && &out[0] == &all[0] {
		t.Error("Compact returned the input slice itself, want a new one")
	}
	if len(errs) != 5 || errs[0] != nil {
		t.Errorf("Compact modified its input: %v", errs)
	}
}

func TestFirstError(t *testing.T) {
	boom := &Failure{Msg: "boom"}
	other := errors.New("other")

	if err := FirstError(nil); err != nil {
		t.Errorf("FirstError(nil) = %v, want nil", err)
	}
	if err := FirstError([]error{nil, Broken(false), nil}); err != nil {
		t.Errorf("FirstError with nothing real = %v (%T), want a bare nil", err, err)
	}
	if err := FirstError([]error{nil, Broken(false), boom, other}); err != error(boom) {
		t.Errorf("FirstError = %v, want %v", err, boom)
	}
	if err := FirstError([]error{other, boom}); err != other {
		t.Errorf("FirstError = %v, want the first one, %v", err, other)
	}
}
