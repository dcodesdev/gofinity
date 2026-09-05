package main

import "fmt"

// Produce returns a receive-only channel carrying 0, 1, ... n-1 in order, and
// closes it when the last value has been sent.
//
// The shape to learn:
//
//	out := make(chan int)
//	go func() {
//		defer close(out)
//		for i := range n {
//			out <- i
//		}
//	}()
//	return out
//
// Produce returns the channel, not the values: it must not block. The sends
// happen on the goroutine it starts, one at a time, as the caller receives.
//
// n of 0 or less returns a channel that is already closed and empty.
func Produce(n int) <-chan int {
	// TODO
	return nil
}

// Collect receives every value from ch until it is closed and returns them in
// the order they arrived.
//
// A range over a channel is the "receive until closed" loop written out for
// you. Collect of an already-closed channel returns an empty, non-nil slice.
//
// Collect does not close ch: it did not create it.
func Collect(ch <-chan int) []int {
	// TODO
	return nil
}

// Double is a pipeline stage. It returns a channel carrying every value of in,
// doubled, in the same order, closed once in is closed and drained.
//
// Like Produce it must return immediately and forward one value at a time.
// Draining in into a slice first and only then sending would deadlock against
// an unbuffered producer that is itself waiting for the stage after it.
func Double(in <-chan int) <-chan int {
	// TODO
	return nil
}

func main() {
	fmt.Println(Collect(Double(Produce(5))))
}
