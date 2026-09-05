package main

import "fmt"

// AppendAll appends every value to dst and returns the result. dst itself is
// never assumed to have changed: the return value is the only answer.
func AppendAll(dst []int, values ...int) []int {
	// append may reallocate, in which case the caller's dst still points at the
	// old array. Returning the result is not a style choice, it is the contract.
	return append(dst, values...)
}

// CloneInts returns a slice with the same elements as s and its own backing
// array, so writing to one cannot be seen through the other. A nil slice
// clones to nil.
func CloneInts(s []int) []int {
	if s == nil {
		return nil
	}
	out := make([]int, len(s))
	copy(out, s)
	return out
}

// CopyInto copies as many elements from src into dst as will fit, and returns
// how many it copied. Neither slice is resized.
func CopyInto(dst, src []int) int {
	// copy stops at the shorter of the two, which is exactly the clamping this
	// needs, and it returns the count already.
	return copy(dst, src)
}

// Insert returns a slice with v inserted at index i, leaving s untouched. i is
// clamped into 0..len(s), so Insert never panics.
func Insert(s []int, i, v int) []int {
	if i < 0 {
		i = 0
	}
	if i > len(s) {
		i = len(s)
	}
	// Build the answer in a fresh array rather than shuffling s in place: the
	// contract says s is untouched, and there is no room for an in-place shift
	// without growing s anyway.
	out := make([]int, 0, len(s)+1)
	out = append(out, s[:i]...)
	out = append(out, v)
	out = append(out, s[i:]...)
	return out
}

// RemoveAt returns a slice with the element at index i removed, leaving s
// untouched. An i outside 0..len(s)-1 removes nothing.
func RemoveAt(s []int, i int) []int {
	if i < 0 || i >= len(s) {
		return CloneInts(s)
	}
	out := make([]int, 0, len(s)-1)
	out = append(out, s[:i]...)
	out = append(out, s[i+1:]...)
	return out
}

// Concat returns a new slice holding a followed by b. The result has its own
// backing array: appending to it, or writing through it, must not be visible
// through a or b.
func Concat(a, b []int) []int {
	// append(a, b...) would reuse a's array whenever it has the room, and then
	// the "new" slice would be a view of a. Allocating up front is the only
	// version that is independent whatever a's capacity happens to be.
	out := make([]int, 0, len(a)+len(b))
	out = append(out, a...)
	out = append(out, b...)
	return out
}

func main() {
	fmt.Println(AppendAll(nil, 1, 2, 3))
	fmt.Println(Insert([]int{1, 2, 4}, 2, 3))
	fmt.Println(RemoveAt([]int{1, 2, 3}, 1))
	fmt.Println(Concat([]int{1, 2}, []int{3, 4}))
}
