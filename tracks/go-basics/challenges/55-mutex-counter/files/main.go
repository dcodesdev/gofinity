package main

import "fmt"

// Counter is a whole number that many goroutines may add to at once.
//
// Its zero value is ready to use: var c Counter works, and so does new(Counter).
// Every method has a pointer receiver, because a copy of a Counter would be a
// copy of the mutex - and a copy of a mutex guards nothing.
type Counter struct {
	// TODO: the number, and the thing that guards it.
}

// Add increases the counter by delta. delta may be negative.
func (c *Counter) Add(delta int) {
	// TODO
}

// Value returns the current total.
//
// Reading is not free of the lock: an int is not written atomically, and a read
// racing a write is undefined behaviour rather than merely a stale answer.
func (c *Counter) Value() int {
	// TODO
	return 0
}

// Tally counts how often each key has been recorded.
//
// Its zero value is ready to use too, so the map has to be created on the first
// Record rather than by a constructor - under the lock, like everything else
// that touches it.
type Tally struct {
	// TODO
}

// Record adds one to the count for key.
func (t *Tally) Record(key string) {
	// TODO
}

// Count returns the count for key, or 0 if it has never been recorded.
func (t *Tally) Count(key string) int {
	// TODO
	return 0
}

// Snapshot returns a copy of every count taken at one moment.
//
// It must be a copy. Returning the internal map would hand the caller a
// reference that outlives the lock, and every later Record would be a race the
// caller cannot see. A Tally that has recorded nothing snapshots to an empty,
// non-nil map.
func (t *Tally) Snapshot() map[string]int {
	// TODO
	return nil
}

// CountMatching reports how many of values satisfy pred, calling pred for every
// value in its own goroutine and adding up the results with a Counter.
//
// pred may be slow, and the goroutines must all be running before any of them
// finishes, so start them all first and wait afterwards.
func CountMatching(values []int, pred func(int) bool) int {
	// TODO
	return 0
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
