package main

import (
	"math"
	"testing"
)

// Named types over primitives: these only compile against the constraints if
// every union member carries a ~.
type ID int

type Celsius float64

func TestSumIntegers(t *testing.T) {
	if got := Sum([]int{1, 2, 3, 4}); got != 10 {
		t.Errorf("Sum([]int) = %d, want 10", got)
	}
	if got := Sum([]int8{1, 2, 3}); got != 6 {
		t.Errorf("Sum([]int8) = %d, want 6", got)
	}
	if got := Sum([]uint16{40, 2}); got != 42 {
		t.Errorf("Sum([]uint16) = %d, want 42", got)
	}
}

func TestSumFloats(t *testing.T) {
	if got := Sum([]float64{1.5, 2.25}); got != 3.75 {
		t.Errorf("Sum([]float64) = %v, want 3.75", got)
	}
	if got := Sum([]float32{0.5, 0.5}); got != 1 {
		t.Errorf("Sum([]float32) = %v, want 1", got)
	}
}

func TestSumOfEmptyIsTheZeroValue(t *testing.T) {
	if got := Sum([]int{}); got != 0 {
		t.Errorf("Sum of an empty slice = %d, want 0", got)
	}
	if got := Sum([]float64(nil)); got != 0 {
		t.Errorf("Sum of a nil slice = %v, want 0", got)
	}
}

func TestConstraintsAcceptNamedTypes(t *testing.T) {
	if got := Sum([]ID{1, 2, 3}); got != 6 {
		t.Errorf("Sum([]ID) = %d, want 6 - every union member needs a ~", got)
	}
	if got := Sum([]Celsius{20.5, 1.5}); got != 22 {
		t.Errorf("Sum([]Celsius) = %v, want 22 - every union member needs a ~", got)
	}
	if got, ok := Max([]ID{3, 9, 4}); !ok || got != 9 {
		t.Errorf("Max([]ID) = (%d, %t), want (9, true)", got, ok)
	}
}

func TestAverage(t *testing.T) {
	got, ok := Average([]int{1, 2, 3, 4})
	if !ok {
		t.Fatal("Average returned ok = false, want true")
	}
	if got != 2.5 {
		t.Errorf("Average([]int) = %v, want 2.5", got)
	}

	got, ok = Average([]float64{1, 2})
	if !ok || got != 1.5 {
		t.Errorf("Average([]float64) = (%v, %t), want (1.5, true)", got, ok)
	}
}

func TestAverageOfEmpty(t *testing.T) {
	got, ok := Average([]int{})
	if ok {
		t.Error("Average of an empty slice returned ok = true, want false")
	}
	if got != 0 {
		t.Errorf("Average of an empty slice = %v, want 0", got)
	}
}

func TestAverageConvertsPerElement(t *testing.T) {
	// Sixty 100s in an int8 sum would have overflowed many times over. Converting
	// each element to float64 first keeps the answer at 100.
	s := make([]int8, 60)
	for i := range s {
		s[i] = 100
	}
	got, ok := Average(s)
	if !ok {
		t.Fatal("Average returned ok = false, want true")
	}
	if math.Abs(got-100) > 1e-9 {
		t.Errorf("Average([]int8 of 100s) = %v, want 100 - convert each element, not the sum", got)
	}
}

func TestMinAndMax(t *testing.T) {
	s := []int{5, -3, 12, 0}
	if got, ok := Min(s); !ok || got != -3 {
		t.Errorf("Min = (%d, %t), want (-3, true)", got, ok)
	}
	if got, ok := Max(s); !ok || got != 12 {
		t.Errorf("Max = (%d, %t), want (12, true)", got, ok)
	}
}

func TestMinAndMaxSeedFromTheFirstElement(t *testing.T) {
	// Every element is negative, so anything seeded from the zero value wins by
	// accident and returns 0.
	s := []int{-5, -3, -12}
	if got, ok := Max(s); !ok || got != -3 {
		t.Errorf("Max(all negative) = (%d, %t), want (-3, true) - seed from s[0]", got, ok)
	}
	pos := []int{5, 3, 12}
	if got, ok := Min(pos); !ok || got != 3 {
		t.Errorf("Min(all positive) = (%d, %t), want (3, true) - seed from s[0]", got, ok)
	}
}

func TestMinAndMaxOfEmpty(t *testing.T) {
	if got, ok := Min([]float64{}); ok || got != 0 {
		t.Errorf("Min of an empty slice = (%v, %t), want (0, false)", got, ok)
	}
	if got, ok := Max([]string{}); ok || got != "" {
		t.Errorf("Max of an empty slice = (%q, %t), want (\"\", false)", got, ok)
	}
}

func TestOrderedIncludesString(t *testing.T) {
	got, ok := Max([]string{"go", "rust", "c"})
	if !ok || got != "rust" {
		t.Errorf("Max([]string) = (%q, %t), want (\"rust\", true) - Ordered includes ~string", got, ok)
	}
	if got, ok := Min([]string{"go", "rust", "c"}); !ok || got != "c" {
		t.Errorf("Min([]string) = (%q, %t), want (\"c\", true)", got, ok)
	}
}

func TestClamp(t *testing.T) {
	cases := []struct {
		v, lo, hi, want int
	}{
		{5, 0, 10, 5},
		{-1, 0, 10, 0},
		{99, 0, 10, 10},
		{0, 0, 10, 0},
		{10, 0, 10, 10},
	}
	for _, c := range cases {
		if got := Clamp(c.v, c.lo, c.hi); got != c.want {
			t.Errorf("Clamp(%d, %d, %d) = %d, want %d", c.v, c.lo, c.hi, got, c.want)
		}
	}
	if got := Clamp(0.5, 0.0, 1.0); got != 0.5 {
		t.Errorf("Clamp(0.5, 0, 1) = %v, want 0.5", got)
	}
}

func TestClampWithAnEmptyRange(t *testing.T) {
	if got := Clamp(5, 10, 0); got != 10 {
		t.Errorf("Clamp(5, 10, 0) = %d, want 10 - an empty range returns lo", got)
	}
}

func TestAbs(t *testing.T) {
	if got := Abs(-7); got != 7 {
		t.Errorf("Abs(-7) = %d, want 7", got)
	}
	if got := Abs(7); got != 7 {
		t.Errorf("Abs(7) = %d, want 7", got)
	}
	if got := Abs(-2.5); got != 2.5 {
		t.Errorf("Abs(-2.5) = %v, want 2.5", got)
	}
	if got := Abs(int64(0)); got != 0 {
		t.Errorf("Abs(0) = %d, want 0", got)
	}
}

func TestSumValues(t *testing.T) {
	if got := SumValues(map[string]int{"a": 1, "b": 2, "c": 3}); got != 6 {
		t.Errorf("SumValues = %d, want 6", got)
	}
	if got := SumValues(map[int]float64{1: 0.5, 2: 0.25}); got != 0.75 {
		t.Errorf("SumValues(floats) = %v, want 0.75", got)
	}
	if got := SumValues(map[string]int{}); got != 0 {
		t.Errorf("SumValues of an empty map = %d, want 0", got)
	}
	var nilMap map[string]int
	if got := SumValues(nilMap); got != 0 {
		t.Errorf("SumValues of a nil map = %d, want 0", got)
	}
}
