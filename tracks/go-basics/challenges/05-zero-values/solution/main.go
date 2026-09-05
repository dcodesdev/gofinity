package main

import "fmt"

// ZeroReport returns one line per zero value, in this order: int, float64,
// string, bool, slice, map, pointer. See description.md for the exact text.
func ZeroReport() string {
	var i int
	var f float64
	var s string
	var b bool
	var sl []int
	var m map[string]int
	var p *int

	return fmt.Sprintf(
		"int: %v\nfloat64: %v\nstring: %q\nbool: %v\nslice: %v nil=%v len=%d\nmap: %v nil=%v len=%d\npointer: %v",
		i, f, s, b, sl, sl == nil, len(sl), m, m == nil, len(m), p,
	)
}

// SumOrZero adds up the numbers. A nil slice sums to 0, with no special case.
func SumOrZero(nums []int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// GrowFromNil appends the numbers to a nil slice and returns the result.
// Its length is len(nums) and it is only nil when nums is empty.
func GrowFromNil(nums ...int) []int {
	var out []int
	for _, n := range nums {
		out = append(out, n)
	}
	return out
}

func main() {
	fmt.Println(ZeroReport())
}
