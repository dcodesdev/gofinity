package main

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

// guard is how long any one step may take. Everything here finishes in
// milliseconds when it is correct, so a step still waiting after this long is
// stuck rather than slow.
const guard = 500 * time.Millisecond

// newBarrier returns a function that blocks until n callers have reached it.
// It is the concurrency proof: work that has to pass a barrier of n can only
// finish if n calls really are running at the same time.
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

// countMatching calls CountMatching and fails rather than blocking for ever, so
// an implementation that never starts its goroutines reports a failure instead
// of hanging the whole run.
func countMatching(t *testing.T, values []int, pred func(int) bool) int {
	t.Helper()
	done := make(chan int, 1)
	go func() { done <- CountMatching(values, pred) }()
	select {
	case n := <-done:
		return n
	case <-time.After(guard):
		t.Fatal("CountMatching did not return in time - every value's pred must run in its own goroutine")
		return 0
	}
}

func TestCounterZeroValueIsReadyToUse(t *testing.T) {
	var c Counter
	if got := c.Value(); got != 0 {
		t.Errorf("a fresh Counter has Value() = %d, want 0", got)
	}
	c.Add(3)
	c.Add(4)
	if got := c.Value(); got != 7 {
		t.Errorf("after Add(3) and Add(4), Value() = %d, want 7", got)
	}
}

func TestCounterAddTakesNegativeDeltas(t *testing.T) {
	c := new(Counter)
	c.Add(10)
	c.Add(-4)
	c.Add(-6)
	if got := c.Value(); got != 0 {
		t.Errorf("Value() = %d after 10, -4 and -6, want 0", got)
	}
}

func TestCounterLosesNoUpdatesUnderContention(t *testing.T) {
	const goroutines, each = 50, 2000
	var c Counter
	var wg sync.WaitGroup
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range each {
				c.Add(1)
			}
		}()
	}
	wg.Wait()

	// An unguarded n++ reads, adds and writes back as three steps, so two
	// goroutines can read the same value and one of the increments vanishes.
	// At this many increments that is not a maybe.
	if want := goroutines * each; c.Value() != want {
		t.Errorf("Value() = %d after %d increments from %d goroutines, want %d - increments were lost",
			c.Value(), want, goroutines, want)
	}
}

func TestTallyZeroValueRecordsAndCounts(t *testing.T) {
	var tally Tally
	if got := tally.Count("missing"); got != 0 {
		t.Errorf("Count of a key never recorded = %d, want 0", got)
	}
	tally.Record("go")
	tally.Record("go")
	tally.Record("rust")
	if got := tally.Count("go"); got != 2 {
		t.Errorf(`Count("go") = %d after two Records, want 2`, got)
	}
	if got := tally.Count("rust"); got != 1 {
		t.Errorf(`Count("rust") = %d after one Record, want 1`, got)
	}
}

func TestTallySnapshotOfNothingIsEmptyNotNil(t *testing.T) {
	var tally Tally
	got := tally.Snapshot()
	if got == nil {
		t.Fatal("Snapshot of an empty Tally returned nil, want an empty non-nil map")
	}
	if len(got) != 0 {
		t.Errorf("Snapshot of an empty Tally = %v, want no entries", got)
	}
}

func TestTallySnapshotIsACopy(t *testing.T) {
	var tally Tally
	tally.Record("go")
	snap := tally.Snapshot()
	if want := map[string]int{"go": 1}; !reflect.DeepEqual(snap, want) {
		t.Fatalf("Snapshot() = %v, want %v", snap, want)
	}

	// Writing to the snapshot must not reach the Tally, and later Records must
	// not reach the snapshot. Either one failing means the internal map was
	// handed out rather than copied.
	snap["go"] = 99
	snap["new"] = 1
	tally.Record("go")

	if got := tally.Count("go"); got != 2 {
		t.Errorf(`Count("go") = %d after writing to a snapshot and recording once more, want 2`, got)
	}
	if got := tally.Count("new"); got != 0 {
		t.Errorf(`Count("new") = %d - a key added to the snapshot reached the Tally`, got)
	}
	if snap["go"] != 99 || len(snap) != 2 {
		t.Errorf("the snapshot changed to %v when the Tally was recorded to again, want a value fixed at the moment it was taken", snap)
	}
}

