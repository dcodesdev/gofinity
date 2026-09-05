package main

import (
	"math"
	"testing"
)

// Neither Rect nor Circle mentions Shape anywhere, and these lines still
// compile: satisfaction in Go is structural and checked here, at the point of
// use. Circle also satisfies Named; Rect does not, and there is no line for it.
var (
	_ Shape = Rect{}
	_ Shape = Circle{}
	_ Named = Circle{}
)

func closeTo(got, want float64) bool {
	return math.Abs(got-want) < 1e-9
}

func TestRect(t *testing.T) {
	r := Rect{W: 2, H: 3}
	if got := r.Area(); !closeTo(got, 6) {
		t.Errorf("Rect{2, 3}.Area() = %v, want 6", got)
	}
	if got := r.Perimeter(); !closeTo(got, 10) {
		t.Errorf("Rect{2, 3}.Perimeter() = %v, want 10", got)
	}
	if got := (Rect{}).Area(); !closeTo(got, 0) {
		t.Errorf("the zero Rect has Area() = %v, want 0", got)
	}
}

func TestCircle(t *testing.T) {
	c := Circle{R: 2}
	if got := c.Area(); !closeTo(got, 4*math.Pi) {
		t.Errorf("Circle{2}.Area() = %v, want %v", got, 4*math.Pi)
	}
	if got := c.Perimeter(); !closeTo(got, 4*math.Pi) {
		t.Errorf("Circle{2}.Perimeter() = %v, want %v", got, 4*math.Pi)
	}
	if got := c.Name(); got != "circle" {
		t.Errorf("Circle.Name() = %q, want %q", got, "circle")
	}
}

func TestTotalArea(t *testing.T) {
	shapes := []Shape{Rect{W: 2, H: 3}, Rect{W: 1, H: 4}, Circle{R: 1}}
	want := 10 + math.Pi
	if got := TotalArea(shapes); !closeTo(got, want) {
		t.Errorf("TotalArea = %v, want %v", got, want)
	}
	if got := TotalArea(nil); !closeTo(got, 0) {
		t.Errorf("TotalArea(nil) = %v, want 0", got)
	}
}

func TestLargest(t *testing.T) {
	if got := Largest(nil); got != nil {
		t.Errorf("Largest(nil) = %v, want nil", got)
	}

	big := Rect{W: 10, H: 10}
	shapes := []Shape{Circle{R: 1}, big, Rect{W: 2, H: 2}}
	got := Largest(shapes)
	if got != Shape(big) {
		t.Errorf("Largest = %v, want %v", got, big)
	}

	// Equal areas: the first one wins.
	first := Rect{W: 3, H: 3}
	tie := []Shape{first, Rect{W: 9, H: 1}}
	if got := Largest(tie); got != Shape(first) {
		t.Errorf("Largest on a tie = %v, want the first shape %v", got, first)
	}
}

func TestDescribe(t *testing.T) {
	tests := []struct {
		name string
		in   Shape
		want string
	}{
		{"named", Circle{R: 1}, "circle: area 3.14, perimeter 6.28"},
		{"unnamed", Rect{W: 2, H: 3}, "shape: area 6.00, perimeter 10.00"},
		{"zero rect", Rect{}, "shape: area 0.00, perimeter 0.00"},
	}
	for _, tt := range tests {
		if got := Describe(tt.in); got != tt.want {
			t.Errorf("%s: Describe = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestCountAtLeast(t *testing.T) {
	shapes := []Shape{Rect{W: 1, H: 1}, Rect{W: 2, H: 2}, Circle{R: 2}}
	if got := CountAtLeast(shapes, 4); got != 2 {
		t.Errorf("CountAtLeast(shapes, 4) = %d, want 2 - the bound is inclusive", got)
	}
	if got := CountAtLeast(shapes, 0); got != 3 {
		t.Errorf("CountAtLeast(shapes, 0) = %d, want 3", got)
	}
	if got := CountAtLeast(shapes, 100); got != 0 {
		t.Errorf("CountAtLeast(shapes, 100) = %d, want 0", got)
	}
}
