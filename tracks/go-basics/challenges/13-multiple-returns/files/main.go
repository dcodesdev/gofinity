package main

import "fmt"

// Divide returns a / b and whether the division was possible. Dividing by zero
// is not, and gives 0, false.
func Divide(a, b int) (int, bool) {
	// TODO
	return 0, false
}

// MinMax returns the smallest and largest value in nums. An empty slice has
// neither, so ok is false and both numbers are 0.
func MinMax(nums []int) (lo, hi int, ok bool) {
	// TODO
	return
}

// MaxOnly returns the largest value in nums, or 0 when there is none. Build it
// on MinMax rather than looping again.
func MaxOnly(nums []int) int {
	// TODO
	return 0
}

// SplitName splits a full name at the first space. "Ada Lovelace" is
// "Ada", "Lovelace". A name with no space is all first name, and last is "".
func SplitName(full string) (first, last string) {
	// TODO
	return
}

// Stats returns how many numbers there are, their sum, and their mean. An
// empty slice is 0, 0, 0 with no division at all.
func Stats(nums []int) (count, sum int, mean float64) {
	// TODO
	return
}

func main() {
	q, ok := Divide(7, 2)
	fmt.Println(q, ok)
	fmt.Println(MinMax([]int{3, 1, 4}))
	fmt.Println(Stats([]int{1, 2, 3, 4}))
}
