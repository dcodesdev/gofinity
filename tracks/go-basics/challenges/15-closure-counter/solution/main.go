package main

import "fmt"

// Counter returns a function that gives 1 the first time it is called, 2 the
// second time, and so on. Two counters count independently.
func Counter() func() int {
	// n lives in Counter's frame, but the returned function refers to it, so it
	// outlives the call. Each Counter() makes a new n, which is why two
	// counters do not interfere.
	n := 0
	return func() int {
		n++
		return n
	}
}

// Accumulator returns a function that adds its argument to a running total and
// returns the new total.
func Accumulator() func(int) int {
	total := 0
	return func(delta int) int {
		total += delta
		return total
	}
}

// Multiplier returns a function that multiplies its argument by n.
func Multiplier(n int) func(int) int {
	// A captured parameter works exactly like a captured local.
	return func(x int) int {
		return n * x
	}
}

// Apply returns a new slice holding f applied to each element of nums. It does
// not change nums, and a nil input gives a nil result.
func Apply(nums []int, f func(int) int) []int {
	// A func type is an ordinary type, so a function can be a parameter.
	var out []int
	for _, n := range nums {
		out = append(out, f(n))
	}
	return out
}

// Compose returns a function that applies g first and then f, so
// Compose(double, increment)(3) is double(increment(3)).
func Compose(f, g func(int) int) func(int) int {
	return func(x int) int {
		return f(g(x))
	}
}

// Countdown returns a function that yields from, then from-1, down to 1. Once
// it has run out it returns 0, false every time.
func Countdown(from int) func() (int, bool) {
	next := from
	return func() (int, bool) {
		if next < 1 {
			return 0, false
		}
		n := next
		next--
		return n, true
	}
}

// Multipliers returns one Multiplier per element of ns, in the same order.
func Multipliers(ns []int) []func(int) int {
	var out []func(int) int
	for _, n := range ns {
		// Since Go 1.22 the range variable is fresh on every iteration, so each
		// closure captures its own n. Before that, all of them shared one
		// variable and every function here multiplied by the last element.
		out = append(out, Multiplier(n))
	}
	return out
}

func main() {
	next := Counter()
	fmt.Println(next(), next(), next())
	fmt.Println(Apply([]int{1, 2, 3}, Multiplier(10)))
}
