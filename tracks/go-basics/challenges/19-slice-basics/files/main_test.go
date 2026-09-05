package main

import (
	"slices"
	"testing"
)

func TestDescribe(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want string
	}{
		{"nil", nil, "len=0 cap=0"},
		{"literal", []int{1, 2, 3}, "len=3 cap=3"},
		{"made with spare capacity", make([]int, 3, 8), "len=3 cap=8"},
		{"empty with capacity", make([]int, 0, 4), "len=0 cap=4"},
	}
	for _, tt := range tests {
		if got := Describe(tt.in); got != tt.want {
			t.Errorf("Describe(%s) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestDescribeAfterReslicing(t *testing.T) {
	// Reslicing keeps the same backing array, so the capacity is measured from
	// the low bound to the end of that array, not from the new length.
	s := make([]int, 5, 10)
	if got, want := Describe(s[1:3]), "len=2 cap=9"; got != want {
		t.Errorf("Describe(s[1:3]) = %q, want %q", got, want)
	}
}

func TestFirst(t *testing.T) {
	if got, ok := First([]int{7, 8, 9}); got != 7 || !ok {
		t.Errorf("First([7 8 9]) = %d, %v, want 7, true", got, ok)
	}
	if got, ok := First([]int{}); got != 0 || ok {
		t.Errorf("First(empty) = %d, %v, want 0, false", got, ok)
	}
	if got, ok := First(nil); got != 0 || ok {
		t.Errorf("First(nil) = %d, %v, want 0, false", got, ok)
	}
}

func TestLast(t *testing.T) {
	if got, ok := Last([]int{7, 8, 9}); got != 9 || !ok {
		t.Errorf("Last([7 8 9]) = %d, %v, want 9, true", got, ok)
	}
	if got, ok := Last([]int{4}); got != 4 || !ok {
		t.Errorf("Last([4]) = %d, %v, want 4, true", got, ok)
	}
	if got, ok := Last(nil); got != 0 || ok {
		t.Errorf("Last(nil) = %d, %v, want 0, false", got, ok)
	}
}

func TestHead(t *testing.T) {
	s := []int{1, 2, 3, 4, 5}
	tests := []struct {
		n    int
		want []int
	}{
		{2, []int{1, 2}},
		{0, []int{}},
		{5, []int{1, 2, 3, 4, 5}},
		{9, []int{1, 2, 3, 4, 5}},
		{-1, []int{}},
	}
	for _, tt := range tests {
		if got := Head(s, tt.n); !slices.Equal(got, tt.want) {
			t.Errorf("Head(%v, %d) = %v, want %v", s, tt.n, got, tt.want)
		}
	}
	if got := Head(nil, 3); len(got) != 0 {
		t.Errorf("Head(nil, 3) = %v, want an empty result", got)
	}
}

func TestTail(t *testing.T) {
	s := []int{1, 2, 3, 4, 5}
	tests := []struct {
		n    int
		want []int
	}{
		{2, []int{4, 5}},
		{0, []int{}},
		{5, []int{1, 2, 3, 4, 5}},
		{9, []int{1, 2, 3, 4, 5}},
		{-1, []int{}},
	}
	for _, tt := range tests {
		if got := Tail(s, tt.n); !slices.Equal(got, tt.want) {
			t.Errorf("Tail(%v, %d) = %v, want %v", s, tt.n, got, tt.want)
		}
	}
	if got := Tail(nil, 3); len(got) != 0 {
		t.Errorf("Tail(nil, 3) = %v, want an empty result", got)
	}
}

func TestSumArray(t *testing.T) {
	if got := SumArray([5]int{1, 2, 3, 4, 5}); got != 15 {
		t.Errorf("SumArray([1 2 3 4 5]) = %d, want 15", got)
	}
	if got := SumArray([5]int{}); got != 0 {
		t.Errorf("SumArray(zero array) = %d, want 0", got)
	}
	if got := SumArray([5]int{-2, 2, 0, 10, -10}); got != 0 {
		t.Errorf("SumArray([-2 2 0 10 -10]) = %d, want 0", got)
	}
}

func TestDoubled(t *testing.T) {
	a := [5]int{1, 2, 3, 4, 5}
	want := [5]int{2, 4, 6, 8, 10}

	if got := Doubled(a); got != want {
		t.Errorf("Doubled(%v) = %v, want %v", a, got, want)
	}
	// An array is a value. Passing one copies it, so the caller's array cannot
	// have changed - unlike a slice, which you will meet in the next challenge.
	if a != ([5]int{1, 2, 3, 4, 5}) {
		t.Errorf("Doubled changed the caller's array to %v, want it untouched", a)
	}
}
