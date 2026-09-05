package main

import (
	"slices"
	"testing"
)

func TestScaleInPlace(t *testing.T) {
	s := []int{1, 2, 3}
	Scale(s, 3)
	if want := []int{3, 6, 9}; !slices.Equal(s, want) {
		t.Errorf("after Scale(s, 3), s = %v, want %v", s, want)
	}
}

func TestScaleThroughAView(t *testing.T) {
	// The view and the original share one array, so scaling the view is
	// visible in the middle of the original.
	s := []int{1, 2, 3, 4, 5}
	Scale(s[1:4], 10)
	if want := []int{1, 20, 30, 40, 5}; !slices.Equal(s, want) {
		t.Errorf("after Scale(s[1:4], 10), s = %v, want %v", s, want)
	}
}

func TestScaleEmpty(t *testing.T) {
	Scale(nil, 2)
	s := []int{}
	Scale(s, 2)
	if len(s) != 0 {
		t.Errorf("Scale(empty) produced %v", s)
	}
}

func TestScaledCopy(t *testing.T) {
	s := []int{1, 2, 3}
	got := ScaledCopy(s, 2)
	if want := []int{2, 4, 6}; !slices.Equal(got, want) {
		t.Errorf("ScaledCopy(%v, 2) = %v, want %v", s, got, want)
	}
	if want := []int{1, 2, 3}; !slices.Equal(s, want) {
		t.Errorf("ScaledCopy changed its argument to %v, want %v", s, want)
	}
	got[0] = 99
	if s[0] != 1 {
		t.Errorf("the result shares an array with the argument: s = %v", s)
	}
}

func TestSplitAt(t *testing.T) {
	tests := []struct {
		i         int
		wantLeft  []int
		wantRight []int
	}{
		{2, []int{1, 2}, []int{3, 4}},
		{0, []int{}, []int{1, 2, 3, 4}},
		{4, []int{1, 2, 3, 4}, []int{}},
		{9, []int{1, 2, 3, 4}, []int{}},
		{-1, []int{}, []int{1, 2, 3, 4}},
	}
	for _, tt := range tests {
		s := []int{1, 2, 3, 4}
		left, right := SplitAt(s, tt.i)
		if !slices.Equal(left, tt.wantLeft) || !slices.Equal(right, tt.wantRight) {
			t.Errorf("SplitAt(s, %d) = %v, %v, want %v, %v", tt.i, left, right, tt.wantLeft, tt.wantRight)
		}
	}
}

func TestSplitAtSharesTheArray(t *testing.T) {
	s := []int{1, 2, 3, 4}
	left, right := SplitAt(s, 2)

	left[0] = 90
	right[0] = 92
	if want := []int{90, 2, 92, 4}; !slices.Equal(s, want) {
		t.Errorf("s = %v, want %v: both halves are views of s, not copies", s, want)
	}
}

func TestWindow(t *testing.T) {
	s := []int{1, 2, 3, 4, 5}
	tests := []struct {
		lo, hi int
		want   []int
	}{
		{1, 4, []int{2, 3, 4}},
		{0, 5, []int{1, 2, 3, 4, 5}},
		{2, 2, []int{}},
		{-3, 2, []int{1, 2}},
		{3, 99, []int{4, 5}},
		{4, 1, []int{}},
		{9, 12, []int{}},
	}
	for _, tt := range tests {
		if got := Window(s, tt.lo, tt.hi); !slices.Equal(got, tt.want) {
			t.Errorf("Window(s, %d, %d) = %v, want %v", tt.lo, tt.hi, got, tt.want)
		}
	}
	if got := Window(nil, 0, 3); len(got) != 0 {
		t.Errorf("Window(nil, 0, 3) = %v, want an empty result", got)
	}
}

func TestWindowReadsThroughToS(t *testing.T) {
	s := []int{1, 2, 3, 4, 5}
	w := Window(s, 1, 4)
	s[2] = 30
	if want := []int{2, 30, 4}; !slices.Equal(w, want) {
		t.Errorf("w = %v, want %v: a window is a view, not a copy", w, want)
	}
}

func TestAppendingToAWindowDoesNotTouchS(t *testing.T) {
	// A plain s[1:4] keeps the spare capacity of s, so appending to it writes
	// over s[4]. The three-index form is what prevents that.
	s := []int{1, 2, 3, 4, 5}
	w := Window(s, 1, 4)

	w = append(w, 99)
	if want := []int{1, 2, 3, 4, 5}; !slices.Equal(s, want) {
		t.Errorf("after appending to the window, s = %v, want %v", s, want)
	}
	if want := []int{2, 3, 4, 99}; !slices.Equal(w, want) {
		t.Errorf("the window = %v, want %v", w, want)
	}
}

func TestDedup(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{"runs", []int{1, 1, 2, 3, 3, 3}, []int{1, 2, 3}},
		{"no duplicates", []int{1, 2, 3}, []int{1, 2, 3}},
		{"all the same", []int{7, 7, 7, 7}, []int{7}},
		{"one", []int{5}, []int{5}},
		{"empty", []int{}, []int{}},
		{"a pair at the end", []int{1, 2, 2}, []int{1, 2}},
		{"a pair at the start", []int{1, 1, 2}, []int{1, 2}},
		{"negatives", []int{-3, -3, -1, 0, 0}, []int{-3, -1, 0}},
	}
	for _, tt := range tests {
		if got := Dedup(tt.in); !slices.Equal(got, tt.want) {
			t.Errorf("Dedup(%s) = %v, want %v", tt.name, got, tt.want)
		}
	}
	if got := Dedup(nil); len(got) != 0 {
		t.Errorf("Dedup(nil) = %v, want an empty result", got)
	}
}

func TestDedupIsInPlace(t *testing.T) {
	s := []int{1, 1, 2, 3, 3, 3}
	got := Dedup(s)

	if want := []int{1, 2, 3}; !slices.Equal(got, want) {
		t.Fatalf("Dedup = %v, want %v", got, want)
	}
	if len(s) != 6 {
		t.Fatalf("the caller's slice is now length %d; Dedup returns a shorter slice, it does not resize s", len(s))
	}
	// The result is the front of s, so writing through it is visible in s.
	got[0] = 90
	if s[0] != 90 {
		t.Errorf("s = %v, want the result to be a prefix of s rather than a fresh slice", s)
	}
}
