package main

import (
	"fmt"
	"time"
)

// TryRecv receives one value from ch if one is ready right now, and otherwise
// gives up immediately. It never blocks.
//
// The shape to learn:
//
//	select {
//	case v := <-ch:
//		return v, true
//	default:
//		return 0, false
//	}
//
// A closed channel is always ready, so it takes the first case and reports the
// zero value with ok true. That is the same thing a plain receive does.
func TryRecv(ch <-chan int) (int, bool) {
	// TODO
	return 0, false
}

// RecvTimeout receives one value from ch, waiting at most d for it. It reports
// whether a value arrived.
//
// Replace the default of TryRecv with a clock:
//
//	case <-time.After(d):
//
// time.After returns a channel that receives after d, which is all a timeout
// is in Go: another case in the select. A closed channel is ready at once, so
// like TryRecv this returns the zero value with ok true straight away rather
// than waiting out d.
func RecvTimeout(ch <-chan int, d time.Duration) (int, bool) {
	// TODO
	return 0, false
}

// Merge fans several channels into one. Every value from every input appears
// exactly once, in no particular order between inputs, and the output closes
// once all of the inputs have closed.
//
// One goroutine per input ranges over it and forwards into out, a WaitGroup
// counts them, and a separate goroutine waits and then closes out. The closer
// must be its own goroutine: waiting on the calling one would block before
// Merge could return the channel anybody is meant to read.
//
// Merging nothing returns an already closed channel.
func Merge(chs ...<-chan int) <-chan int {
	// TODO
	return nil
}

// GenerateUntil sends 0, 1, 2, ... on the returned channel and stops when done
// is closed, closing its output on the way out.
//
// The send and the done receive belong to the same select:
//
//	select {
//	case out <- i:
//	case <-done:
//		return
//	}
//
// Written as a bare send followed by a check, the goroutine parks on a send
// nobody is waiting for and never looks at done again. That is a goroutine
// leak: no crash, no error, just a goroutine that lives for ever.
func GenerateUntil(done <-chan struct{}) <-chan int {
	// TODO
	return nil
}

func main() {
	done := make(chan struct{})
	ch := GenerateUntil(done)
	fmt.Println(<-ch, <-ch, <-ch)
	close(done)

	fmt.Println(RecvTimeout(make(chan int), 10*time.Millisecond))
}
