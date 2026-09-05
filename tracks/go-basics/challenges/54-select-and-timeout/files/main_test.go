package main

import (
	"reflect"
	"runtime"
	"slices"
	"testing"
	"time"
)

// guard is how long any one step may take. Everything correct here finishes in
// microseconds, so a step still waiting after this long is stuck - and the
// file stays well inside the sandbox's wall clock even when every guard fires.
const guard = 300 * time.Millisecond

// closed is a channel that is ready for ever: a receive from it returns the
// zero value immediately.
func closedChan() chan int {
	ch := make(chan int)
	close(ch)
	return ch
}

// call runs fn on its own goroutine and fails rather than letting the binary
// hang. Every mistake in this challenge blocks instead of returning a wrong
// answer, so nothing is called directly.
func call[T any](t *testing.T, what string, fn func() T) T {
	t.Helper()
	done := make(chan T, 1)
	go func() { done <- fn() }()
	select {
	case v := <-done:
		return v
	case <-time.After(guard):
		var zero T
		t.Fatalf("%s did not return in time", what)
		return zero
	}
}

type recvResult struct {
	value int
	ok    bool
}

func tryRecv(t *testing.T, what string, ch <-chan int) recvResult {
	t.Helper()
	return call(t, what, func() recvResult {
		v, ok := TryRecv(ch)
		return recvResult{v, ok}
	})
}

func recvTimeout(t *testing.T, what string, ch <-chan int, d time.Duration) recvResult {
	t.Helper()
	return call(t, what, func() recvResult {
		v, ok := RecvTimeout(ch, d)
		return recvResult{v, ok}
	})
}

// recvOne takes one value from ch without blocking for ever.
func recvOne(t *testing.T, what string, ch <-chan int) (int, bool) {
	t.Helper()
	select {
	case v, ok := <-ch:
		return v, ok
	case <-time.After(guard):
		t.Fatalf("%s: nothing arrived in time", what)
		return 0, false
	}
}

// isOpen reports whether ch is still open, without consuming a value and
// without waiting: only a closed channel is ready when nothing was sent.
func isOpen(t *testing.T, ch <-chan int) bool {
	t.Helper()
	select {
	case _, ok := <-ch:
		return ok
	case <-time.After(10 * time.Millisecond):
		return true
	}
}

// drain receives everything left in ch and reports whether it ended closed.
func drain(t *testing.T, what string, ch <-chan int) ([]int, bool) {
	t.Helper()
	got := []int{}
	deadline := time.After(guard)
	for {
		select {
		case v, ok := <-ch:
			if !ok {
				return got, true
			}
			got = append(got, v)
		case <-deadline:
			t.Fatalf("%s: the channel never closed", what)
			return got, false
		}
	}
}

func TestTryRecvTakesAWaitingValue(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 7
	if got := tryRecv(t, "TryRecv of a channel holding 7", ch); got != (recvResult{7, true}) {
		t.Errorf("TryRecv = %v, %v, want 7, true", got.value, got.ok)
	}
}

func TestTryRecvGivesUpOnAnEmptyChannel(t *testing.T) {
	// Nobody will ever send here, so anything but an immediate false is a
	// TryRecv that blocked.
	if got := tryRecv(t, "TryRecv of an empty channel", make(chan int)); got.ok {
		t.Errorf("TryRecv of an empty open channel = %v, true, want 0, false", got.value)
	}
}

func TestTryRecvOfAClosedChannelSucceeds(t *testing.T) {
	// A closed channel is ready, so it takes the receive case and never the
	// default, exactly as a plain receive would.
	if got := tryRecv(t, "TryRecv of a closed channel", closedChan()); got != (recvResult{0, true}) {
		t.Errorf("TryRecv of a closed channel = %v, %v, want 0, true", got.value, got.ok)
	}
}

func TestRecvTimeoutReturnsAWaitingValueWithoutWaiting(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 99
	// The deadline is far longer than the guard, so a RecvTimeout that waits
	// it out rather than taking the ready case fails here.
	if got := recvTimeout(t, "RecvTimeout of a channel holding 99", ch, time.Minute); got != (recvResult{99, true}) {
		t.Errorf("RecvTimeout = %v, %v, want 99, true", got.value, got.ok)
	}
}

func TestRecvTimeoutTakesAValueThatArrivesLater(t *testing.T) {
	ch := make(chan int)
	go func() { ch <- 5 }()
	if got := recvTimeout(t, "RecvTimeout of a channel about to receive 5", ch, time.Minute); got != (recvResult{5, true}) {
		t.Errorf("RecvTimeout = %v, %v, want 5, true", got.value, got.ok)
	}
}

