package main

import "fmt"

// Truncate converts f to an int by dropping the fractional part, toward zero.
// Truncate(2.9) is 2 and Truncate(-2.9) is -2.
func Truncate(f float64) int {
	// TODO
	return 0
}

// Average returns the mean of nums. An empty slice averages to 0.
func Average(nums []int) float64 {
	// TODO
	return 0
}

// FitsInt8 converts n to an int8 and reports whether the conversion kept the
// value. When n is out of range the returned int8 is the wrapped result, which
// is what Go produces, and the bool is false.
func FitsInt8(n int) (int8, bool) {
	// TODO
	return 0, false
}

func main() {
	fmt.Println(Truncate(2.9), Average([]int{1, 2}))
	fmt.Println(FitsInt8(200))
}
