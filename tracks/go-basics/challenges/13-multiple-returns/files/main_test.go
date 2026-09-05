package main

import (
	"math"
	"testing"
)

func TestDivide(t *testing.T) {
	tests := []struct {
		a, b   int
		want   int
		wantOk bool
	}{
		{a: 7, b: 2, want: 3, wantOk: true},
		{a: 8, b: 2, want: 4, wantOk: true},
		{a: -9, b: 3, want: -3, wantOk: true},
		{a: 0, b: 5, want: 0, wantOk: true},
		{a: 5, b: 0, want: 0, wantOk: false},
		{a: 0, b: 0, want: 0, wantOk: false},
	}

	for _, tt := range tests {
		got, ok := Divide(tt.a, tt.b)
		if got != tt.want || ok != tt.wantOk {
			t.Errorf("Divide(%d, %d) = %d, %t, want %d, %t", tt.a, tt.b, got, ok, tt.want, tt.wantOk)
		}
	}
}

func TestMinMax(t *testing.T) {
	tests := []struct {
		nums   []int
		wantLo int
		wantHi int
		wantOk bool
	}{
		{nums: []int{3, 1, 4, 1, 5}, wantLo: 1, wantHi: 5, wantOk: true},
		{nums: []int{42}, wantLo: 42, wantHi: 42, wantOk: true},
		{nums: []int{-7, -2, -9}, wantLo: -9, wantHi: -2, wantOk: true},
		{nums: []int{}, wantLo: 0, wantHi: 0, wantOk: false},
		{nums: nil, wantLo: 0, wantHi: 0, wantOk: false},
	}

	for _, tt := range tests {
		lo, hi, ok := MinMax(tt.nums)
		if lo != tt.wantLo || hi != tt.wantHi || ok != tt.wantOk {
			t.Errorf("MinMax(%v) = %d, %d, %t, want %d, %d, %t", tt.nums, lo, hi, ok, tt.wantLo, tt.wantHi, tt.wantOk)
		}
	}
}

func TestMinMaxLeavesInputAlone(t *testing.T) {
	nums := []int{3, 1, 4, 1, 5}
	MinMax(nums)
	for i, want := range []int{3, 1, 4, 1, 5} {
		if nums[i] != want {
			t.Fatalf("MinMax sorted its argument: got %v, want [3 1 4 1 5]", nums)
		}
	}
}

func TestMaxOnly(t *testing.T) {
	tests := []struct {
		nums []int
		want int
	}{
		{nums: []int{3, 1, 4}, want: 4},
		{nums: []int{-7, -2}, want: -2},
		{nums: []int{0}, want: 0},
		{nums: nil, want: 0},
	}

	for _, tt := range tests {
		if got := MaxOnly(tt.nums); got != tt.want {
			t.Errorf("MaxOnly(%v) = %d, want %d", tt.nums, got, tt.want)
		}
	}
}

func TestSplitName(t *testing.T) {
	tests := []struct {
		full      string
		wantFirst string
		wantLast  string
	}{
		{full: "Ada Lovelace", wantFirst: "Ada", wantLast: "Lovelace"},
		{full: "Rob Pike", wantFirst: "Rob", wantLast: "Pike"},
		{full: "Grace Brewster Hopper", wantFirst: "Grace", wantLast: "Brewster Hopper"},
		{full: "Prince", wantFirst: "Prince", wantLast: ""},
		{full: "", wantFirst: "", wantLast: ""},
	}

	for _, tt := range tests {
		first, last := SplitName(tt.full)
		if first != tt.wantFirst || last != tt.wantLast {
			t.Errorf("SplitName(%q) = %q, %q, want %q, %q", tt.full, first, last, tt.wantFirst, tt.wantLast)
		}
	}
}

func TestStats(t *testing.T) {
	tests := []struct {
		nums      []int
		wantCount int
		wantSum   int
		wantMean  float64
	}{
		{nums: []int{1, 2, 3, 4}, wantCount: 4, wantSum: 10, wantMean: 2.5},
		{nums: []int{5}, wantCount: 1, wantSum: 5, wantMean: 5},
		{nums: []int{-2, 2}, wantCount: 2, wantSum: 0, wantMean: 0},
		{nums: nil, wantCount: 0, wantSum: 0, wantMean: 0},
	}

	for _, tt := range tests {
		count, sum, mean := Stats(tt.nums)
		// Floats are approximations, so compare the mean with a tolerance
		// rather than with ==.
		if count != tt.wantCount || sum != tt.wantSum || math.Abs(mean-tt.wantMean) > 1e-9 {
			t.Errorf("Stats(%v) = %d, %d, %g, want %d, %d, %g", tt.nums, count, sum, mean, tt.wantCount, tt.wantSum, tt.wantMean)
		}
	}
}

func TestStatsMeanIsNotIntegerDivision(t *testing.T) {
	_, _, mean := Stats([]int{1, 2})
	if math.Abs(mean-1.5) > 1e-9 {
		t.Errorf("Stats([1 2]) mean = %g, want 1.5 (convert to float64 before dividing)", mean)
	}
}
