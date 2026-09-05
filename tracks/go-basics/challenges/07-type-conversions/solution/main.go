package main

import (
	"fmt"
	"math"
)

// Truncate converts f to an int by dropping the fractional part, toward zero.
// Truncate(2.9) is 2 and Truncate(-2.9) is -2.
func Truncate(f float64) int {
	return int(f)
}

// Average returns the mean of nums. An empty slice averages to 0.
func Average(nums []int) float64 {
	if len(nums) == 0 {
		return 0
	}
	sum := 0
	for _, n := range nums {
		sum += n
	}
	// Both operands are converted first: sum/len(nums) would divide as ints
	// and throw the fraction away before the result ever became a float64.
	return float64(sum) / float64(len(nums))
}

// FitsInt8 converts n to an int8 and reports whether the conversion kept the
// value. When n is out of range the returned int8 is the wrapped result, which
// is what Go produces, and the bool is false.
func FitsInt8(n int) (int8, bool) {
	return int8(n), n >= math.MinInt8 && n <= math.MaxInt8
}

func main() {
	fmt.Println(Truncate(2.9), Average([]int{1, 2}))
	fmt.Println(FitsInt8(200))
}
