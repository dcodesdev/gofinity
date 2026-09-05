package main

import "fmt"

// AppendAll appends every value to dst and returns the result. dst itself is
// never assumed to have changed: the return value is the only answer.
func AppendAll(dst []int, values ...int) []int {
	// TODO
	return nil
}

// CloneInts returns a slice with the same elements as s and its own backing
// array, so writing to one cannot be seen through the other. A nil slice
// clones to nil.
func CloneInts(s []int) []int {
	// TODO
	return nil
}

// CopyInto copies as many elements from src into dst as will fit, and returns
// how many it copied. Neither slice is resized.
func CopyInto(dst, src []int) int {
	// TODO
	return 0
}

// Insert returns a slice with v inserted at index i, leaving s untouched. i is
// clamped into 0..len(s), so Insert never panics.
func Insert(s []int, i, v int) []int {
	// TODO
	return nil
}

// RemoveAt returns a slice with the element at index i removed, leaving s
// untouched. An i outside 0..len(s)-1 removes nothing.
func RemoveAt(s []int, i int) []int {
	// TODO
	return nil
}

// Concat returns a new slice holding a followed by b. The result has its own
// backing array: appending to it, or writing through it, must not be visible
// through a or b.
func Concat(a, b []int) []int {
	// TODO
	return nil
}

func main() {
	fmt.Println(AppendAll(nil, 1, 2, 3))
	fmt.Println(Insert([]int{1, 2, 4}, 2, 3))
	fmt.Println(RemoveAt([]int{1, 2, 3}, 1))
	fmt.Println(Concat([]int{1, 2}, []int{3, 4}))
}
