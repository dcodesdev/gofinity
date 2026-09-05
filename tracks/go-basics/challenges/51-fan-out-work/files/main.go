package main

import (
	"fmt"
	"strconv"
)

// Assignments returns which item indices each worker takes, striding rather
// than blocking: worker w takes w, w+workers, w+2*workers, and so on.
//
//	Assignments(9, 3) -> [[0 3 6] [1 4 7] [2 5 8]]
//	Assignments(4, 6) -> [[0] [1] [2] [3]]
//	Assignments(3, 0) -> [[0 1 2]]
//	Assignments(0, 4) -> no workers
//
// Striding shares out a run of expensive items instead of dumping it on one
// goroutine. Index i still belongs to exactly one worker, because i%workers has
// exactly one value, so the safety argument is unchanged.
//
// workers < 1 is treated as 1, and a worker is never given an empty list.
func Assignments(items, workers int) [][]int {
	// TODO
	return nil
}

// FanOut applies f to every element of in using at most `workers` goroutines,
// and returns the results in the order of in - not in the order they finished.
//
// Writing each result to its own index is what buys that ordering. Appending
// from several goroutines would be both a data race and a shuffle.
//
// f is called exactly once per element. An empty input returns an empty,
// non-nil slice.
func FanOut[T, R any](in []T, workers int, f func(T) R) []R {
	// TODO
	return nil
}

// FanOutErr is FanOut for an f that can fail.
//
// If every call succeeds it returns the results and a nil error. If any call
// fails it returns nil results and the error from the *lowest failing index*,
// so the same input always produces the same error however the goroutines were
// scheduled. Returning whichever error arrived first would be a coin toss.
//
// It does not stop early: f is still called once per element even after a
// failure. Cancelling the rest needs a channel and then a context, which are
// the next two concepts.
func FanOutErr[T, R any](in []T, workers int, f func(T) (R, error)) ([]R, error) {
	// TODO
	return nil, nil
}

func main() {
	fmt.Println(Assignments(9, 3))
	fmt.Println(FanOut([]int{1, 2, 3}, 2, func(v int) int { return v * v }))
	fmt.Println(FanOutErr([]string{"1", "2", "x"}, 2, strconv.Atoi))
}
