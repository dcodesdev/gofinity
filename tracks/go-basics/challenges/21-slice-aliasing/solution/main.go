package main

import "fmt"

// Scale multiplies every element of s by factor, in place. It returns nothing:
// the caller sees the change through the slice it already holds.
func Scale(s []int, factor int) {
	// The slice header was copied when the call was made, but the pointer
	// inside it was copied too, so both headers address the same array.
	for i := range s {
		s[i] *= factor
	}
}

// ScaledCopy returns a new slice with every element of s multiplied by factor,
// leaving s untouched.
func ScaledCopy(s []int, factor int) []int {
	out := make([]int, len(s))
	for i, n := range s {
		out[i] = n * factor
	}
	return out
}

// SplitAt returns two views of s: everything before index i, and everything
// from i onward. Both share s's backing array, so writing through either one
// is visible in s. i is clamped into 0..len(s).
func SplitAt(s []int, i int) (left, right []int) {
	if i < 0 {
		i = 0
	}
	if i > len(s) {
		i = len(s)
	}
	return s[:i], s[i:]
}

// Window returns the elements of s from lo up to but not including hi, with
// both bounds clamped into range. Appending to the result must never overwrite
// an element of s.
func Window(s []int, lo, hi int) []int {
	if lo < 0 {
		lo = 0
	}
	if lo > len(s) {
		lo = len(s)
	}
	if hi > len(s) {
		hi = len(s)
	}
	if hi < lo {
		hi = lo
	}
	// The third index is the capacity bound. s[lo:hi:hi] has no spare capacity,
	// so the first append allocates a new array instead of writing over s[hi].
	return s[lo:hi:hi]
}

// Dedup removes runs of equal neighbours from s in place and returns the
// shortened prefix. s is sorted, so equal elements are always adjacent. The
// elements past the returned length are left as whatever the shuffle put there.
func Dedup(s []int) []int {
	if len(s) == 0 {
		return s
	}
	// n is where the next kept element goes. Reading ahead of writing is what
	// makes the in-place version safe.
	n := 1
	for _, v := range s[1:] {
		if v != s[n-1] {
			s[n] = v
			n++
		}
	}
	return s[:n]
}

func main() {
	s := []int{1, 2, 3, 4}
	Scale(s, 2)
	fmt.Println(s)
	left, right := SplitAt(s, 2)
	fmt.Println(left, right)
	fmt.Println(Dedup([]int{1, 1, 2, 3, 3, 3}))
}
