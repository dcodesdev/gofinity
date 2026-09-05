package main

import (
	"math"
	"testing"
)

func TestSum(t *testing.T) {
	if got := Sum(); got != 0 {
		t.Errorf("Sum() = %d, want 0", got)
	}
	if got := Sum(4); got != 4 {
		t.Errorf("Sum(4) = %d, want 4", got)
	}
	if got := Sum(1, 2, 3, 4); got != 10 {
		t.Errorf("Sum(1, 2, 3, 4) = %d, want 10", got)
	}
	if got := Sum(-3, 3, -1); got != -1 {
		t.Errorf("Sum(-3, 3, -1) = %d, want -1", got)
	}
}

func TestSumSpreadsASlice(t *testing.T) {
	nums := []int{5, 10, 15}
	if got := Sum(nums...); got != 30 {
		t.Errorf("Sum(nums...) = %d, want 30", got)
	}
	if got := Sum([]int(nil)...); got != 0 {
		t.Errorf("Sum(nil...) = %d, want 0", got)
	}
}

func TestLargest(t *testing.T) {
	tests := []struct {
		first int
		rest  []int
		want  int
	}{
		{first: 7, rest: nil, want: 7},
		{first: 3, rest: []int{1, 4, 1}, want: 4},
		{first: 9, rest: []int{2, 5}, want: 9},
		{first: -5, rest: []int{-2, -9}, want: -2},
	}

	for _, tt := range tests {
		if got := Largest(tt.first, tt.rest...); got != tt.want {
			t.Errorf("Largest(%d, %v...) = %d, want %d", tt.first, tt.rest, got, tt.want)
		}
	}
}

func TestSumAll(t *testing.T) {
	tests := []struct {
		groups [][]int
		want   int
	}{
		{groups: [][]int{{1, 2}, {3}}, want: 6},
		{groups: [][]int{{1, 2, 3}, {}, {4}}, want: 10},
		{groups: [][]int{{}}, want: 0},
		{groups: [][]int{nil, {5}}, want: 5},
		{groups: nil, want: 0},
	}

	for _, tt := range tests {
		if got := SumAll(tt.groups...); got != tt.want {
			t.Errorf("SumAll(%v...) = %d, want %d", tt.groups, got, tt.want)
		}
	}
}

func TestAverage(t *testing.T) {
	tests := []struct {
		nums   []int
		want   float64
		wantOk bool
	}{
		{nums: []int{1, 2, 3, 4}, want: 2.5, wantOk: true},
		{nums: []int{5}, want: 5, wantOk: true},
		{nums: []int{1, 2}, want: 1.5, wantOk: true},
		{nums: nil, want: 0, wantOk: false},
	}

	for _, tt := range tests {
		got, ok := Average(tt.nums...)
		// Floats are approximations, so compare with a tolerance, not with ==.
		if math.Abs(got-tt.want) > 1e-9 || ok != tt.wantOk {
			t.Errorf("Average(%v...) = %g, %t, want %g, %t", tt.nums, got, ok, tt.want, tt.wantOk)
		}
	}
}

func TestDescribe(t *testing.T) {
	tests := []struct {
		format string
		args   []any
		want   string
	}{
		{format: "total: %d", args: []any{6}, want: "total: 6"},
		{format: "%s has %d", args: []any{"go", 2}, want: "go has 2"},
		{format: "%s", args: []any{"solo"}, want: "solo"},
		{format: "nothing to fill in", args: nil, want: "nothing to fill in"},
	}

	for _, tt := range tests {
		if got := Describe(tt.format, tt.args...); got != tt.want {
			t.Errorf("Describe(%q, %v...) = %q, want %q", tt.format, tt.args, got, tt.want)
		}
	}
}

func TestSumEvens(t *testing.T) {
	if got := SumEvens(1, 2, 3, 4, 5, 6); got != 12 {
		t.Errorf("SumEvens(1..6) = %d, want 12", got)
	}
	if got := SumEvens(1, 3, 5); got != 0 {
		t.Errorf("SumEvens(1, 3, 5) = %d, want 0", got)
	}
	if got := SumEvens(-4, 7); got != -4 {
		t.Errorf("SumEvens(-4, 7) = %d, want -4", got)
	}
	if got := SumEvens(); got != 0 {
		t.Errorf("SumEvens() = %d, want 0", got)
	}
}

func TestSumEvensLeavesTheCallersSliceAlone(t *testing.T) {
	// Spreading a slice does not copy it: inside SumEvens, nums is the same
	// backing array. Reordering or overwriting it there is visible out here.
	nums := []int{1, 2, 3, 4}
	SumEvens(nums...)
	for i, want := range []int{1, 2, 3, 4} {
		if nums[i] != want {
			t.Fatalf("SumEvens wrote through its variadic parameter: got %v, want [1 2 3 4]", nums)
		}
	}
}
