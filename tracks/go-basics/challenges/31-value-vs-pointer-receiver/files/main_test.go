package main

import (
	"fmt"
	"testing"
)

// Add is declared with a pointer receiver, so it belongs to the method set of
// *Counter and not to that of Counter. This line compiles; the same line with
// Counter{} on the right does not, and that is the whole rule.
type Adder interface{ Add(int) }

var _ Adder = (*Counter)(nil)

func TestValueAndAdd(t *testing.T) {
	// The zero Counter is usable straight away, and c is addressable, so
	// c.Add(3) is shorthand for (&c).Add(3).
	var c Counter
	c.Add(3)
	c.Add(4)
	if got := c.Value(); got != 7 {
		t.Errorf("after Add(3) and Add(4), Value() = %d, want 7", got)
	}
	if c.N != 7 {
		t.Errorf("N = %d, want 7 - Add must write through the pointer receiver", c.N)
	}

	p := &Counter{N: 1}
	p.Add(-5)
	if p.N != -4 {
		t.Errorf("Add(-5) on &Counter{N: 1} left N = %d, want -4", p.N)
	}
}

func TestPlusLeavesTheReceiverAlone(t *testing.T) {
	c := Counter{N: 5}
	got := c.Plus(3)
	if got.N != 8 {
		t.Errorf("Plus(3) = %+v, want N 8", got)
	}
	if c.N != 5 {
		t.Errorf("Plus changed the receiver to %d - a value receiver is a copy", c.N)
	}
	if got := c.Plus(0); got != c {
		t.Errorf("Plus(0) = %+v, want an equal Counter %+v", got, c)
	}
}

func TestReset(t *testing.T) {
	c := Counter{N: 42}
	c.Reset()
	if c.N != 0 {
		t.Errorf("after Reset, N = %d, want 0", c.N)
	}
	c.Reset()
	if c.N != 0 {
		t.Errorf("Reset is not idempotent: N = %d, want 0", c.N)
	}
}

func TestSumValues(t *testing.T) {
	cs := []Counter{{N: 1}, {N: 2}, {N: 39}}
	if got := SumValues(cs); got != 42 {
		t.Errorf("SumValues = %d, want 42", got)
	}
	if got := SumValues(nil); got != 0 {
		t.Errorf("SumValues(nil) = %d, want 0", got)
	}
}

func TestAddEach(t *testing.T) {
	cs := []Counter{{N: 1}, {N: 2}, {N: 3}}
	AddEach(cs, 10)
	for i, want := range []int{11, 12, 13} {
		if cs[i].N != want {
			t.Errorf("cs[%d].N = %d, want %d - calling Add on the range copy changes nothing", i, cs[i].N, want)
		}
	}

	AddEach(nil, 1) // must not panic
}

func TestTemperature(t *testing.T) {
	if got := Temperature(21.5).Warmer(0.5); got != Temperature(22) {
		t.Errorf("Warmer(0.5) = %v, want 22C", float64(got))
	}
	if got := Temperature(0).Warmer(-3.5); got != Temperature(-3.5) {
		t.Errorf("Warmer(-3.5) from zero = %v, want -3.5C", float64(got))
	}

	original := Temperature(10)
	original.Warmer(5)
	if original != Temperature(10) {
		t.Errorf("Warmer changed its receiver to %v - it returns a new value", float64(original))
	}
}

func TestTemperatureString(t *testing.T) {
	tests := []struct {
		in   Temperature
		want string
	}{
		{21.5, "21.5C"},
		{0, "0.0C"},
		{-3.5, "-3.5C"},
		{100, "100.0C"},
	}
	for _, tt := range tests {
		if got := tt.in.String(); got != tt.want {
			t.Errorf("Temperature(%v).String() = %q, want %q", float64(tt.in), got, tt.want)
		}
	}

	// A String method is what fmt reaches for, so printing the value formats
	// it without anyone asking.
	if got := fmt.Sprint(Temperature(21.5)); got != "21.5C" {
		t.Errorf("fmt.Sprint(Temperature(21.5)) = %q, want %q", got, "21.5C")
	}
}
