package main

import (
	"slices"
	"testing"
)

func TestLIFOReversesTheDeferOrder(t *testing.T) {
	cases := []struct {
		name   string
		labels []string
		want   []string
	}{
		{name: "three labels", labels: []string{"a", "b", "c"}, want: []string{"c", "b", "a"}},
		{name: "one label", labels: []string{"only"}, want: []string{"only"}},
		{name: "none", labels: nil, want: nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := LIFO(tc.labels)
			if !slices.Equal(got, tc.want) {
				t.Errorf("LIFO(%v) = %v, want %v", tc.labels, got, tc.want)
			}
		})
	}
}

func TestLIFODoesNotTouchItsInput(t *testing.T) {
	labels := []string{"a", "b"}
	LIFO(labels)
	if !slices.Equal(labels, []string{"a", "b"}) {
		t.Errorf("labels = %v after LIFO, want it left alone", labels)
	}
}

func TestSteps(t *testing.T) {
	want := []string{"enter", "work", "cleanup-2", "cleanup-1"}
	if got := Steps(); !slices.Equal(got, want) {
		t.Errorf("Steps() = %v, want %v (deferred calls run last, in reverse)", got, want)
	}
}

func TestCapturedValue(t *testing.T) {
	if got := CapturedValue(); got != 1 {
		t.Errorf("CapturedValue() = %d, want 1: a deferred call's arguments are evaluated at the defer statement", got)
	}
}

func TestCapturedVariable(t *testing.T) {
	if got := CapturedVariable(); got != 99 {
		t.Errorf("CapturedVariable() = %d, want 99: a closure reads the variable when it runs, not when it is deferred", got)
	}
}

func TestDoubleResult(t *testing.T) {
	cases := []struct{ in, want int }{
		{in: 21, want: 42},
		{in: 0, want: 0},
		{in: -3, want: -6},
	}

	for _, tc := range cases {
		if got := DoubleResult(tc.in); got != tc.want {
			t.Errorf("DoubleResult(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}