func TestTallyRecordsFromManyGoroutines(t *testing.T) {
	const goroutines, each = 40, 250
	keys := []string{"a", "b", "c", "d"}
	var tally Tally
	var wg sync.WaitGroup
	for g := range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range each {
				tally.Record(keys[(g+i)%len(keys)])
			}
		}()
	}
	wg.Wait()

	want := map[string]int{}
	for _, k := range keys {
		want[k] = goroutines * each / len(keys)
	}
	if got := tally.Snapshot(); !reflect.DeepEqual(got, want) {
		t.Errorf("Snapshot() = %v after %d concurrent Records, want %v", got, goroutines*each, want)
	}
}

func TestCountMatchingCountsTheMatches(t *testing.T) {
	values := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	if got := countMatching(t, values, func(v int) bool { return v%3 == 0 }); got != 3 {
		t.Errorf("CountMatching(%v, multiple of 3) = %d, want 3", values, got)
	}
	if got := countMatching(t, values, func(v int) bool { return false }); got != 0 {
		t.Errorf("CountMatching with a pred that never matches = %d, want 0", got)
	}
	if got := countMatching(t, values, func(v int) bool { return true }); got != len(values) {
		t.Errorf("CountMatching with a pred that always matches = %d, want %d", got, len(values))
	}
}

func TestCountMatchingOfNothingIsZero(t *testing.T) {
	if got := countMatching(t, nil, func(v int) bool { return true }); got != 0 {
		t.Errorf("CountMatching(nil, ...) = %d, want 0", got)
	}
}

func TestCountMatchingCallsPredForEveryValue(t *testing.T) {
	values := []int{5, 5, 5, 5, 5, 5}
	var seen Counter
	got := countMatching(t, values, func(v int) bool {
		seen.Add(1)
		return v == 5
	})
	if seen.Value() != len(values) {
		t.Errorf("pred ran %d times for %d values, want one call each - duplicates are not deduplicated", seen.Value(), len(values))
	}
	if got != len(values) {
		t.Errorf("CountMatching = %d, want %d", got, len(values))
	}
}

func TestCountMatchingRunsEveryPredAtOnce(t *testing.T) {
	values := []int{1, 2, 3, 4, 5, 6, 7, 8}
	// Every call waits until all of them have arrived, so an implementation
	// that ran the values one after another can never finish.
	wait := newBarrier(len(values))
	got := countMatching(t, values, func(v int) bool {
		wait()
		return v%2 == 0
	})
	if got != len(values)/2 {
		t.Errorf("CountMatching over %v = %d, want %d", values, got, len(values)/2)
	}
}

func TestCountMatchingIsCorrectUnderContention(t *testing.T) {
	values := make([]int, 5000)
	for i := range values {
		values[i] = i
	}
	// Five thousand goroutines all adding to the same Counter: with the lock
	// this is exact, without it the answer is short.
	if got := countMatching(t, values, func(v int) bool { return true }); got != len(values) {
		t.Errorf("CountMatching over %d values = %d, want %d - matches were lost", len(values), got, len(values))
	}
}

func TestCounterAndTallyAreUsedThroughPointers(t *testing.T) {
	// A method on a value receiver would copy the mutex, so this must not
	// compile as a value method set. Taking the address of a local and calling
	// through it is what the rest of the tests do; this only documents it.
	c := &Counter{}
	c.Add(1)
	tally := &Tally{}
	tally.Record("x")
	if fmt.Sprint(c.Value(), tally.Count("x")) != "1 1" {
		t.Errorf("Counter and Tally through pointers gave %d and %d, want 1 and 1", c.Value(), tally.Count("x"))
	}
}
