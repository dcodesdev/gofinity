package main

import (
	"sync"
	"testing"
	"time"
)

// guard is how long any one step may take. Everything here finishes in
// microseconds when it is correct.
const guard = 500 * time.Millisecond

// newBarrier returns a function that blocks until n callers have reached it.
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

// counter is the smallest safe tally the tests need for themselves.
type counter struct {
	mu sync.Mutex
	n  int
}

func (c *counter) add() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n++
}

func (c *counter) value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

// waitFor runs work in a goroutine and fails if it has not finished in time, so
// a missing goroutine or a held lock reports a failure rather than hanging.
func waitFor(t *testing.T, what string, work func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		work()
	}()
	select {
	case <-done:
	case <-time.After(guard):
		t.Fatalf("%s did not finish in time", what)
	}
}

func TestOnceValueIsLazyThenConstant(t *testing.T) {
	var loads counter
	get := OnceValue(func() int {
		loads.add()
		return 42
	})
	if get == nil {
		t.Fatal("OnceValue returned nil")
	}
	if loads.value() != 0 {
		t.Fatalf("load ran %d times before the returned function was called, want 0 - it is lazy", loads.value())
	}
	for i := range 5 {
		if got := get(); got != 42 {
			t.Fatalf("call %d returned %d, want 42", i+1, got)
		}
	}
	if got := loads.value(); got != 1 {
		t.Errorf("load ran %d times over five calls, want exactly 1", got)
	}
}

func TestOnceValueRunsLoadOnceUnderContention(t *testing.T) {
	const goroutines = 32
	var loads counter
	// load blocks until every goroutine has arrived at the returned function,
	// so they all reach it before any of them has a value to read.
	ready := newBarrier(goroutines)
	get := OnceValue(func() int {
		loads.add()
		return 7
	})
	if get == nil {
		t.Fatal("OnceValue returned nil")
	}
	got := make([]int, goroutines)

	waitFor(t, "32 concurrent first calls", func() {
		var wg sync.WaitGroup
		for i := range goroutines {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ready()
				got[i] = get()
			}()
		}
		wg.Wait()
	})

	if n := loads.value(); n != 1 {
		t.Errorf("load ran %d times across %d concurrent callers, want exactly 1", n, goroutines)
	}
	for i, v := range got {
		if v != 7 {
			t.Fatalf("goroutine %d got %d, want 7 - a caller returned before load had finished", i, v)
		}
	}
}

func TestOnceValueKeepsEachFunctionSeparate(t *testing.T) {
	a := OnceValue(func() int { return 1 })
	b := OnceValue(func() int { return 2 })
	if a == nil || b == nil {
		t.Fatal("OnceValue returned nil")
	}
	if a() != 1 || b() != 2 || a() != 1 {
		t.Errorf("two OnceValue functions gave %d and %d, want 1 and 2 - they must not share state", a(), b())
	}
}

func TestCloserZeroValueDoneIsOpenAndNotNil(t *testing.T) {
	var c Closer
	done := c.Done()
	if done == nil {
		t.Fatal("Done() on a fresh Closer returned nil - a receive on nil blocks for ever")
	}
	if c.IsClosed() {
		t.Error("IsClosed() on a fresh Closer = true, want false")
	}
	select {
	case <-done:
		t.Error("Done() was already closed before Close was called")
	default:
	}
}

func TestCloserDoneIsAlwaysTheSameChannel(t *testing.T) {
	var c Closer
	if c.Done() != c.Done() {
		t.Error("Done() returned two different channels, want the same one every time")
	}
	c.Close()
	if c.Done() != c.Done() {
		t.Error("Done() returned two different channels after Close, want the same one every time")
	}
}

func TestCloserCloseClosesDone(t *testing.T) {
	var c Closer
	done := c.Done()
	c.Close()
	select {
	case _, ok := <-done:
		if ok {
			t.Error("Done() delivered a value, want a closed channel")
		}
	case <-time.After(guard):
		t.Fatal("Done() was still open after Close")
	}
	if !c.IsClosed() {
		t.Error("IsClosed() = false after Close, want true")
	}
	// The channel handed out before Close is the one that closes, so a
	// goroutine already waiting on it is woken.
	if c.Done() != done {
		t.Error("Done() after Close returned a different channel from the one handed out before")
	}
}

