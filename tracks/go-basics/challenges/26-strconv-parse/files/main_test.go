package main

import (
	"errors"
	"maps"
	"strconv"
	"testing"
)

func TestParseIntOr(t *testing.T) {
	tests := []struct {
		in       string
		fallback int
		want     int
	}{
		{"42", -1, 42},
		{"  42\t", -1, 42},
		{"-7", 0, -7},
		{"+7", 0, 7},
		{"", -1, -1},
		{"forty", -1, -1},
		{"4.0", -1, -1},
		{"9999999999999999999999", -1, -1},
	}
	for _, tt := range tests {
		if got := ParseIntOr(tt.in, tt.fallback); got != tt.want {
			t.Errorf("ParseIntOr(%q, %d) = %d, want %d", tt.in, tt.fallback, got, tt.want)
		}
	}
}

func TestParseFloatOr(t *testing.T) {
	tests := []struct {
		in       string
		fallback float64
		want     float64
	}{
		{"1.5", -1, 1.5},
		{" 1.5 ", -1, 1.5},
		{"2", -1, 2},
		{"-0.25", -1, -0.25},
		{"1e3", -1, 1000},
		{"", -1, -1},
		{"one point five", -1, -1},
	}
	for _, tt := range tests {
		if got := ParseFloatOr(tt.in, tt.fallback); got != tt.want {
			t.Errorf("ParseFloatOr(%q, %v) = %v, want %v", tt.in, tt.fallback, got, tt.want)
		}
	}
}

func TestParseBoolOr(t *testing.T) {
	tests := []struct {
		in             string
		fallback, want bool
	}{
		{"true", false, true},
		{"TRUE", false, true},
		{" t ", false, true},
		{"1", false, true},
		{"false", true, false},
		{"0", true, false},
		{"", true, true},
		{"yes", true, true},
		{"yes", false, false},
	}
	for _, tt := range tests {
		if got := ParseBoolOr(tt.in, tt.fallback); got != tt.want {
			t.Errorf("ParseBoolOr(%q, %v) = %v, want %v", tt.in, tt.fallback, got, tt.want)
		}
	}
}

func TestSumFields(t *testing.T) {
	ok := []struct {
		in   string
		want int
	}{
		{"1,2,3", 6},
		{" 1 , 2 ,3 ", 6},
		{"", 0},
		{"  ", 0},
		{"1,,2", 3},
		{"-4,4", 0},
		{"10", 10},
	}
	for _, tt := range ok {
		got, err := SumFields(tt.in)
		if err != nil {
			t.Errorf("SumFields(%q) returned %v, want no error", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("SumFields(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestSumFieldsErrors(t *testing.T) {
	got, err := SumFields("1,x,3")
	if err == nil {
		t.Fatal("SumFields(\"1,x,3\") returned no error, want one")
	}
	if got != 0 {
		t.Errorf("SumFields(\"1,x,3\") = %d, want 0 alongside the error", got)
	}
	if !errors.Is(err, strconv.ErrSyntax) {
		t.Errorf("SumFields error = %v, want it to wrap strconv.ErrSyntax", err)
	}

	_, err = SumFields("9999999999999999999999")
	if !errors.Is(err, strconv.ErrRange) {
		t.Errorf("SumFields(huge) error = %v, want it to wrap strconv.ErrRange", err)
	}
}

func TestParseKeyValues(t *testing.T) {
	tests := []struct {
		in   string
		want map[string]int
	}{
		{"a=1,b=2", map[string]int{"a": 1, "b": 2}},
		{" a = 1 , b = 2 ", map[string]int{"a": 1, "b": 2}},
		{"", map[string]int{}},
		{"a=1,,b=2", map[string]int{"a": 1, "b": 2}},
		{"a=1,a=2", map[string]int{"a": 2}},
		{"a=-1", map[string]int{"a": -1}},
	}
	for _, tt := range tests {
		got, err := ParseKeyValues(tt.in)
		if err != nil {
			t.Errorf("ParseKeyValues(%q) returned %v, want no error", tt.in, err)
			continue
		}
		if got == nil {
			t.Errorf("ParseKeyValues(%q) = nil, want a map", tt.in)
			continue
		}
		if !maps.Equal(got, tt.want) {
			t.Errorf("ParseKeyValues(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}

	for _, bad := range []string{"a", "a=1,b", "a=x", "=1,"} {
		if _, err := ParseKeyValues(bad); err == nil {
			t.Errorf("ParseKeyValues(%q) returned no error, want one", bad)
		}
	}
}

func TestJoinInts(t *testing.T) {
	tests := []struct {
		in   []int
		want string
	}{
		{[]int{1, 2, 3}, "1,2,3"},
		{[]int{-1}, "-1"},
		{[]int{}, ""},
		{nil, ""},
	}
	for _, tt := range tests {
		if got := JoinInts(tt.in); got != tt.want {
			t.Errorf("JoinInts(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	nums := []int{5, -3, 0, 12}
	sum, err := SumFields(JoinInts(nums))
	if err != nil {
		t.Fatalf("SumFields(JoinInts(...)) returned %v", err)
	}
	if sum != 14 {
		t.Errorf("round trip sum = %d, want 14", sum)
	}
}
