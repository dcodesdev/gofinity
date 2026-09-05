package main

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

// guard is how long any one step may take. Everything here is microseconds of
// work when it is correct, so a step still waiting after this long is blocked.
const guard = 500 * time.Millisecond

// newBarrier returns a function that blocks until n callers have reached it.
// With n readers it proves that read locks really are shared: under a plain
// Mutex the second reader never arrives.
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

// waitFor runs work in a goroutine and fails if it has not finished in time,
// so a lock held too long reports a failure rather than hanging the run.
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

func TestZeroValueCacheIsEmptyAndUsable(t *testing.T) {
	var c Cache
	if _, ok := c.Get("nothing"); ok {
		t.Error("Get on a fresh cache reported a hit, want a miss")
	}
	if got := c.Len(); got != 0 {
		t.Errorf("Len() of a fresh cache = %d, want 0", got)
	}
	if got := c.Keys(); got == nil {
		t.Error("Keys() of a fresh cache returned nil, want an empty non-nil slice")
	} else if len(got) != 0 {
		t.Errorf("Keys() of a fresh cache = %v, want no keys", got)
	}
	c.Set("a", 1)
	if v, ok := c.Get("a"); !ok || v != 1 {
		t.Errorf(`Get("a") = %d, %v after Set("a", 1), want 1, true`, v, ok)
	}
}

func TestSetReplacesAndKeysAreSorted(t *testing.T) {
	var c Cache
	c.Set("pear", 1)
	c.Set("apple", 2)
	c.Set("pear", 3)
	if v, _ := c.Get("pear"); v != 3 {
		t.Errorf(`Get("pear") = %d after Set("pear", 3), want 3`, v)
	}
	if got := c.Len(); got != 2 {
		t.Errorf("Len() = %d after setting two distinct keys, want 2", got)
	}
	if got, want := c.Keys(), []string{"apple", "pear"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Keys() = %v, want %v - sorted", got, want)
	}
}

func TestForEachVisitsEveryEntryOnce(t *testing.T) {
	var c Cache
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	seen := map[string]int{}
	c.ForEach(func(key string, value int) { seen[key] += value })
	if want := map[string]int{"a": 1, "b": 2, "c": 3}; !reflect.DeepEqual(seen, want) {
		t.Errorf("ForEach visited %v, want %v exactly once each", seen, want)
	}

	var empty Cache
	calls := 0
	empty.ForEach(func(string, int) { calls++ })
	if calls != 0 {
		t.Errorf("ForEach on an empty cache called f %d times, want 0", calls)
	}
}

func TestManyForEachesRunAtTheSameTime(t *testing.T) {
	var c Cache
	c.Set("a", 1)

	const readers = 3
	// Each walk blocks inside f until all three walks have arrived. Three read
	// locks can be held at once; three write locks cannot, so a Lock here
	// never gets past the first reader.
	wait := newBarrier(readers)
	waitFor(t, fmt.Sprintf("%d concurrent ForEach calls", readers), func() {
		var wg sync.WaitGroup
		for range readers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				c.ForEach(func(string, int) { wait() })
			}()
		}
		wg.Wait()
	})
}

func TestManyGetsRunAtTheSameTime(t *testing.T) {
	var c Cache
	for i := range 200 {
		c.Set(fmt.Sprint(i), i)
	}
	// Not a lock proof, an answer proof: hammering Get from many goroutines
	// while nobody writes must return the stored value every time.
	waitFor(t, "concurrent Gets", func() {
		var wg sync.WaitGroup
		for g := range 20 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range 200 {
					if v, ok := c.Get(fmt.Sprint(i)); !ok || v != i {
						t.Errorf("goroutine %d: Get(%d) = %d, %v, want %d, true", g, i, v, ok, i)
						return
					}
				}
			}()
		}
		wg.Wait()
	})
}

func TestGetOrComputeComputesOnlyOnAMiss(t *testing.T) {
	var c Cache
	v, computed := c.GetOrCompute("a", func() int { return 7 })
	if v != 7 || !computed {
		t.Errorf(`GetOrCompute("a") on a miss = %d, %v, want 7, true`, v, computed)
	}
	v, computed = c.GetOrCompute("a", func() int { return 99 })
	if v != 7 || computed {
		t.Errorf(`GetOrCompute("a") on a hit = %d, %v, want 7, false - compute must not run`, v, computed)
	}
	if got := c.Len(); got != 1 {
		t.Errorf("Len() = %d after computing one key, want 1", got)
	}
}

func TestGetOrComputeStoresWhatItComputed(t *testing.T) {
	var c Cache
	c.GetOrCompute("a", func() int { return 42 })
	if v, ok := c.Get("a"); !ok || v != 42 {
		t.Errorf(`Get("a") = %d, %v after GetOrCompute, want 42, true - the value was not stored`, v, ok)
	}
}

func TestGetOrComputeRunsComputeExactlyOnceUnderContention(t *testing.T) {
	var c Cache
	const goroutines = 32
	var computes counter
	// Every goroutine is at the door before any of them knocks, so they all
	// miss the first check together. Without the second check under the write
	// lock, several of them compute.
	wait := newBarrier(goroutines)
	values := make([]int, goroutines)
	flags := make([]bool, goroutines)

	waitFor(t, "concurrent GetOrCompute on one key", func() {
		var wg sync.WaitGroup
		for i := range goroutines {
			wg.Add(1)
			go func() {
				defer wg.Done()
				wait()
				values[i], flags[i] = c.GetOrCompute("shared", func() int {
					computes.add()
					return 1234
				})
			}()
		}
		wg.Wait()
	})

	if got := computes.value(); got != 1 {
		t.Errorf("compute ran %d times for one missing key, want exactly 1", got)
	}
	trues := 0
	for i := range goroutines {
		if values[i] != 1234 {
			t.Fatalf("goroutine %d got %d, want the one computed value 1234", i, values[i])
		}
		if flags[i] {
			trues++
		}
	}
	if trues != 1 {
		t.Errorf("%d callers were told they computed the value, want exactly 1", trues)
	}
}

func TestGetOrComputeIsCorrectAcrossManyKeys(t *testing.T) {
	var c Cache
	const keys, goroutines = 50, 8
	var computes counter

	waitFor(t, "concurrent GetOrCompute across keys", func() {
		var wg sync.WaitGroup
		for range goroutines {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for k := range keys {
					v, _ := c.GetOrCompute(fmt.Sprint(k), func() int {
						computes.add()
						return k * 10
					})
					if v != k*10 {
						t.Errorf("GetOrCompute(%d) = %d, want %d", k, v, k*10)
						return
					}
				}
			}()
		}
		wg.Wait()
	})

	if got := computes.value(); got != keys {
		t.Errorf("compute ran %d times for %d distinct keys, want %d - once each", got, keys, keys)
	}
	if got := c.Len(); got != keys {
		t.Errorf("Len() = %d, want %d", got, keys)
	}
}