func TestCloserCloseIsIdempotent(t *testing.T) {
	// A second close of the same channel panics, so this test either passes or
	// takes the whole run down with it.
	var c Closer
	c.Close()
	c.Close()
	c.Close()
	if !c.IsClosed() {
		t.Error("IsClosed() = false after three Closes, want true")
	}
}

func TestCloserSurvivesManyClosersAtOnce(t *testing.T) {
	var c Closer
	const goroutines = 24
	ready := newBarrier(goroutines)
	waitFor(t, "24 concurrent Closes", func() {
		var wg sync.WaitGroup
		for range goroutines {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ready()
				c.Close()
			}()
		}
		wg.Wait()
	})
	if !c.IsClosed() {
		t.Error("IsClosed() = false after many concurrent Closes, want true")
	}
}

func TestCloserWakesEveryWaiter(t *testing.T) {
	var c Closer
	const waiters = 8
	woken := make(chan int, waiters)
	for i := range waiters {
		go func() {
			<-c.Done()
			woken <- i
		}()
	}
	// Nothing has been closed yet, so nobody should be through.
	time.Sleep(10 * time.Millisecond)
	if len(woken) != 0 {
		t.Fatalf("%d waiters got past Done() before Close", len(woken))
	}
	c.Close()
	for range waiters {
		select {
		case <-woken:
		case <-time.After(guard):
			t.Fatal("not every waiter was woken by one Close - closing must reach all of them")
		}
	}
}

func TestMemoizeCallsFOncePerArgument(t *testing.T) {
	var calls counter
	double := Memoize(func(v int) int {
		calls.add()
		return v * 2
	})
	if double == nil {
		t.Fatal("Memoize returned nil")
	}
	for range 3 {
		if got := double(21); got != 42 {
			t.Fatalf("double(21) = %d, want 42", got)
		}
		if got := double(-4); got != -8 {
			t.Fatalf("double(-4) = %d, want -8", got)
		}
	}
	if got := calls.value(); got != 2 {
		t.Errorf("f ran %d times for two distinct arguments asked six times, want 2", got)
	}
}

func TestMemoizeCallsFOnceForOneArgumentUnderContention(t *testing.T) {
	const goroutines = 32
	var calls counter
	ready := newBarrier(goroutines)
	square := Memoize(func(v int) int {
		calls.add()
		return v * v
	})
	if square == nil {
		t.Fatal("Memoize returned nil")
	}
	got := make([]int, goroutines)

	waitFor(t, "32 concurrent calls with the same argument", func() {
		var wg sync.WaitGroup
		for i := range goroutines {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ready()
				got[i] = square(9)
			}()
		}
		wg.Wait()
	})

	if n := calls.value(); n != 1 {
		t.Errorf("f ran %d times for one argument asked by %d goroutines at once, want exactly 1", n, goroutines)
	}
	for i, v := range got {
		if v != 81 {
			t.Fatalf("goroutine %d got %d, want 81", i, v)
		}
	}
}

func TestMemoizeRunsDifferentArgumentsAtTheSameTime(t *testing.T) {
	const keys = 4
	// Each call blocks inside f until all four are inside. That can only
	// happen if the map's lock is released before f runs.
	inside := newBarrier(keys)
	triple := Memoize(func(v int) int {
		inside()
		return v * 3
	})
	if triple == nil {
		t.Fatal("Memoize returned nil")
	}

	waitFor(t, "four different arguments computed at once", func() {
		var wg sync.WaitGroup
		for k := range keys {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if got := triple(k); got != k*3 {
					t.Errorf("triple(%d) = %d, want %d", k, got, k*3)
				}
			}()
		}
		wg.Wait()
	})
}

func TestMemoizeKeepsEachFunctionSeparate(t *testing.T) {
	up := Memoize(func(v int) int { return v + 1 })
	down := Memoize(func(v int) int { return v - 1 })
	if up == nil || down == nil {
		t.Fatal("Memoize returned nil")
	}
	if up(10) != 11 || down(10) != 9 || up(10) != 11 {
		t.Errorf("two memoized functions gave %d and %d for 10, want 11 and 9 - they must not share a cache", up(10), down(10))
	}
}
