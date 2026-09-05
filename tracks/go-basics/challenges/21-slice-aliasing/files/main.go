package main

import "fmt"

// Scale multiplies every element of s by factor, in place. It returns nothing:
// the caller sees the change through the slice it already holds.
func Scale(s []int, factor int) {
	// TODO
}

// ScaledCopy returns a new slice with every element of s multiplied by factor,
// leaving s untouched.
func ScaledCopy(s []int, factor int) []int {
	// TODO
	return nil
}

// SplitAt returns two views of s: everything before index i, and everything
// from i onward. Both share s's backing array, so writing through either one
// is visible in s. i is clamped into 0..len(s).
func SplitAt(s []int, i int) (left, right []int) {
	// TODO
	return nil, nil
}

// Window returns the elements of s from lo up to but not including hi, with
// both bounds clamped into range. Appending to the result must never overwrite
// an element of s.
func Window(s []int, lo, hi int) []int {
	// TODO
	return nil
}

// Dedup removes runs of equal neighbours from s in place and returns the
// shortened prefix. s is sorted, so equal elements are always adjacent. The
// elements past the returned length are left as whatever the shuffle put there.
func Dedup(s []int) []int {
	// TODO
	return nil
}

func main() {
	s := []int{1, 2, 3, 4}
	Scale(s, 2)
	fmt.Println(s)
	left, right := SplitAt(s, 2)
	fmt.Println(left, right)
	fmt.Println(Dedup([]int{1, 1, 2, 3, 3, 3}))
}
