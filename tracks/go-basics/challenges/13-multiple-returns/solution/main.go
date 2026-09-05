package main

import (
	"fmt"
	"strings"
)

// Divide returns a / b and whether the division was possible. Dividing by zero
// is not, and gives 0, false.
func Divide(a, b int) (int, bool) {
	// The comma-ok shape: the second result says whether the first one means
	// anything. Guard first and return the zero value, never a half-answer.
	if b == 0 {
		return 0, false
	}
	return a / b, true
}

// MinMax returns the smallest and largest value in nums. An empty slice has
// neither, so ok is false and both numbers are 0.
func MinMax(nums []int) (lo, hi int, ok bool) {
	// Named results are declared and zeroed on entry, so an empty slice can
	// take a bare `return`: it hands back 0, 0, false.
	if len(nums) == 0 {
		return
	}

	lo, hi, ok = nums[0], nums[0], true
	for _, n := range nums[1:] {
		if n < lo {
			lo = n
		}
		if n > hi {
			hi = n
		}
	}
	return lo, hi, ok
}

// MaxOnly returns the largest value in nums, or 0 when there is none. Build it
// on MinMax rather than looping again.
func MaxOnly(nums []int) int {
	// `_` throws the minimum away. Go will not let an ordinary unused variable
	// compile, and the blank identifier is how you say "yes, on purpose".
	_, hi, ok := MinMax(nums)
	if !ok {
		return 0
	}
	return hi
}

// SplitName splits a full name at the first space. "Ada Lovelace" is
// "Ada", "Lovelace". A name with no space is all first name, and last is "".
func SplitName(full string) (first, last string) {
	// strings.Cut is itself a comma-ok function: the value plus "did it
	// happen". Ignoring the bool here is safe because Cut leaves before as the
	// whole string when there is no separator.
	first, last, _ = strings.Cut(full, " ")
	return first, last
}

// Stats returns how many numbers there are, their sum, and their mean. An
// empty slice is 0, 0, 0 with no division at all.
func Stats(nums []int) (count, sum int, mean float64) {
	if len(nums) == 0 {
		return
	}

	count = len(nums)
	for _, n := range nums {
		sum += n
	}
	// Both operands have to be float64 or this is integer division and every
	// mean loses its fraction.
	mean = float64(sum) / float64(count)
	return count, sum, mean
}

func main() {
	q, ok := Divide(7, 2)
	fmt.Println(q, ok)
	fmt.Println(MinMax([]int{3, 1, 4}))
	fmt.Println(Stats([]int{1, 2, 3, 4}))
}
