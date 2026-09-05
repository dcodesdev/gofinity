package main

import "fmt"

// Describe reports the length and capacity of s in the form "len=3 cap=5".
func Describe(s []int) string {
	// len is how many elements you can index; cap is how many the backing
	// array has room for before append has to allocate a new one.
	return fmt.Sprintf("len=%d cap=%d", len(s), cap(s))
}

// First returns the first element of s and true. On an empty or nil slice it
// returns 0 and false rather than indexing.
func First(s []int) (int, bool) {
	if len(s) == 0 {
		return 0, false
	}
	return s[0], true
}

// Last returns the final element of s and true. On an empty or nil slice it
// returns 0 and false.
func Last(s []int) (int, bool) {
	if len(s) == 0 {
		return 0, false
	}
	return s[len(s)-1], true
}

// Head returns a slice of the first n elements of s. A negative n gives an
// empty result, and an n larger than the slice gives all of it - Head never
// panics.
func Head(s []int, n int) []int {
	// The clamp is the whole job: s[:n] panics the moment n leaves 0..len(s).
	if n < 0 {
		n = 0
	}
	if n > len(s) {
		n = len(s)
	}
	return s[:n]
}

// Tail returns a slice of the last n elements of s, with the same clamping
// rules as Head.
func Tail(s []int, n int) []int {
	if n < 0 {
		n = 0
	}
	if n > len(s) {
		n = len(s)
	}
	return s[len(s)-n:]
}

// SumArray adds up a fixed-size array of five ints.
func SumArray(a [5]int) int {
	total := 0
	for _, n := range a {
		total += n
	}
	return total
}

// Doubled returns an array with every element of a doubled. An array is a
// value, so the caller's array is not touched.
func Doubled(a [5]int) [5]int {
	// a is already a copy: the assignment happened when the call was made, so
	// writing into it here cannot reach the caller's array.
	for i := range a {
		a[i] *= 2
	}
	return a
}

func main() {
	s := make([]int, 3, 8)
	fmt.Println(Describe(s))
	fmt.Println(Head([]int{1, 2, 3, 4}, 2), Tail([]int{1, 2, 3, 4}, 2))
	fmt.Println(SumArray([5]int{1, 2, 3, 4, 5}))
}
