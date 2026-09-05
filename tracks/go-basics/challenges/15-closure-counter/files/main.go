package main

import "fmt"

// Counter returns a function that gives 1 the first time it is called, 2 the
// second time, and so on. Two counters count independently.
func Counter() func() int {
	// TODO
	return nil
}

// Accumulator returns a function that adds its argument to a running total and
// returns the new total.
func Accumulator() func(int) int {
	// TODO
	return nil
}

// Multiplier returns a function that multiplies its argument by n.
func Multiplier(n int) func(int) int {
	// TODO
	return nil
}

// Apply returns a new slice holding f applied to each element of nums. It does
// not change nums, and a nil input gives a nil result.
func Apply(nums []int, f func(int) int) []int {
	// TODO
	return nil
}

// Compose returns a function that applies g first and then f, so
// Compose(double, increment)(3) is double(increment(3)).
func Compose(f, g func(int) int) func(int) int {
	// TODO
	return nil
}

// Countdown returns a function that yields from, then from-1, down to 1. Once
// it has run out it returns 0, false every time.
func Countdown(from int) func() (int, bool) {
	// TODO
	return nil
}

// Multipliers returns one Multiplier per element of ns, in the same order.
func Multipliers(ns []int) []func(int) int {
	// TODO
	return nil
}

func main() {
	next := Counter()
	fmt.Println(next(), next(), next())
	fmt.Println(Apply([]int{1, 2, 3}, Multiplier(10)))
}
