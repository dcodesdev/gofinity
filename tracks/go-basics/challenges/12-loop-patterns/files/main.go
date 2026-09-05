package main

import "fmt"

// SumTo adds every integer from 1 to n. n below 1 sums to 0.
func SumTo(n int) int {
	// TODO
	return 0
}

// CollatzSteps counts the steps the Collatz sequence takes to reach 1: halve an
// even number, triple an odd one and add 1. CollatzSteps(1) is 0, and any n
// below 1 is -1.
func CollatzSteps(n int) int {
	// TODO
	return 0
}

// OddSquares returns the squares of the odd numbers from 1 to n, in order.
// OddSquares(7) is [1 9 25 49].
func OddSquares(n int) []int {
	// TODO
	return nil
}

// NextPowerOfTwo returns the smallest power of two that is greater than or
// equal to n. Anything below 1 gives 1.
func NextPowerOfTwo(n int) int {
	// TODO
	return 0
}

// FirstIndex returns the index of the first item equal to target, or -1 if it
// is not there. It stops looking as soon as it finds one.
func FirstIndex(items []string, target string) int {
	// TODO
	return 0
}

// FindCell searches a grid row by row and returns the row and column of the
// first cell equal to target, or -1, -1 when no cell holds it. Rows may have
// different lengths.
func FindCell(grid [][]int, target int) (int, int) {
	// TODO
	return 0, 0
}

func main() {
	fmt.Println(SumTo(10), CollatzSteps(27), OddSquares(7))
}
