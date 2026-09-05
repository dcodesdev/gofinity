package main

import "fmt"

// Map returns a new slice holding f(v) for every v in s, in order. The input
// element type and the output element type are independent: Map over a
// []string with a func(string) int gives you a []int.
//
// An empty input gives an empty, non-nil slice.
func Map[T, U any](s []T, f func(T) U) []U {
	// TODO
	return nil
}

// Filter returns a new slice holding the elements of s for which keep returns
// true, in order. Only one type parameter this time: what goes in comes out.
//
// An empty input, or one where nothing is kept, gives an empty, non-nil slice.
func Filter[T any](s []T, keep func(T) bool) []T {
	// TODO
	return nil
}

// Reduce folds s into a single value: it starts at init and replaces the
// accumulator with f(acc, v) for each element, left to right.
//
//	Reduce([]int{1, 2, 3}, 0, func(a, v int) int { return a + v }) == 6
//
// The accumulator type U is not the element type T - reducing a []string to an
// int must compile.
func Reduce[T, U any](s []T, init U, f func(U, T) U) U {
	// TODO
	return init
}

// Keys returns the keys of m in unspecified order. A map key is comparable by
// definition, so K is constrained to comparable; the value type is not
// constrained at all.
//
// An empty or nil map gives an empty, non-nil slice.
func Keys[K comparable, V any](m map[K]V) []K {
	// TODO
	return nil
}

// Contains reports whether v appears in s. Comparing with == needs more than
// any, which is why T here is comparable.
func Contains[T comparable](s []T, v T) bool {
	// TODO
	return false
}

// IndexOf returns the index of the first element of s equal to v, or -1 when
// there is none.
func IndexOf[T comparable](s []T, v T) int {
	// TODO
	return 0
}

// First returns the first element of s for which pred is true, and whether
// there was one. On a miss it returns the zero value of T, which is spelled
// with a var declaration because you cannot write nil or 0 for a T.
func First[T any](s []T, pred func(T) bool) (T, bool) {
	// TODO
	var zero T
	return zero, false
}

func main() {
	words := []string{"go", "is", "fun"}
	fmt.Println(Map(words, func(s string) int { return len(s) }))
	fmt.Println(Filter([]int{1, 2, 3, 4}, func(n int) bool { return n%2 == 0 }))
}
