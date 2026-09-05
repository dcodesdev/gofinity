package main

import "fmt"

// Describe reports the length and capacity of s in the form "len=3 cap=5".
func Describe(s []int) string {
	// TODO
	return ""
}

// First returns the first element of s and true. On an empty or nil slice it
// returns 0 and false rather than indexing.
func First(s []int) (int, bool) {
	// TODO
	return 0, false
}

// Last returns the final element of s and true. On an empty or nil slice it
// returns 0 and false.
func Last(s []int) (int, bool) {
	// TODO
	return 0, false
}

// Head returns a slice of the first n elements of s. A negative n gives an
// empty result, and an n larger than the slice gives all of it - Head never
// panics.
func Head(s []int, n int) []int {
	// TODO
	return nil
}

// Tail returns a slice of the last n elements of s, with the same clamping
// rules as Head.
func Tail(s []int, n int) []int {
	// TODO
	return nil
}

// SumArray adds up a fixed-size array of five ints.
func SumArray(a [5]int) int {
	// TODO
	return 0
}

// Doubled returns an array with every element of a doubled. An array is a
// value, so the caller's array is not touched.
func Doubled(a [5]int) [5]int {
	// TODO
	return a
}

func main() {
	s := make([]int, 3, 8)
	fmt.Println(Describe(s))
	fmt.Println(Head([]int{1, 2, 3, 4}, 2), Tail([]int{1, 2, 3, 4}, 2))
	fmt.Println(SumArray([5]int{1, 2, 3, 4, 5}))
}
