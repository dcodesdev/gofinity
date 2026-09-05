package main

import "fmt"

// RunAll runs every task concurrently and returns only after all of them have
// finished.
//
// The shape to learn:
//
//	var wg sync.WaitGroup
//	for _, task := range tasks {
//		wg.Add(1)            // before the go statement
//		go func() {
//			defer wg.Done() // exactly once, whatever happens
//			task()
//		}()
//	}
//	wg.Wait()
//
// An empty slice is not a special case: the counter never leaves zero and
// Wait returns immediately.
func RunAll(tasks []func()) {
	// TODO
}

// Squares returns the squares of 0..n-1, computed by one goroutine per
// element: Squares(4) is []int{0, 1, 4, 9}.
//
// Preallocate the result with make and have each goroutine write its own
// index. Two goroutines appending to the same slice is a data race on the
// slice header; two goroutines writing different elements of a slice that
// already has the right length is not.
//
// Squares(0) returns an empty slice, not nil.
func Squares(n int) []int {
	// TODO
	return nil
}

// Gather runs every function concurrently and returns their results in the
// order of the input slice, whatever order they actually finished in.
//
// This is Squares generalised, and it is the pattern behind most "do these N
// things at once" code you will write. Index the output rather than appending
// to it and the ordering takes care of itself.
func Gather(fns []func() int) []int {
	// TODO
	return nil
}

func main() {
	fmt.Println(Squares(5))
	fmt.Println(Gather([]func() int{
		func() int { return 1 },
		func() int { return 2 },
	}))
}
