package main

import "fmt"

// Sum adds every argument it is given. Sum() with no arguments is 0.
func Sum(nums ...int) int {
	// Inside the function, nums is an ordinary []int - nil when no arguments
	// were passed, which ranges zero times and needs no special case.
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// Largest returns the biggest of the values it is given. The first one is a
// required parameter, so there is always an answer.
func Largest(first int, rest ...int) int {
	// One required parameter in front of the variadic is how you say "at least
	// one" in the signature instead of at runtime.
	best := first
	for _, n := range rest {
		if n > best {
			best = n
		}
	}
	return best
}

// SumAll adds every number in every group. SumAll([]int{1, 2}, []int{3}) is 6.
// Reuse Sum rather than writing a second loop over the numbers.
func SumAll(groups ...[]int) int {
	total := 0
	for _, group := range groups {
		// `group...` spreads the slice back out into Sum's arguments. Without
		// the dots this would not compile: a []int is not an int.
		total += Sum(group...)
	}
	return total
}

// Average returns the mean of its arguments, and false when there are none.
func Average(nums ...int) (float64, bool) {
	if len(nums) == 0 {
		return 0, false
	}
	// Convert both operands, or this is integer division and the mean of 1 and
	// 2 comes out as 1.
	return float64(Sum(nums...)) / float64(len(nums)), true
}

// Describe formats its arguments with fmt.Sprintf. The whole job is forwarding
// one variadic to another.
func Describe(format string, args ...any) string {
	// The dots are what forwards the arguments. `fmt.Sprintf(format, args)`
	// would pass the slice itself as a single argument and print it as [go 2].
	return fmt.Sprintf(format, args...)
}

// SumEvens adds only the even numbers among its arguments, without changing
// what the caller passed in.
func SumEvens(nums ...int) int {
	// Read only. When the caller spreads a slice, nums shares that backing
	// array, so sorting or overwriting it here would reach back out.
	total := 0
	for _, n := range nums {
		if n%2 == 0 {
			total += n
		}
	}
	return total
}

func main() {
	fmt.Println(Sum(1, 2, 3), Sum())
	fmt.Println(SumAll([]int{1, 2}, []int{3}))
	fmt.Println(Describe("total: %d", 6))
}
