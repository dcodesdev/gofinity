# Sync Once

Some work must happen exactly once: open the connection pool, parse the
configuration, close the shutdown channel. A `bool` guard does not do it -
`if !done { done = true; ... }` is the same three-step race as `n++`, and a
mutex around it still leaves the second caller returning before the first has
finished the work.

[`sync.Once`](https://pkg.go.dev/sync#Once) is the answer, and it is one
method.

```go
var once sync.Once
once.Do(f)
```

- `f` runs on the **first** call to `Do`, and never again on that `Once`.
- Every other caller **blocks until that first `f` has returned**. That is the
  half people forget: `Do` is not "skip if busy", it is "the work is finished by
  the time you get your turn". So whatever `f` wrote is safe to read the moment
  `Do` returns, with no lock of your own.
- The zero value is ready, so a `Once` needs no constructor - but it must never
  be **copied** after use. Keep it in a struct you pass by pointer, or captured
  by a closure, never in a map value or a value receiver.
- It is once *per `Once` value*, not once per program. Two `Once`s do their work
  twice.

`Do` counts a call as done even if `f` panics, so the work will not be retried.
If retrying matters, that is a mutex and a flag you manage yourself.

## Once as a lazy value

The closure form is the one you will write most:

```go
func OnceValue(load func() int) func() int {
	var once sync.Once
	var value int
	return func() int {
		once.Do(func() { value = load() })
		return value
	}
}
```

Nothing loads until somebody asks, and then only once. The standard library has
this as [`sync.OnceValue`](https://pkg.go.dev/sync#OnceValue) since Go 1.21;
writing it once yourself is the point here.

## Once as "closed"

Closing an already closed channel panics, and there is no way to ask a channel
whether it is closed without receiving from it. So an idempotent shutdown signal
is a `Once` around the `close`:

```go
func (c *Closer) Close() {
	c.closeOnce.Do(func() { close(c.done) })
}
```

Any goroutine may now call `Close`, as often as it likes, and every waiter on
`Done()` is woken exactly once. Keeping the zero value usable means the channel
itself has to be created lazily too, by whichever of `Done` and `Close` is
reached first: a second `Once`.

## Once per key

The last challenge's `GetOrCompute` ran `compute` under the one write lock, so
unrelated keys queued behind each other. Give **each key its own `Once`** and
both properties hold at the same time:

```go
mu.Lock()
e, ok := entries[key]
if !ok {
	e = &entry{}          // a pointer: a Once must not be copied
	entries[key] = e
}
mu.Unlock()

e.once.Do(func() { e.value = f(key) })   // outside the lock
```

The map's lock is held only while the map is touched. The waiting happens on
the entry, so two callers of the same key share one call and two callers of
different keys do not wait for each other at all. This is
[`singleflight`](https://pkg.go.dev/golang.org/x/sync/singleflight) in eight
lines, and it is worth being able to write from memory.

## Task

Implement `OnceValue`, `Closer` and `Memoize`.

## Hints

- `OnceValue` captures the `Once` and the value in the closure. Do not run
  `load` before the returned function is first called.
- `Closer` needs two `Once`s: one that makes the channel, one that closes it.
  `Done` and `Close` both call the first.
- `IsClosed` is a `select` with a `default` - a receive from a closed channel is
  always ready, so the default arm is taken exactly while it is open.
- `Memoize`'s map holds `*entry`, not `entry`. A map value is copied on the way
  out, and a copied `Once` has forgotten everything.
- If the "different arguments at the same time" test hangs, the map's lock is
  still held while `f` runs.
