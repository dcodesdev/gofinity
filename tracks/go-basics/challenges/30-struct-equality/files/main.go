package main

import "fmt"

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
	// TODO
	return false
}

// CountUnique returns how many distinct points the slice holds.
func CountUnique(points []Point) int {
	// TODO
	return 0
}

// EqualConfigs compares two Configs field by field: Name and Where with ==, and
// Hosts element by element. A nil Hosts and an empty Hosts count as equal,
// because both hold no hosts.
func EqualConfigs(a, b Config) bool {
	// TODO
	return false
}

// DeepEqualConfigs compares two Configs with reflect.DeepEqual, which is
// stricter: it tells a nil slice from an empty one.
func DeepEqualConfigs(a, b Config) bool {
	// TODO
	return false
}

// SameNode reports whether two nodes are == to each other. Two nodes pointing
// at different but identical Next nodes are NOT equal, because == on a pointer
// compares addresses.
func SameNode(a, b Node) bool {
	// TODO
	return false
}

// CanCompare reports whether v's dynamic type can be compared with ==. A nil
// interface value has no dynamic type, so it reports false.
func CanCompare(v any) bool {
	// TODO
	return false
}

// Dedup returns points with later duplicates removed, keeping the first
// occurrence of each. No points gives an empty slice.
func Dedup(points []Point) []Point {
	// TODO
	return nil
}

func main() {
	fmt.Println(SamePoint(Point{1, 2}, Point{1, 2}))
	fmt.Println(CountUnique([]Point{{1, 2}, {1, 2}, {3, 4}}))
	fmt.Println(CanCompare(Point{}), CanCompare([]string{}), CanCompare(nil))
}
