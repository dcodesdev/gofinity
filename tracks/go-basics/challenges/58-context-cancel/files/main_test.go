package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

// guard is how long any one step may take. Nothing here waits for time to
// pass: every result is triggered by a cancel or a send, so a step still
// running after this long is stuck rather than slow.
const guard = 500 * time.Millisecond

// waitErr calls Wait in its own goroutine so a version that never returns
// fails the test instead of hanging the run.
func waitErr(t *testing.T, ctx context.Context, ready <-chan struct{}) error {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- Wait(ctx, ready) }()
	select {
	case err := <-done:
		return err
	case <-time.After(guard):
		t.Fatal("Wait did not return - it must select on ctx.Done() as well as ready")
		return nil
	}
}

// recv takes one value from ch, reporting whether the channel was still open.
func recv(t *testing.T, ch <-chan int) (int, bool) {
	t.Helper()
	select {
	case v, open := <-ch:
		return v, open
	case <-time.After(guard):
		t.Fatal("nothing arrived on the channel and it was never closed")
		return 0, false
	}
}

// send hands v to CountTicks, failing rather than blocking for ever when
// nothing is receiving.
func send(t *testing.T, ticks chan int, v int) {
	t.Helper()
	select {
	case ticks <- v:
	case <-time.After(guard):
		t.Fatalf("nothing received %d - CountTicks must keep receiving until the channel closes or the context is done", v)
	}
}

// expired returns a context whose deadline is already in the past, so no test
// has to wait for one to pass.
func expired(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	t.Cleanup(cancel)
	return ctx
}

func TestWaitReturnsNilWhenReadyFires(t *testing.T) {
	ready := make(chan struct{})
	close(ready)
	if err := waitErr(t, context.Background(), ready); err != nil {
		t.Errorf("Wait with a ready channel = %v, want nil", err)
	}
}

func TestWaitReturnsNilWhenReadyFiresLater(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ready := make(chan struct{})
	go func() { ready <- struct{}{} }()
	if err := waitErr(t, ctx, ready); err != nil {
		t.Errorf("Wait = %v when ready received a value, want nil", err)
	}
}

func TestWaitReturnsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Nothing will ever arrive on ready, so the only way out is ctx.Done().
	done := make(chan error, 1)
	go func() { done <- Wait(ctx, make(chan struct{})) }()
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Wait after cancel() = %v, want context.Canceled - return ctx.Err()", err)
		}
	case <-time.After(guard):
		t.Fatal("Wait did not return after the context was cancelled")
	}
}

func TestWaitReturnsDeadlineExceeded(t *testing.T) {
	err := waitErr(t, expired(t), make(chan struct{}))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Wait on a context past its deadline = %v, want context.DeadlineExceeded", err)
	}
	if errors.Is(err, context.Canceled) {
		t.Error("Wait reported context.Canceled for an expired deadline - ctx.Err() tells the two apart, a hand-written error cannot")
	}
}

func TestCancelledIsFalseWhileTheContextIsLive(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan bool, 1)
	go func() { done <- Cancelled(ctx) }()
	select {
	case got := <-done:
		if got {
			t.Error("Cancelled(ctx) = true for a live context, want false")
		}
	case <-time.After(guard):
		t.Fatal("Cancelled blocked on a live context - a select needs a default to be non-blocking")
	}
}

func TestCancelledIsTrueAfterCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if !Cancelled(ctx) {
		t.Error("Cancelled(ctx) = false after cancel(), want true")
	}
	if !Cancelled(expired(t)) {
		t.Error("Cancelled(ctx) = false for a context past its deadline, want true")
	}
}

func TestCountTicksStopsWhenTheChannelCloses(t *testing.T) {
	ticks := make(chan int, 3)
	ticks <- 1
	ticks <- 2
	ticks <- 3
	close(ticks)

	got := make(chan int, 1)
	go func() { got <- CountTicks(context.Background(), ticks) }()
	select {
	case n := <-got:
		if n != 3 {
			t.Errorf("CountTicks over three values on a closed channel = %d, want 3", n)
		}
	case <-time.After(guard):
		t.Fatal("CountTicks did not return on a closed channel - a closed channel receives for ever, so check the second value")
	}
}

func TestCountTicksStopsWhenTheContextIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ticks := make(chan int)

	got := make(chan int, 1)
	go func() { got <- CountTicks(ctx, ticks) }()

	// The channel is unbuffered, so each send only completes once CountTicks
	// has received it. Two sends means two values counted, and nothing else
	// can be in flight when we cancel.
	send(t, ticks, 10)
	send(t, ticks, 20)
	cancel()

	select {
	case n := <-got:
		if n != 2 {
			t.Errorf("CountTicks = %d after two values and a cancel, want 2", n)
		}
	case <-time.After(guard):
		t.Fatal("CountTicks did not return after the context was cancelled")
	}
}

func TestCountTicksOnACancelledContextCountsNothing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got := make(chan int, 1)
	// A nil channel never receives, so only ctx.Done() can end this.
	go func() { got <- CountTicks(ctx, nil) }()
	select {
	case n := <-got:
		if n != 0 {
			t.Errorf("CountTicks on an already-cancelled context = %d, want 0", n)
		}
	case <-time.After(guard):
		t.Fatal("CountTicks did not return for an already-cancelled context")
	}
}

func TestProduceDeliversEveryValueInOrderAndCloses(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var got []int
	ch := Produce(ctx, []int{3, 1, 4, 1, 5})
	for {
		v, open := recv(t, ch)
		if !open {
			break
		}
		got = append(got, v)
	}

	want := []int{3, 1, 4, 1, 5}
	if len(got) != len(want) {
		t.Fatalf("Produce delivered %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Produce delivered %v, want %v - values must keep their order", got, want)
		}
	}
}

func TestProduceOfNothingClosesStraightAway(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if v, open := recv(t, Produce(ctx, nil)); open {
		t.Errorf("Produce(ctx, nil) delivered %d, want a channel that is closed and empty", v)
	}
}

func TestProduceReturnsBeforeItsValuesAreConsumed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// The channel is unbuffered and nobody is reading yet, so a Produce that
	// sent its values before returning could never get here.
	done := make(chan (<-chan int), 1)
	go func() { done <- Produce(ctx, []int{1, 2, 3}) }()
	select {
	case ch := <-done:
		if v, open := recv(t, ch); !open || v != 1 {
			t.Errorf("first value = %d, open = %v, want 1 and an open channel", v, open)
		}
	case <-time.After(guard):
		t.Fatal("Produce did not return until its values were read - the sends belong in a goroutine")
	}
}

func TestProduceStopsAndClosesWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := Produce(ctx, []int{1, 2, 3, 4, 5})

	if v, open := recv(t, ch); !open || v != 1 {
		t.Fatalf("first value = %d, open = %v, want 1 and an open channel", v, open)
	}

	// Nobody is receiving now, so the producer is parked on its next send.
	// Cancelling is the only thing that can free it, and it must close the
	// channel on the way out.
	cancel()
	if v, open := recv(t, ch); open {
		t.Errorf("Produce delivered %d after the context was cancelled, want the channel closed", v)
	}
}

func TestProduceOnACancelledContextStillCloses(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Both cases of the producer's select are ready here, so it may deliver a
	// value or two before it notices. What it may never do is block: the
	// channel has to end up closed either way.
	ch := Produce(ctx, []int{1, 2, 3})
	for i := 0; i <= 3; i++ {
		if _, open := recv(t, ch); !open {
			return
		}
	}
	t.Error("Produce delivered every value for an already-cancelled context and never closed the channel")
}
