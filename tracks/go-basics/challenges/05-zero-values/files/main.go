package main

import "fmt"

// ZeroReport returns one line per zero value, in this order: int, float64,
// string, bool, slice, map, pointer. See description.md for the exact text.
func ZeroReport() string {
	// TODO: declare one variable per type with `var`, no initial value, and
	// format a line for each with fmt.Sprintf.
	return ""
}

// SumOrZero adds up the numbers. A nil slice sums to 0, with no special case.
func SumOrZero(nums []int) int {
	// TODO
	return 0
}

// GrowFromNil appends the numbers to a nil slice and returns the result.
// Its length is len(nums) and it is only nil when nums is empty.
func GrowFromNil(nums ...int) []int {
	// TODO
	return nil
}

func main() {
	fmt.Println(ZeroReport())
}
