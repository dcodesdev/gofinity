package main

import "fmt"

// Cache is a map of string to int that many goroutines may read at once and one
// may write at a time.
//
// Its zero value is an empty cache. The map is created on the first write.
type Cache struct {
	// TODO: the entries, and a sync.RWMutex above them.
}

// Get returns the value stored for key, and whether it was there at all.
//
// This is a read: it takes the read lock, so any number of Gets can run
// together.
func (c *Cache) Get(key string) (int, bool) {
	// TODO
	return 0, false
}

// Set stores value under key, replacing anything already there.
func (c *Cache) Set(key string, value int) {
	// TODO
}

// Len returns how many entries the cache holds.
func (c *Cache) Len() int {
	// TODO
	return 0
}

// Keys returns every key in the cache, sorted, so the result is worth comparing.
// An empty cache returns an empty, non-nil slice.
func (c *Cache) Keys() []string {
	// TODO
	return nil
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
	// TODO
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
func (c *Cache) GetOrCompute(key string, compute func() int) (value int, computed bool) {
	// TODO
	return 0, false
}

func main() {
	var c Cache
	c.Set("a", 1)
	fmt.Println(c.Get("a"))
	fmt.Println(c.GetOrCompute("b", func() int { return 2 }))
	fmt.Println(c.GetOrCompute("b", func() int { return 99 }))
	fmt.Println(c.Keys(), c.Len())
}
