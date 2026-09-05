package main

import (
	"fmt"
	"slices"
	"sync"
)

// Cache is a map of string to int that many goroutines may read at once and one
// may write at a time.
//
// Its zero value is an empty cache. The map is created on the first write.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]int
}

// Get returns the value stored for key, and whether it was there at all.
//
// This is a read: it takes the read lock, so any number of Gets can run
// together.
func (c *Cache) Get(key string) (int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.entries[key]
	return v, ok
}

// Set stores value under key, replacing anything already there.
func (c *Cache) Set(key string, value int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.entries == nil {
		c.entries = map[string]int{}
	}
	c.entries[key] = value
}

// Len returns how many entries the cache holds.
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// Keys returns every key in the cache, sorted, so the result is worth comparing.
// An empty cache returns an empty, non-nil slice.
func (c *Cache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]string, 0, len(c.entries))
	for k := range c.entries {
		keys = append(keys, k)
	}
	// Sorting happens under the lock only because it is cheap and touches no
	// caller code. The rule it does not break: nothing here can block.
	slices.Sort(keys)
	return keys
}

// ForEach calls f for every entry, in an unspecified order, while holding the
// read lock.
//
// That is the point of it: a caller that walked Keys and then Got each one
// would see a cache that changed underneath it, whereas ForEach sees one
// consistent moment. The price is written on the tin - f must not call back
// into the cache, because a write would block behind the read lock f is
// running under, and a sync.RWMutex is not reentrant.
func (c *Cache) ForEach(f func(key string, value int)) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for k, v := range c.entries {
		f(k, v)
	}
}

// GetOrCompute returns the value for key, calling compute to produce it if the
// cache does not have one. computed reports whether compute ran.
//
// Even when many goroutines ask for a missing key at the same time, compute
// runs exactly once and every caller gets that one value.
//
// The shape is the double check:
//
//	read lock:   if it is there, return it and stop
//	write lock:  look again, because another goroutine may have stored it
//	             while we had no lock at all, then compute and store
//
// The second look is not paranoia. Between releasing the read lock and taking
// the write lock this goroutine holds nothing, and that gap is exactly where
// another one wins the race.
func (c *Cache) GetOrCompute(key string, compute func() int) (int, bool) {
	// The fast path: a hit costs a read lock, which is what makes the read
	// lock worth having at all.
	c.mu.RLock()
	v, ok := c.entries[key]
	c.mu.RUnlock()
	if ok {
		return v, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	// The second check. Without it, two goroutines that both missed would both
	// compute, and the second would overwrite the first.
	if v, ok := c.entries[key]; ok {
		return v, false
	}
	if c.entries == nil {
		c.entries = map[string]int{}
	}
	// compute runs while the write lock is held, which is what makes "exactly
	// once" true and every other key wait. See the description.
	v = compute()
	c.entries[key] = v
	return v, true
}

func main() {
	var c Cache
	c.Set("a", 1)
	fmt.Println(c.Get("a"))
	fmt.Println(c.GetOrCompute("b", func() int { return 2 }))
	fmt.Println(c.GetOrCompute("b", func() int { return 99 }))
	fmt.Println(c.Keys(), c.Len())
}
