package main

import (
	"slices"
	"testing"
)

func TestSumTo(t *testing.T) {
	tests := []struct {
		n    int
		want int
	}{
		{n: 1, want: 1},
		{n: 3, want: 6},
		{n: 10, want: 55},
		{n: 100, want: 5050},
		{n: 0, want: 0},
		{n: -5, want: 0},
	}

	for _, tt := range tests {
		if got := SumTo(tt.n); got != tt.want {
			t.Errorf("SumTo(%d) = %d, want %d", tt.n, got, tt.want)
		}
	}
}

func TestCollatzSteps(t *testing.T) {
	tests := []struct {
		n    int
		want int
	}{
		{n: 1, want: 0},
		{n: 2, want: 1},
		{n: 3, want: 7},
		{n: 6, want: 8},
		{n: 27, want: 111},
		{n: 0, want: -1},
		{n: -4, want: -1},
	}

	for _, tt := range tests {
		if got := CollatzSteps(tt.n); got != tt.want {
			t.Errorf("CollatzSteps(%d) = %d, want %d", tt.n, got, tt.want)
		}
	}
}

func TestOddSquares(t *testing.T) {
	tests := []struct {
		n    int
		want []int
	}{
		{n: 7, want: []int{1, 9, 25, 49}},
		{n: 8, want: []int{1, 9, 25, 49}},
		{n: 1, want: []int{1}},
		{n: 0, want: nil},
		{n: -3, want: nil},
	}

	for _, tt := range tests {
		got := OddSquares(tt.n)
		if len(got) != len(tt.want) || !slices.Equal(got, tt.want) {
			t.Errorf("OddSquares(%d) = %v, want %v", tt.n, got, tt.want)
		}
	}
}

func TestNextPowerOfTwo(t *testing.T) {
	tests := []struct {
		n    int
		want int
	}{
		{n: 1, want: 1},
		{n: 2, want: 2},
		{n: 3, want: 4},
		{n: 5, want: 8},
		{n: 64, want: 64},
		{n: 65, want: 128},
		{n: 0, want: 1},
		{n: -9, want: 1},
	}

	for _, tt := range tests {
		if got := NextPowerOfTwo(tt.n); got != tt.want {
			t.Errorf("NextPowerOfTwo(%d) = %d, want %d", tt.n, got, tt.want)
		}
	}
}

func TestFirstIndex(t *testing.T) {
	items := []string{"go", "rust", "go", "zig"}

	tests := []struct {
		target string
		want   int
	}{
		{target: "go", want: 0},
		{target: "rust", want: 1},
		{target: "zig", want: 3},
		{target: "ada", want: -1},
		{target: "", want: -1},
	}

	for _, tt := range tests {
		if got := FirstIndex(items, tt.target); got != tt.want {
			t.Errorf("FirstIndex(%v, %q) = %d, want %d", items, tt.target, got, tt.want)
		}
	}

	if got := FirstIndex(nil, "go"); got != -1 {
		t.Errorf("FirstIndex(nil, \"go\") = %d, want -1", got)
	}
}

func TestFindCell(t *testing.T) {
	grid := [][]int{
		{1, 2, 3},
		{4},
		{},
		{5, 6, 3},
	}

	tests := []struct {
		target  int
		wantRow int
		wantCol int
	}{
		{target: 1, wantRow: 0, wantCol: 0},
		{target: 3, wantRow: 0, wantCol: 2},
		{target: 4, wantRow: 1, wantCol: 0},
		{target: 6, wantRow: 3, wantCol: 1},
		{target: 99, wantRow: -1, wantCol: -1},
	}

	for _, tt := range tests {
		row, col := FindCell(grid, tt.target)
		if row != tt.wantRow || col != tt.wantCol {
			t.Errorf("FindCell(grid, %d) = %d, %d, want %d, %d", tt.target, row, col, tt.wantRow, tt.wantCol)
		}
	}

	if row, col := FindCell(nil, 1); row != -1 || col != -1 {
		t.Errorf("FindCell(nil, 1) = %d, %d, want -1, -1", row, col)
	}
}
