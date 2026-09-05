package main

import (
	"fmt"
	"sync"
)

// Counter is a whole number that many goroutines may add to at once.
//
// Its zero value is ready to use: var c Counter works, and so does new(Counter).
// Every method has a pointer receiver, because a copy of a Counter would be a
// copy of the mutex - and a copy of a mutex guards nothing.
type Counter struct {
	// The mutex sits directly above the fields it protects. That is the
	// convention, and it is the only documentation most Go code gives.
	mu sync.Mutex
	n  int
}

// Add increases the counter by delta. delta may be negative.
func (c *Counter) Add(delta int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n += delta
}

// Value returns the current total.
//
// Reading is not free of the lock: an int is not written atomically, and a read
// racing a write is undefined behaviour rather than merely a stale answer.
func (c *Counter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

// Tally counts how often each key has been recorded.
//
// Its zero value is ready to use too, so the map has to be created on the first
// Record rather than by a constructor - under the lock, like everything else
// that touches it.
type Tally struct {
	mu     sync.Mutex
	counts map[string]int
}

// Record adds one to the count for key.
func (t *Tally) Record(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.counts == nil {
		t.counts = map[string]int{}
	}
	t.counts[key]++
}

// Count returns the count for key, or 0 if it has never been recorded.
func (t *Tally) Count(key string) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	// Reading a nil map is legal and gives the zero value, so there is nothing
	// to create here.
	return t.counts[key]
}

// Snapshot returns a copy of every count taken at one moment.
//
// It must be a copy. Returning the internal map would hand the caller a
// reference that outlives the lock, and every later Record would be a race the
// caller cannot see. A Tally that has recorded nothing snapshots to an empty,
// non-nil map.
func (t *Tally) Snapshot() map[string]int {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make(map[string]int, len(t.counts))
	for k, v := range t.counts {
		out[k] = v
	}
	return out
}

// CountMatching reports how many of values satisfy pred, calling pred for every
// value in its own goroutine and adding up the results with a Counter.
//
// pred may be slow, and the goroutines must all be running before any of them
// finishes, so start them all first and wait afterwards.
func CountMatching(values []int, pred func(int) bool) int {
	var matched Counter
	var wg sync.WaitGroup
	for _, v := range values {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// pred runs outside the lock. Holding a mutex across a caller's
			// function is how a fast program becomes a serial one.
			if pred(v) {
				matched.Add(1)
			}
		}()
	}
	wg.Wait()
	return matched.Value()
}

func main() {
	var c Counter
	c.Add(2)
	c.Add(-1)
	fmt.Println(c.Value())

	var t Tally
	t.Record("go")
	t.Record("go")
	t.Record("rust")
	fmt.Println(t.Count("go"), t.Snapshot())

	fmt.Println(CountMatching([]int{1, 2, 3, 4, 5, 6}, func(v int) bool { return v%2 == 0 }))
}
