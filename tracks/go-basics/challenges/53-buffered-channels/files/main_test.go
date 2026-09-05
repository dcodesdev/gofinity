package main

import (
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"
)

// guard is how long any one step may take. Everything here finishes in
// microseconds when it is correct, so a step still waiting after this long is
// stuck - and the whole file stays well inside the sandbox's wall clock even
// when every guard fires.
const guard = 300 * time.Millisecond

// fillUpTo calls FillUpTo and fails rather than blocking forever: sending one
// value too many into a full channel would never return.
func fillUpTo(t *testing.T, ch chan int, values []int) int {
	t.Helper()
	done := make(chan int, 1)
	go func() { done <- FillUpTo(ch, values) }()
	select {
	case n := <-done:
		return n
	case <-time.After(guard):
		t.Fatal("FillUpTo did not return in time - it must stop once the channel is full")
		return 0
	}
}

// mapLimited is the same guard for MapLimited.
func mapLimited(t *testing.T, in []int, limit int, f func(int) int) []int {
	t.Helper()
	done := make(chan []int, 1)
	go func() { done <- MapLimited(in, limit, f) }()
	select {
	case got := <-done:
		return got
	case <-time.After(guard):
		t.Fatalf("MapLimited with limit %d did not return in time", limit)
		return nil
	}
}

// drain receives everything left in ch and reports whether it ended closed.
func drain(t *testing.T, ch <-chan int) ([]int, bool) {
	t.Helper()
	got := []int{}
	for {
		select {
		case v, ok := <-ch:
			if !ok {
				return got, true
			}
			got = append(got, v)
		case <-time.After(guard):
			t.Fatalf("the channel neither sent nor closed in time after %v", got)
			return got, false
		}
	}
}

// tracker records how many calls were inside f at the same time.
type tracker struct {
	mu       sync.Mutex
	now, max int
}

func (t *tracker) enter() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.now++
	if t.now > t.max {
		t.max = t.now
	}
}

func (t *tracker) leave() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.now--
}

func (t *tracker) peak() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.max
}

// newBarrier returns a function that blocks until n callers have reached it.
// It is the concurrency proof: with n set to the limit, the work can only
// finish if the implementation really does run that many calls at once.
func newBarrier(n int) func() {
	var mu sync.Mutex
	count := 0
	open := make(chan struct{})
	return func() {
		mu.Lock()
		count++
		if count == n {
			close(open)
		}
		mu.Unlock()
		<-open
	}
}

func TestBufferedHoldsEveryValueAlready(t *testing.T) {
	values := []int{4, 8, 15, 16}
	ch := Buffered(values)
	if ch == nil {
		t.Fatal("Buffered returned a nil channel")
	}
	if cap(ch) != len(values) {
		t.Errorf("cap(Buffered(%v)) = %d, want %d", values, cap(ch), len(values))
	}
	// The values are in the buffer before Buffered returns, so no goroutine is
	// needed and none of the sends can have blocked.
	if len(ch) != len(values) {
		t.Errorf("len(Buffered(%v)) = %d, want %d - every value should already be buffered", values, len(ch), len(values))
	}

	got, closed := drain(t, ch)
	if !reflect.DeepEqual(got, values) {
		t.Errorf("draining Buffered(%v) gave %v", values, got)
	}
	if !closed {
		t.Error("Buffered's channel was still open after its last value, want it closed")
	}
}

func TestBufferedOfNothingIsAClosedEmptyChannel(t *testing.T) {
	ch := Buffered(nil)
	if cap(ch) != 0 {
		t.Errorf("cap(Buffered(nil)) = %d, want 0", cap(ch))
	}
	if got, closed := drain(t, ch); !closed || len(got) != 0 {
		t.Errorf("draining Buffered(nil) gave %v, closed=%v, want no values and closed", got, closed)
	}
}

func TestFillUpToStopsWhenTheBufferIsFull(t *testing.T) {
	ch := make(chan int, 3)
	if got := fillUpTo(t, ch, []int{1, 2, 3, 4, 5}); got != 3 {
		t.Errorf("FillUpTo into a channel of capacity 3 = %d, want 3", got)
	}
	if len(ch) != 3 {
		t.Errorf("len(ch) = %d after filling a channel of capacity 3, want 3", len(ch))
	}
	close(ch)
	if got, _ := drain(t, ch); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Errorf("the channel holds %v, want the first three values in order", got)
	}
}

