package main

import (
	"fmt"
	"testing"
)

// Celsius never mentions fmt.Stringer, and satisfies it anyway.
var _ fmt.Stringer = Celsius(0)

func TestCelsiusString(t *testing.T) {
	tests := []struct {
		in   Celsius
		want string
	}{
		{21.5, "21.5C"},
		{0, "0.0C"},
		{-4, "-4.0C"},
	}
	for _, tt := range tests {
		if got := tt.in.String(); got != tt.want {
			t.Errorf("Celsius(%v).String() = %q, want %q", float64(tt.in), got, tt.want)
		}
	}
}

func TestDescribe(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"nil", nil, "nil"},
		{"int", 42, "int 42"},
		{"negative int", -7, "int -7"},
		{"float64", 3.5, "float64 3.50"},
		{"string", "hi", `string "hi"`},
		{"empty string", "", `string ""`},
		{"bool", true, "bool true"},
		{"false", false, "bool false"},
		{"slice", []int{1, 2, 3}, "[]int of 3"},
		{"empty slice", []int{}, "[]int of 0"},
		{"nil slice", []int(nil), "[]int of 0"},
		{"stringer", Celsius(21.5), "stringer 21.5C"},
		{"struct", Point{X: 1, Y: 2}, "other main.Point"},
		{"other slice", []string{"a"}, "other []string"},
	}
	for _, tt := range tests {
		if got := Describe(tt.in); got != tt.want {
			t.Errorf("%s: Describe(%v) = %q, want %q", tt.name, tt.in, got, tt.want)
		}
	}
}

func TestDescribeCaseOrder(t *testing.T) {
	// Celsius is a float64 underneath, and a type switch matches the named
	// type, not the underlying one. A "case float64" never sees it.
	if got := Describe(Celsius(1)); got != "stringer 1.0C" {
		t.Errorf("Describe(Celsius(1)) = %q, want %q", got, "stringer 1.0C")
	}
	// A plain float64 has no String method, so it must land on the float case.
	if got := Describe(1.0); got != "float64 1.00" {
		t.Errorf("Describe(1.0) = %q, want %q", got, "float64 1.00")
	}
}

func TestSumNumbers(t *testing.T) {
	vals := []any{1, int64(2), 3.5, "4", true, nil, Point{}}
	if got := SumNumbers(vals); got != 6.5 {
		t.Errorf("SumNumbers = %v, want 6.5 - the string and the bool do not count", got)
	}
	if got := SumNumbers(nil); got != 0 {
		t.Errorf("SumNumbers(nil) = %v, want 0", got)
	}
	if got := SumNumbers([]any{"a", "b"}); got != 0 {
		t.Errorf("SumNumbers with no numbers = %v, want 0", got)
	}
}

func TestAsInt(t *testing.T) {
	if n, ok := AsInt(7); !ok || n != 7 {
		t.Errorf("AsInt(7) = %d, %t, want 7, true", n, ok)
	}
	for _, v := range []any{int64(7), 7.0, "7", nil, true} {
		if n, ok := AsInt(v); ok {
			t.Errorf("AsInt(%#v) = %d, true - only a plain int matches", v, n)
		}
	}
	if n, _ := AsInt("nope"); n != 0 {
		t.Errorf("AsInt on a miss returned %d, want the zero int", n)
	}
}

func TestJoinStrings(t *testing.T) {
	vals := []any{"a", 1, "b", Celsius(20), nil, "c"}
	want := "a, b, 20.0C, c"
	if got := JoinStrings(vals, ", "); got != want {
		t.Errorf("JoinStrings = %q, want %q", got, want)
	}
	if got := JoinStrings(nil, ", "); got != "" {
		t.Errorf("JoinStrings(nil) = %q, want an empty string", got)
	}
	if got := JoinStrings([]any{1, 2}, "-"); got != "" {
		t.Errorf("JoinStrings with nothing to join = %q, want an empty string", got)
	}
	if got := JoinStrings([]any{"only"}, "-"); got != "only" {
		t.Errorf("JoinStrings of one = %q, want %q - no trailing separator", got, "only")
	}
}
