package main

import (
	"fmt"
	"reflect"
	"slices"
)

// Point is comparable: every field is comparable, so == works on it and it can
// be a map key.
type Point struct {
	X, Y int
}

// Config is not comparable, because a slice field is not. Writing a == b on two
// Configs does not compile.
type Config struct {
	Name  string
	Where Point
	Hosts []string
}

// Node has a pointer field. Pointers are comparable, so Node is - but == on two
// Nodes compares the pointers themselves, not what they point at.
type Node struct {
	Label string
	Next  *Node
}

// SamePoint reports whether two points are equal.
func SamePoint(a, b Point) bool {
	// Every field of Point is comparable, so the whole struct is, and == is a
	// field-by-field comparison rather than a pointer comparison.
	return a == b
}

// CountUnique returns how many distinct points the slice holds.
func CountUnique(points []Point) int {
	// A comparable struct is a legal map key, which makes "distinct" a
	// one-liner. A Config could never be used here.
	seen := make(map[Point]struct{}, len(points))
	for _, p := range points {
		seen[p] = struct{}{}
	}
	return len(seen)
}

// EqualConfigs compares two Configs field by field: Name and Where with ==, and
// Hosts element by element. A nil Hosts and an empty Hosts count as equal,
// because both hold no hosts.
func EqualConfigs(a, b Config) bool {
	if a.Name != b.Name || a.Where != b.Where {
		return false
	}
	// slices.Equal walks the elements and treats nil and empty as the same
	// thing, because both have length 0.
	return slices.Equal(a.Hosts, b.Hosts)
}

// DeepEqualConfigs compares two Configs with reflect.DeepEqual, which is
// stricter: it tells a nil slice from an empty one.
func DeepEqualConfigs(a, b Config) bool {
	// DeepEqual works on anything, at the cost of reflection and of being
	// stricter than you usually want: nil != empty, here.
	return reflect.DeepEqual(a, b)
}

// SameNode reports whether two nodes are == to each other. Two nodes pointing
// at different but identical Next nodes are NOT equal, because == on a pointer
// compares addresses.
func SameNode(a, b Node) bool {
	// Label compares by value and Next compares by address, in the same
	// expression. That asymmetry is the trap.
	return a == b
}

// CanCompare reports whether v's dynamic type can be compared with ==. A nil
// interface value has no dynamic type, so it reports false.
func CanCompare(v any) bool {
	t := reflect.TypeOf(v)
	if t == nil {
		// A nil interface carries no type at all, so there is nothing to ask.
		return false
	}
	return t.Comparable()
}

// Dedup returns points with later duplicates removed, keeping the first
// occurrence of each. No points gives an empty slice.
func Dedup(points []Point) []Point {
	out := []Point{}
	seen := make(map[Point]bool, len(points))
	for _, p := range points {
		if seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return out
}

func main() {
	fmt.Println(SamePoint(Point{1, 2}, Point{1, 2}))
	fmt.Println(CountUnique([]Point{{1, 2}, {1, 2}, {3, 4}}))
	fmt.Println(CanCompare(Point{}), CanCompare([]string{}), CanCompare(nil))
}
