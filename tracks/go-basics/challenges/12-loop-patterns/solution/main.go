package main

import "fmt"

// SumTo adds every integer from 1 to n. n below 1 sums to 0.
func SumTo(n int) int {
	// The three-clause form: init, condition, post. The counter is scoped to
	// the loop and does not exist after it.
	total := 0
	for i := 1; i <= n; i++ {
		total += i
	}
	return total
}

// CollatzSteps counts the steps the Collatz sequence takes to reach 1: halve an
// even number, triple an odd one and add 1. CollatzSteps(1) is 0, and any n
// below 1 is -1.
func CollatzSteps(n int) int {
	if n < 1 {
		return -1
	}

	// Condition only: Go's `for` with one clause is what other languages spell
	// `while`.
	steps := 0
	for n != 1 {
		if n%2 == 0 {
			n /= 2
		} else {
			n = 3*n + 1
		}
		steps++
	}
	return steps
}

// OddSquares returns the squares of the odd numbers from 1 to n, in order.
// OddSquares(7) is [1 9 25 49].
func OddSquares(n int) []int {
	var out []int
	for i := 1; i <= n; i++ {
		if i%2 == 0 {
			// `continue` skips the rest of this iteration and runs the post
			// statement, so the counter still advances.
			continue
		}
		out = append(out, i*i)
	}
	return out
}

// NextPowerOfTwo returns the smallest power of two that is greater than or
// equal to n. Anything below 1 gives 1.
func NextPowerOfTwo(n int) int {
	// No clauses at all: an infinite loop, left with `break`.
	p := 1
	for {
		if p >= n {
			break
		}
		p *= 2
	}
	return p
}

// FirstIndex returns the index of the first item equal to target, or -1 if it
// is not there. It stops looking as soon as it finds one.
func FirstIndex(items []string, target string) int {
	for i, item := range items {
		if item == target {
			return i
		}
	}
	return -1
}

// FindCell searches a grid row by row and returns the row and column of the
// first cell equal to target, or -1, -1 when no cell holds it. Rows may have
// different lengths.
func FindCell(grid [][]int, target int) (int, int) {
	// A label lets `break` leave the *outer* loop. Without one it would only
	// end the inner range and the search would carry on with the next row.
	row, col := -1, -1
search:
	for r, cells := range grid {
		for c, cell := range cells {
			if cell == target {
				row, col = r, c
				break search
			}
		}
	}
	return row, col
}

func main() {
	fmt.Println(SumTo(10), CollatzSteps(27), OddSquares(7))
}
