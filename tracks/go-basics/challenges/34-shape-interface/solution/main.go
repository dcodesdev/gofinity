package main

import (
	"fmt"
	"math"
)

// Shape is the whole contract: two methods, nothing else. Nothing declares
// that it implements Shape - a type implements it the moment it has both
// methods.
type Shape interface {
	Area() float64
	Perimeter() float64
}

// Named is a second, even smaller interface. A type can satisfy both, one, or
// neither, and it never has to say which.
type Named interface {
	Name() string
}

// Rect is a rectangle. It satisfies Shape and not Named.
type Rect struct {
	W, H float64
}

// Area returns W * H.
func (r Rect) Area() float64 {
	return r.W * r.H
}

// Perimeter returns twice the sum of the sides.
func (r Rect) Perimeter() float64 {
	return 2 * (r.W + r.H)
}

// Circle is a circle of radius R. It satisfies Shape and Named.
type Circle struct {
	R float64
}

// Area returns pi * R * R. math.Pi is the constant you want.
func (c Circle) Area() float64 {
	return math.Pi * c.R * c.R
}

// Perimeter returns the circumference, 2 * pi * R.
func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.R
}

// Name returns "circle". Having this method is all it takes to satisfy Named.
func (c Circle) Name() string {
	return "circle"
}

// TotalArea adds up the areas of every shape. The slice can hold Rects and
// Circles side by side, because both fit in a Shape.
func TotalArea(shapes []Shape) float64 {
	total := 0.0
	for _, s := range shapes {
		total += s.Area()
	}
	return total
}

// Largest returns the shape with the greatest area, or nil when the slice is
// empty. The first shape wins a tie.
func Largest(shapes []Shape) Shape {
	// The zero value of an interface is nil, so an empty slice falls out of
	// the loop with best still nil.
	var best Shape
	for _, s := range shapes {
		if best == nil || s.Area() > best.Area() {
			best = s
		}
	}
	return best
}

// Describe renders a shape as one line. When the shape also satisfies Named,
// use its name:
//
//	"circle: area 3.14, perimeter 6.28"
//
// and when it does not, use the word "shape":
//
//	"shape: area 6.00, perimeter 10.00"
//
// Both numbers are formatted with two digits after the point. A type assertion
// to an interface, s.(Named), is how you ask whether the value behind s has a
// Name method.
func Describe(s Shape) string {
	label := "shape"
	// The comma-ok form asks the question instead of insisting on the answer:
	// a plain s.(Named) would panic on a Rect.
	if n, ok := s.(Named); ok {
		label = n.Name()
	}
	return fmt.Sprintf("%s: area %.2f, perimeter %.2f", label, s.Area(), s.Perimeter())
}

// CountAtLeast counts the shapes whose area is min or greater.
func CountAtLeast(shapes []Shape, min float64) int {
	n := 0
	for _, s := range shapes {
		if s.Area() >= min {
			n++
		}
	}
	return n
}

func main() {
	shapes := []Shape{Rect{W: 2, H: 3}, Circle{R: 1}}
	for _, s := range shapes {
		fmt.Println(Describe(s))
	}
	fmt.Printf("total %.2f\n", TotalArea(shapes))
	_ = math.Pi
}
