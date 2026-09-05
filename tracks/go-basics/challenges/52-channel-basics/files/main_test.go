package main

import (
	"reflect"
	"testing"
	"time"
)

// guard is how long any one step is allowed to take. Everything here finishes
// in microseconds when it is correct, so a step that is still waiting after
// this long is stuck, and the whole file has to stay well inside the sandbox's
// wall clock even when every guard fires.
const guard = 300 * time.Millisecond

// The helpers below are how these tests refuse to hang. Every mistake in this
// challenge - a nil channel, a missing close, a send with no goroutine behind
// it - shows up as a block rather than as a wrong value, so nothing here is
// allowed to wait longer than that guard for anything.

// produce calls Produce and fails if the call itself blocks. Producing on the
// calling goroutine instead of a new one would never return.
func produce(t *testing.T, n int) <-chan int {
	t.Helper()
	type result struct{ ch <-chan int }
	done := make(chan result, 1)
	go func() { done <- result{Produce(n)} }()
	select {
	case r := <-done:
		return r.ch
	case <-time.After(guard):
		t.Fatalf("Produce(%d) did not return in time - it must not block", n)
		return nil
	}
}

// double is the same guard for Double, which must also return straight away.
func double(t *testing.T, in <-chan int) <-chan int {
	t.Helper()
	type result struct{ ch <-chan int }
	done := make(chan result, 1)
	go func() { done <- result{Double(in)} }()
	select {
	case r := <-done:
		return r.ch
	case <-time.After(guard):
		t.Fatal("Double did not return in time - it must not block")
		return nil
	}
}

// collect is the same guard for Collect, which returns only once the channel
// it was given is closed.
func collect(t *testing.T, ch <-chan int) []int {
	t.Helper()
	done := make(chan []int, 1)
	go func() { done <- Collect(ch) }()
	select {
	case got := <-done:
		return got
	case <-time.After(guard):
		t.Fatal("Collect did not return in time")
		return nil
	}
}

// recv takes one value from ch, failing rather than blocking forever.
func recv(t *testing.T, what string, ch <-chan int) (int, bool) {
	t.Helper()
	select {
	case v, ok := <-ch:
		return v, ok
	case <-time.After(guard):
		t.Fatalf("%s: nothing arrived in time", what)
		return 0, false
	}
}

// send puts one value into ch, failing rather than blocking forever.
func send(t *testing.T, what string, ch chan<- int, v int) {
	t.Helper()
	select {
	case ch <- v:
	case <-time.After(guard):
		t.Fatalf("%s: nothing received %d in time", what, v)
	}
}

func TestProduceReturnsWithoutBlocking(t *testing.T) {
	// Nobody is receiving yet. Produce must hand back the channel and let its
	// own goroutine block on the first send.
	if produce(t, 3) == nil {
		t.Fatal("Produce returned a nil channel - a receive from nil blocks forever")
	}
}

func TestProduceSendsTheSequenceInOrder(t *testing.T) {
	ch := produce(t, 5)
	for want := range 5 {
		got, ok := recv(t, "Produce(5)", ch)
		if !ok {
			t.Fatalf("Produce(5) closed after %d values, want 5", want)
		}
		if got != want {
			t.Fatalf("Produce(5) sent %d, want %d - the values must arrive in order", got, want)
		}
	}
}

func TestProduceClosesWhenDone(t *testing.T) {
	ch := produce(t, 3)
	for range 3 {
		recv(t, "Produce(3)", ch)
	}
	if _, ok := recv(t, "Produce(3) after its last value", ch); ok {
		t.Error("Produce(3) sent a fourth value, want the channel closed")
	}
}

func TestProduceOfNothingIsAnEmptyClosedChannel(t *testing.T) {
	for _, n := range []int{0, -1} {
		if _, ok := recv(t, "Produce", produce(t, n)); ok {
			t.Errorf("Produce(%d) sent a value, want an immediately closed channel", n)
		}
	}
}

func TestCollectReturnsEveryValueInOrder(t *testing.T) {
	ch := make(chan int, 4)
	for _, v := range []int{7, 1, 7, 9} {
		ch <- v
	}
	close(ch)

	got := collect(t, ch)
	if want := []int{7, 1, 7, 9}; !reflect.DeepEqual(got, want) {
		t.Errorf("Collect = %v, want %v", got, want)
	}
}

func TestCollectOfAClosedEmptyChannelIsEmptyNotNil(t *testing.T) {
	ch := make(chan int)
	close(ch)

	got := collect(t, ch)
	if got == nil {
		t.Fatal("Collect returned nil, want an empty non-nil slice")
	}
	if len(got) != 0 {
		t.Errorf("Collect of a closed empty channel = %v, want no values", got)
	}
}

func TestCollectDrainsAProducerAsItSends(t *testing.T) {
	// An unbuffered producer only makes progress while something is receiving,
	// so this passes only if Collect really is a receive loop.
	got := collect(t, produce(t, 4))
	if want := []int{0, 1, 2, 3}; !reflect.DeepEqual(got, want) {
		t.Errorf("Collect(Produce(4)) = %v, want %v", got, want)
	}
}

func TestDoubleDoublesEveryValue(t *testing.T) {
	got := collect(t, double(t, produce(t, 5)))
	if want := []int{0, 2, 4, 6, 8}; !reflect.DeepEqual(got, want) {
		t.Errorf("Collect(Double(Produce(5))) = %v, want %v", got, want)
	}
}

func TestDoubleOfAClosedChannelClosesItsOutput(t *testing.T) {
	in := make(chan int)
	close(in)

	if _, ok := recv(t, "Double of a closed channel", double(t, in)); ok {
		t.Error("Double sent a value for an input that was already closed")
	}
}

func TestDoubleForwardsOneValueAtATime(t *testing.T) {
	// The test sends a value and then insists on the doubled value before it
	// sends the next one. A stage that drained its input into a slice first
	// would still be waiting for a second send while the test waits for a
	// first result, and neither would ever move.
	in := make(chan int)
	out := double(t, in)
	if out == nil {
		t.Fatal("Double returned a nil channel")
	}
	for _, v := range []int{3, 10, -4} {
		send(t, "Double", in, v)
		got, ok := recv(t, "Double", out)
		if !ok {
			t.Fatalf("Double closed its output while %d was still in flight", v)
		}
		if got != v*2 {
			t.Fatalf("Double forwarded %d for input %d, want %d", got, v, v*2)
		}
	}
	close(in)
	if _, ok := recv(t, "Double after its input closed", out); ok {
		t.Error("Double kept its output open after the input closed, want it closed too")
	}
}