func TestFillUpToStopsWhenTheValuesRunOut(t *testing.T) {
	ch := make(chan int, 5)
	if got := fillUpTo(t, ch, []int{9, 9}); got != 2 {
		t.Errorf("FillUpTo with two values into a channel of capacity 5 = %d, want 2", got)
	}
	if got := fillUpTo(t, ch, nil); got != 0 {
		t.Errorf("FillUpTo with no values = %d, want 0", got)
	}
}

func TestFillUpToCountsOnlyTheRemainingRoom(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 100
	if got := fillUpTo(t, ch, []int{1, 2, 3}); got != 2 {
		t.Errorf("FillUpTo into a channel of capacity 3 already holding one value = %d, want 2", got)
	}
}

func TestFillUpToSendsNothingIntoAnUnbufferedChannel(t *testing.T) {
	// cap is 0, so it is full before it starts. A send here would block for
	// ever: nobody is receiving.
	if got := fillUpTo(t, make(chan int), []int{1, 2}); got != 0 {
		t.Errorf("FillUpTo into an unbuffered channel = %d, want 0", got)
	}
}

func TestMapLimitedReturnsResultsInInputOrder(t *testing.T) {
	in := []int{1, 2, 3, 4, 5, 6, 7, 8}
	got := mapLimited(t, in, 3, func(v int) int { return v * v })
	want := []int{1, 4, 9, 16, 25, 36, 49, 64}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("MapLimited squares = %v, want %v - results follow the input order, not the finishing order", got, want)
	}
}

func TestMapLimitedOfNothingIsEmptyNotNil(t *testing.T) {
	got := mapLimited(t, nil, 4, func(v int) int { return v })
	if got == nil {
		t.Fatal("MapLimited of no input returned nil, want an empty non-nil slice")
	}
	if len(got) != 0 {
		t.Errorf("MapLimited of no input = %v, want no results", got)
	}
}

func TestMapLimitedRunsExactlyLimitCallsAtOnce(t *testing.T) {
	const limit = 3
	in := []int{1, 2, 3, 4, 5, 6, 7, 8}
	var seen tracker
	// Every call waits until limit of them are inside f, so a MapLimited that
	// ran them one after another can never finish. The tracker then proves it
	// never ran more than limit at a time either.
	wait := newBarrier(limit)

	got := mapLimited(t, in, limit, func(v int) int {
		seen.enter()
		wait()
		defer seen.leave()
		return v + 1
	})

	if want := []int{2, 3, 4, 5, 6, 7, 8, 9}; !reflect.DeepEqual(got, want) {
		t.Errorf("MapLimited = %v, want %v", got, want)
	}
	if peak := seen.peak(); peak != limit {
		t.Errorf("at most %d calls were inside f at once, want exactly %d", peak, limit)
	}
}

func TestMapLimitedAboveTheInputLengthRunsEverythingAtOnce(t *testing.T) {
	in := []int{1, 2, 3, 4}
	var seen tracker
	wait := newBarrier(len(in))

	mapLimited(t, in, 100, func(v int) int {
		seen.enter()
		wait()
		defer seen.leave()
		return v
	})

	if peak := seen.peak(); peak != len(in) {
		t.Errorf("at most %d calls ran at once with a limit above the input length, want %d", peak, len(in))
	}
}

func TestMapLimitedTreatsANonPositiveLimitAsOne(t *testing.T) {
	for _, limit := range []int{0, -5} {
		in := []int{1, 2, 3, 4, 5, 6, 7, 8}
		var seen tracker

		got := mapLimited(t, in, limit, func(v int) int {
			seen.enter()
			// Give the scheduler every chance to overlap this call with
			// another one, so an unlimited implementation is caught.
			runtime.Gosched()
			defer seen.leave()
			return v * 10
		})

		if want := []int{10, 20, 30, 40, 50, 60, 70, 80}; !reflect.DeepEqual(got, want) {
			t.Errorf("MapLimited with limit %d = %v, want %v", limit, got, want)
		}
		if peak := seen.peak(); peak != 1 {
			t.Errorf("a limit of %d let %d calls run at once, want 1", limit, peak)
		}
	}
}
