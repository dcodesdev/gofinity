# RWMutex Cache

A [`sync.Mutex`](https://pkg.go.dev/sync#Mutex) lets one goroutine in at a
time, and that is right when the code inside changes something. For a cache it
is wasteful: reads do not interfere with each other, and a cache is almost all
reads.

[`sync.RWMutex`](https://pkg.go.dev/sync#RWMutex) splits the lock in two.

```go
c.mu.RLock()          // a read lock: any number at once
defer c.mu.RUnlock()

c.mu.Lock()           // a write lock: one, and no readers
defer c.mu.Unlock()
```

Many `RLock` holders can hold it together. A `Lock` waits until every reader has
let go, and while a writer is waiting new readers queue behind it - which is
what stops a stream of readers starving the writer for ever.

The pairs do not mix: `RLock` is released by `RUnlock`, `Lock` by `Unlock`.
Mismatch them and you get a runtime panic or a permanent hang.

It is not free. An `RWMutex` is slower than a `Mutex` for a lock held only
briefly, so the rule of thumb is plain `Mutex` first, `RWMutex` once you have
many readers and long enough read sections to pay for it.

## Not reentrant, in either direction

A goroutine that already holds the read lock must not take it again, and must
never take the write lock while holding the read lock. Both deadlock. That is
why `ForEach` here documents that `f` must not call back into the cache: `f`
runs while the read lock is held.

## The double check

Reading first and writing only if you must is the standard cache shape, and the
gap in the middle is the whole lesson:

```go
c.mu.RLock()
v, ok := c.entries[key]
c.mu.RUnlock()
if ok {
	return v, false
}

c.mu.Lock()
defer c.mu.Unlock()
if v, ok := c.entries[key]; ok {   // look again
	return v, false
}
...
```

Between `RUnlock` and `Lock` this goroutine holds **no lock at all**. Any number
of other goroutines can run in that window, and one of them may store exactly
the key you are about to compute. Without the second check both of you compute
it and the loser's value is thrown away - or worse, kept, if `compute` opened
something that now leaks.

Upgrading a read lock to a write lock in place is not possible in Go, and that
is deliberate: two goroutines both trying to upgrade would deadlock. Release,
retake, look again.

## Computing under the lock

`GetOrCompute` calls `compute` while holding the **write** lock. That is what
makes "exactly once" true, and it costs exactly what you would expect: every
other key waits behind it, however unrelated.

The last concept's rule - do not hold a lock across a caller's function - is
being broken on purpose here, and it is a real trade. The version that does not
break it needs a lock **per key**, so that two different keys can be computed at
the same time while two callers of the same key still share one call. That is
the next challenge.

## Task

Implement `Cache`: `Get`, `Set`, `Len`, `Keys`, `ForEach` and `GetOrCompute`.

## Hints

- Reads take `RLock`, writes take `Lock`. `Get`, `Len`, `Keys` and `ForEach`
  are all reads; `Set` is a write; `GetOrCompute` is a read that may become a
  write.
- The map is `nil` until something writes to it. Only the two write paths need
  to create it.
- `Keys` returns `make([]string, 0, len(c.entries))` filled and sorted with
  [`slices.Sort`](https://pkg.go.dev/slices#Sort), so an empty cache gives an
  empty non-nil slice.
- `GetOrCompute` returns `computed` `false` on a hit, including the hit it
  finds on the second look.
- The tests hold a barrier open inside `ForEach`'s `f` across three concurrent
  calls. Only a read lock lets all three arrive.
