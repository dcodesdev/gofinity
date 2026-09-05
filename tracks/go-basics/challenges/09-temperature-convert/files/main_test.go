package main

import (
	"math"
	"testing"
)

// nearly reports whether two floats agree to within a small tolerance. Binary
// floating point rarely lands on an exact decimal, so == is the wrong test.
func nearly(got, want float64) bool {
	return math.Abs(got-want) < 1e-9
}

func TestCToF(t *testing.T) {
	tests := []struct {
		c    float64
		want float64
	}{
		{c: 0, want: 32},
		{c: 100, want: 212},
		{c: -40, want: -40},
		{c: 37, want: 98.6},
		{c: 21.5, want: 70.7},
	}

	for _, tt := range tests {
		if got := CToF(tt.c); !nearly(got, tt.want) {
			t.Errorf("CToF(%v) = %v, want %v", tt.c, got, tt.want)
		}
	}
}

func TestFToC(t *testing.T) {
	tests := []struct {
		f    float64
		want float64
	}{
		{f: 32, want: 0},
		{f: 212, want: 100},
		{f: -40, want: -40},
		{f: 98.6, want: 37},
	}

	for _, tt := range tests {
		if got := FToC(tt.f); !nearly(got, tt.want) {
			t.Errorf("FToC(%v) = %v, want %v", tt.f, got, tt.want)
		}
	}
}

func TestRoundTripIsStable(t *testing.T) {
	for _, c := range []float64{-273.15, -12.5, 0, 18.3, 99.9} {
		if got := FToC(CToF(c)); !nearly(got, c) {
			t.Errorf("FToC(CToF(%v)) = %v, want %v", c, got, c)
		}
	}
}

func TestRoundTenth(t *testing.T) {
	tests := []struct {
		in   float64
		want float64
	}{
		{in: 0, want: 0},
		{in: 1.24, want: 1.2},
		{in: 1.25, want: 1.3},
		{in: 1.26, want: 1.3},
		{in: -1.25, want: -1.3},
		{in: -1.24, want: -1.2},
		{in: 98.64, want: 98.6},
	}

	for _, tt := range tests {
		if got := RoundTenth(tt.in); !nearly(got, tt.want) {
			t.Errorf("RoundTenth(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestFahrenheitWhole(t *testing.T) {
	tests := []struct {
		c    float64
		want int
	}{
		{c: 0, want: 32},
		{c: 100, want: 212},
		{c: 37, want: 99},
		{c: 21.5, want: 71},
		{c: -40, want: -40},
		{c: -17.5, want: 1},
		{c: -18, want: 0},
		{c: -20, want: -4},
	}

	for _, tt := range tests {
		if got := FahrenheitWhole(tt.c); got != tt.want {
			t.Errorf("FahrenheitWhole(%v) = %d, want %d", tt.c, got, tt.want)
		}
	}
}

func TestReport(t *testing.T) {
	tests := []struct {
		c    float64
		want string
	}{
		{c: 21.5, want: "21.5°C = 70.7°F"},
		{c: 0, want: "0.0°C = 32.0°F"},
		{c: 100, want: "100.0°C = 212.0°F"},
		{c: -40, want: "-40.0°C = -40.0°F"},
	}

	for _, tt := range tests {
		if got := Report(tt.c); got != tt.want {
			t.Errorf("Report(%v) = %q, want %q", tt.c, got, tt.want)
		}
	}
}
