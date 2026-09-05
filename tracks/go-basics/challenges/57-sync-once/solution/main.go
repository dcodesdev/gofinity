package main

import (
	"fmt"
	"sync"
)

// OnceValue returns a function that calls load the first time it is called and
// returns that same value on every later call.
//
// load must not run until the returned function is called for the first time,
// and must never run twice, however many goroutines call at once. Every caller
// gets the value load produced - including the ones that arrived while it was
// still running, which have to wait for it.
func OnceValue(load func() int) func() int {
	var once sync.Once
	var value int
	return func() int {
		// Do returns only after the function has finished, on every goroutine
		// that called it. That is the guarantee that makes reading value here
		// safe without any lock of our own.
		once.Do(func() { value = load() })
		return value
	}
}

// Closer is a one-shot cancellation signal: a channel that is closed at most
// once, and a Close that any number of goroutines may call any number of times.
//
// Its zero value is ready to use, so the channel has to be created lazily -
// and by whichever of Done and Close is reached first, which is itself a job
// for a sync.Once.
type Closer struct {
	// Two Onces, because they answer two different questions: has the channel
	// been made, and has it been closed.
	initOnce  sync.Once
	closeOnce sync.Once
	done      chan struct{}
}

// init creates the channel the first time anything needs it.
func (c *Closer) init() {
	c.initOnce.Do(func() { c.done = make(chan struct{}) })
}

// Done returns a channel that is closed once Close has been called. Before
// that it is open, and never nil: a receive on a nil channel blocks for ever
// even after Close, which would defeat the whole point.
//
// Every call returns the same channel.
func (c *Closer) Done() <-chan struct{} {
	c.init()
	return c.done
}

// Close closes the Done channel. Calling it twice, or from twenty goroutines at
// once, is fine and does nothing the second time - closing an already closed
// channel panics, which is why this needs a Once and not a bool.
func (c *Closer) Close() {
	c.init()
	c.closeOnce.Do(func() { close(c.done) })
}

// IsClosed reports whether Close has been called, without blocking.
func (c *Closer) IsClosed() bool {
	// A receive from a closed channel is always ready, so the default arm is
	// taken exactly while it is still open.
	select {
	case <-c.Done():
		return true
	default:
		return false
	}
}

// entry is one memoized result and the Once that fills it.
type entry struct {
	once  sync.Once
	value int
}

// Memoize returns a function that remembers f's result for each argument: f runs
// at most once per argument, however many goroutines ask for it.
//
// Two different arguments must be able to run f at the same time. That is what
// makes this harder than the cache in the last challenge, where compute ran
// under the one write lock: here the map's lock may only be held while the map
// itself is being touched, and never while f runs.
//
// The shape: a map from the argument to a small entry holding its own
// sync.Once. Take the lock, find or create the entry, release the lock, then do
// the entry's Once outside it.
func Memoize(f func(int) int) func(int) int {
	var mu sync.Mutex
	// Pointers, not values: a sync.Once must never be copied, and a map value
	// is copied every time it is read out.
	entries := map[int]*entry{}
	return func(key int) int {
		mu.Lock()
		e, ok := entries[key]
		if !ok {
			e = &entry{}
			entries[key] = e
		}
		mu.Unlock()

		// Outside the lock. Two goroutines with different keys hold different
		// entries and run f side by side; two with the same key hold the same
		// entry, so the second one waits here and then reads what the first
		// one stored.
		e.once.Do(func() { e.value = f(key) })
		return e.value
	}
}

func main() {
	answer := OnceValue(func() int { fmt.Println("loading"); return 42 })
	fmt.Println(answer(), answer())

	var c Closer
	fmt.Println(c.IsClosed())
	c.Close()
	c.Close()
	<-c.Done()
	fmt.Println(c.IsClosed())

	double := Memoize(func(v int) int { return v * 2 })
	fmt.Println(double(21), double(21))
}
