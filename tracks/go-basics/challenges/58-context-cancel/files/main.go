package main

import (
	"context"
	"fmt"
)

// Wait blocks until ready receives a value or ctx finishes, whichever happens
// first.
//
// It returns nil when ready fired, and ctx.Err() when the context finished
// first - context.Canceled or context.DeadlineExceeded, never a message of
// your own. The caller decides what a cancellation means; your job is to
// report it faithfully.
func Wait(ctx context.Context, ready <-chan struct{}) error {
	// TODO
	return nil
}

// Cancelled reports whether ctx has already finished, without blocking for a
// moment either way.
//
// A receive from a channel blocks; a select with a default does not.
func Cancelled(ctx context.Context) bool {
	// TODO
	return false
}

// CountTicks receives from ticks until ctx finishes or ticks is closed, and
// returns how many values it received.
//
// A closed channel is always ready to receive, so a receive alone cannot tell
// "closed" from "a value arrived" - the two-value form can.
func CountTicks(ctx context.Context, ticks <-chan int) int {
	// TODO
	return 0
}

// Produce returns a channel carrying every value in order, sent from its own
// goroutine.
//
// It stops early when ctx finishes, and it closes the channel on the way out
// whichever way it stops. A producer that forgets to close leaves every ranging
// consumer blocked for ever, and a producer that sends without watching
// ctx.Done blocks for ever itself once the consumer walks away.
func Produce(ctx context.Context, values []int) <-chan int {
	// TODO
	return nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	ready := make(chan struct{})
	close(ready)
	fmt.Println(Wait(ctx, ready))

	for v := range Produce(ctx, []int{1, 2, 3}) {
		fmt.Println(v)
	}

	cancel()
	fmt.Println(Cancelled(ctx), Wait(ctx, make(chan struct{})))
}
