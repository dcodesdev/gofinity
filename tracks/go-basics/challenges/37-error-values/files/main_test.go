package main

import (
	"errors"
	"testing"
)

func TestSentinelsAreStableValues(t *testing.T) {
	if ErrEmpty == nil || ErrNotFound == nil {
		t.Fatal("both sentinels must be non-nil package-level values")
	}
	if ErrEmpty == ErrNotFound {
		t.Error("ErrEmpty and ErrNotFound are the same value, want two distinct sentinels")
	}
	if got := ErrNotFound.Error(); got != "not found" {
		t.Errorf("ErrNotFound.Error() = %q, want %q", got, "not found")
	}
}

func TestDivide(t *testing.T) {
	got, err := Divide(84, 7)
	if err != nil {
		t.Fatalf("Divide(84, 7) returned %v, want no error", err)
	}
	if got != 12 {
		t.Errorf("Divide(84, 7) = %d, want 12", got)
	}

	got, err = Divide(7, 0)
	if err == nil {
		t.Fatal("Divide(7, 0) returned no error, want one")
	}
	if want := "divide 7 by zero"; err.Error() != want {
		t.Errorf("Divide(7, 0) error = %q, want %q", err.Error(), want)
	}
	if got != 0 {
		t.Errorf("Divide(7, 0) = %d beside an error, want the zero value 0", got)
	}
}

func TestFirst(t *testing.T) {
	got, err := First([]int{4, 5, 6})
	if err != nil || got != 4 {
		t.Errorf("First([4 5 6]) = (%d, %v), want (4, nil)", got, err)
	}

	got, err = First(nil)
	if !errors.Is(err, ErrEmpty) {
		t.Errorf("First(nil) error = %v, want ErrEmpty itself", err)
	}
	if got != 0 {
		t.Errorf("First(nil) = %d, want 0", got)
	}
}

func TestLookup(t *testing.T) {
	m := map[string]int{"a": 1}

	got, err := Lookup(m, "a")
	if err != nil || got != 1 {
		t.Errorf(`Lookup(m, "a") = (%d, %v), want (1, nil)`, got, err)
	}

	got, err = Lookup(m, "b")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf(`Lookup(m, "b") error = %v, want ErrNotFound`, err)
	}
	if got != 0 {
		t.Errorf(`Lookup(m, "b") = %d, want 0`, got)
	}

	if _, err := Lookup(nil, "a"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Lookup(nil, ...) error = %v, want ErrNotFound", err)
	}

	// A stored zero is a hit, not a miss. Use the comma-ok form, not `m[key] == 0`.
	if got, err := Lookup(map[string]int{"z": 0}, "z"); err != nil || got != 0 {
		t.Errorf(`Lookup({"z": 0}, "z") = (%d, %v), want (0, nil)`, got, err)
	}
}

func TestSumQuotients(t *testing.T) {
	got, err := SumQuotients(12, []int{1, 2, 3})
	if err != nil {
		t.Fatalf("SumQuotients(12, [1 2 3]) returned %v, want no error", err)
	}
	if got != 22 { // 12 + 6 + 4
		t.Errorf("SumQuotients(12, [1 2 3]) = %d, want 22", got)
	}

	if got, err := SumQuotients(12, nil); err != nil || got != 0 {
		t.Errorf("SumQuotients(12, nil) = (%d, %v), want (0, nil)", got, err)
	}

	got, err = SumQuotients(7, []int{1, 0, 2})
	if err == nil {
		t.Fatal("SumQuotients with a zero divisor returned no error, want one")
	}
	if want := "divide 7 by zero"; err.Error() != want {
		t.Errorf("SumQuotients error = %q, want %q - return the error Divide gave you", err.Error(), want)
	}
	if got != 0 {
		t.Errorf("SumQuotients = %d beside an error, want 0", got)
	}
}

func TestDescribe(t *testing.T) {
	if got := Describe(Divide(84, 7)); got != "ok: 12" {
		t.Errorf("Describe(Divide(84, 7)) = %q, want %q", got, "ok: 12")
	}
	if got := Describe(Divide(7, 0)); got != "failed: divide 7 by zero" {
		t.Errorf("Describe(Divide(7, 0)) = %q, want %q", got, "failed: divide 7 by zero")
	}
	if got := Describe(0, ErrEmpty); got != "failed: empty input" {
		t.Errorf("Describe(0, ErrEmpty) = %q, want %q", got, "failed: empty input")
	}
}
