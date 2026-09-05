package main

import "fmt"

// Sum adds every argument it is given. Sum() with no arguments is 0.
func Sum(nums ...int) int {
	// TODO
	return 0
}

// Largest returns the biggest of the values it is given. The first one is a
// required parameter, so there is always an answer.
func Largest(first int, rest ...int) int {
	// TODO
	return 0
}

// SumAll adds every number in every group. SumAll([]int{1, 2}, []int{3}) is 6.
// Reuse Sum rather than writing a second loop over the numbers.
func SumAll(groups ...[]int) int {
	// TODO
	return 0
}

// Average returns the mean of its arguments, and false when there are none.
func Average(nums ...int) (float64, bool) {
	// TODO
	return 0, false
}

// Describe formats its arguments with fmt.Sprintf. The whole job is forwarding
// one variadic to another.
func Describe(format string, args ...any) string {
	// TODO
	return ""
}

// SumEvens adds only the even numbers among its arguments, without changing
// what the caller passed in.
func SumEvens(nums ...int) int {
	// TODO
	return 0
}

func main() {
	fmt.Println(Sum(1, 2, 3), Sum())
	fmt.Println(SumAll([]int{1, 2}, []int{3}))
	fmt.Println(Describe("total: %d", 6))
}