func TestRecvTimeoutGivesUpWhenNothingArrives(t *testing.T) {
	const d = 20 * time.Millisecond
	start := time.Now()
	got := recvTimeout(t, "RecvTimeout of an empty channel", make(chan int), d)
	elapsed := time.Since(start)

	if got.ok {
		t.Errorf("RecvTimeout on an empty channel = %v, true, want 0, false", got.value)
	}
	if elapsed < d {
		t.Errorf("RecvTimeout returned after %v, want it to wait at least %v", elapsed, d)
	}
}

func TestRecvTimeoutOfAClosedChannelSucceedsAtOnce(t *testing.T) {
	if got := recvTimeout(t, "RecvTimeout of a closed channel", closedChan(), time.Minute); got != (recvResult{0, true}) {
		t.Errorf("RecvTimeout of a closed channel = %v, %v, want 0, true", got.value, got.ok)
	}
}

func TestMergeCarriesEveryValueFromEveryInput(t *testing.T) {
	a := make(chan int, 3)
	b := make(chan int, 2)
	for _, v := range []int{1, 2, 3} {
		a <- v
	}
	for _, v := range []int{10, 20} {
		b <- v
	}
	close(a)
	close(b)

	got, _ := drain(t, "Merge of two closed channels", Merge(a, b))
	slices.Sort(got)
	if want := []int{1, 2, 3, 10, 20}; !reflect.DeepEqual(got, want) {
		t.Errorf("Merge produced %v, want %v in some order", got, want)
	}
}

func TestMergeOfNothingIsAClosedChannel(t *testing.T) {
	if got, closed := drain(t, "Merge()", Merge()); !closed || len(got) != 0 {
		t.Errorf("Merge() produced %v, closed=%v, want no values and closed", got, closed)
	}
}

func TestMergeReadsItsInputsAtTheSameTime(t *testing.T) {
	// Nothing is ever sent on a. A Merge that drained its inputs one after
	// another would still be waiting on a while b's value sits there unread.
	a := make(chan int)
	b := make(chan int)
	out := Merge(a, b)
	go func() { b <- 42 }()

	if v, ok := recvOne(t, "Merge with a value on its second input", out); !ok || v != 42 {
		t.Errorf("Merge produced %v, %v, want 42, true", v, ok)
	}
}

func TestMergeClosesOnlyAfterEveryInputHasClosed(t *testing.T) {
	a := make(chan int)
	b := make(chan int)
	out := Merge(a, b)

	close(a)
	if !isOpen(t, out) {
		t.Fatal("Merge closed its output after one of two inputs closed, want it open until both have")
	}

	go func() { b <- 8 }()
	if v, ok := recvOne(t, "Merge after one input closed", out); !ok || v != 8 {
		t.Fatalf("Merge produced %v, %v after one input closed, want 8, true", v, ok)
	}

	close(b)
	if _, closed := drain(t, "Merge after both inputs closed", out); !closed {
		t.Error("Merge never closed its output, want it closed once every input had")
	}
}

func TestGenerateUntilCountsUpFromZero(t *testing.T) {
	done := make(chan struct{})
	defer close(done)

	ch := GenerateUntil(done)
	for want := range 5 {
		got, ok := recvOne(t, "GenerateUntil", ch)
		if !ok {
			t.Fatalf("GenerateUntil closed after %d values, want it to keep counting until done closes", want)
		}
		if got != want {
			t.Fatalf("GenerateUntil sent %d, want %d", got, want)
		}
	}
}

func TestGenerateUntilClosesItsOutputWhenDoneCloses(t *testing.T) {
	done := make(chan struct{})
	ch := GenerateUntil(done)
	recvOne(t, "GenerateUntil", ch)
	close(done)

	if _, closed := drain(t, "GenerateUntil after done closed", ch); !closed {
		t.Error("GenerateUntil never closed its output, want it closed on the way out")
	}
}

func TestGenerateUntilStopsWhenDoneIsAlreadyClosed(t *testing.T) {
	done := make(chan struct{})
	close(done)

	if _, closed := drain(t, "GenerateUntil with done already closed", GenerateUntil(done)); !closed {
		t.Error("GenerateUntil kept its output open although done was closed before it started")
	}
}

func TestGenerateUntilLeavesNoGoroutineBehind(t *testing.T) {
	// The real reason the send and the done receive share a select. Written as
	// a send followed by a check, the goroutine parks on a send nobody is
	// waiting for and never returns, however long done has been closed.
	base := runtime.NumGoroutine()
	done := make(chan struct{})
	ch := GenerateUntil(done)
	recvOne(t, "GenerateUntil", ch)
	close(done)

	// Nothing receives from ch again: only a select can notice done now.
	deadline := time.Now().Add(guard)
	for runtime.NumGoroutine() > base {
		if time.Now().After(deadline) {
			t.Fatalf("%d goroutines are still running, want back to %d - GenerateUntil's goroutine is stuck on a send", runtime.NumGoroutine(), base)
		}
		time.Sleep(time.Millisecond)
	}
}
