package main

import (
	"math"
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		in   float64
		want int
	}{
		{in: 0, want: 0},
		{in: 2.9, want: 2},
		{in: 2.0, want: 2},
		{in: -2.9, want: -2},
		{in: -0.5, want: 0},
		{in: 7.999999, want: 7},
	}

	for _, tt := range tests {
		if got := Truncate(tt.in); got != tt.want {
			t.Errorf("Truncate(%v) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestAverage(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want float64
	}{
		{name: "empty", in: nil, want: 0},
		{name: "one", in: []int{4}, want: 4},
		{name: "exact", in: []int{2, 4}, want: 3},
		{name: "fractional", in: []int{1, 2}, want: 1.5},
		{name: "three quarters", in: []int{0, 1, 1, 1}, want: 0.75},
		{name: "negative", in: []int{-3, 2}, want: -0.5},
	}

	for _, tt := range tests {
		got := Average(tt.in)
		if math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("%s: Average(%v) = %v, want %v", tt.name, tt.in, got, tt.want)
		}
	}
}

func TestFitsInt8(t *testing.T) {
	tests := []struct {
		in     int
		want   int8
		wantOK bool
	}{
		{in: 0, want: 0, wantOK: true},
		{in: 42, want: 42, wantOK: true},
		{in: 127, want: 127, wantOK: true},
		{in: -128, want: -128, wantOK: true},
		{in: 128, want: -128, wantOK: false},
		{in: 200, want: -56, wantOK: false},
		{in: -129, want: 127, wantOK: false},
	}

	for _, tt := range tests {
		got, ok := FitsInt8(tt.in)
		if got != tt.want || ok != tt.wantOK {
			t.Errorf("FitsInt8(%d) = (%d, %v), want (%d, %v)", tt.in, got, ok, tt.want, tt.wantOK)
		}
	}
}
