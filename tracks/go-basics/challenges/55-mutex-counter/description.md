# Mutex Counter

The last two concepts moved values between goroutines through channels. This one
does the opposite: it lets several goroutines touch the **same** memory, and
makes that safe.

You need it more often than it sounds. A counter, a cache, a map of results, a
connection pool: anything one goroutine writes and another reads.

## Why `n++` is not one step

```go
var n int
go func() { n++ }()
go func() { n++ }()
```

`n++` is three machine steps - read `n`, add one, write it back. Two goroutines
can read the same value, both add one to it, and both write the same result
back. One increment is simply gone. That is a **data race**, and Go says a
program with one has no defined behaviour at all: not "a stale number", no
behaviour. The [race detector](https://go.dev/doc/articles/race_detector),
`go test -race` and `go run -race`, exists to find them.

## The mutex

[`sync.Mutex`](https://pkg.go.dev/sync#Mutex) is a lock with two methods.
Between `Lock` and `Unlock` exactly one goroutine is running that code;
everyone else waits.

```go
type Counter struct {
	mu sync.Mutex
	n  int
}

func (c *Counter) Add(delta int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n += delta
}
```

Four rules, and the whole concept is in them:

- **The mutex goes next to what it guards**, directly above those fields. In Go
  that placement is the documentation.
- **`defer c.mu.Unlock()` on the line after `Lock`.** An early return or a
  panic between them would otherwise leave the lock held for ever, and every
  other goroutine parked behind it.
- **Reads lock too.** `Value()` needs the lock as much as `Add` does. A read
  racing a write is a race whether or not the answer looks plausible.
- **Pointer receivers, always.** A method on a value receiver gets a *copy* of
  the struct, mutex included, and a copy of a mutex guards nothing.
  [`go vet`](https://pkg.go.dev/cmd/vet) catches this as "passes lock by
  value".

## The zero value works

`sync.Mutex`'s zero value is an unlocked mutex, so a struct containing one needs
no constructor:

```go
var c Counter   // ready
```

That is worth keeping for your own types. It does mean a map field is still
`nil`, so a type that guards a map creates it on first write - under the lock,
like every other access.

## Never hand out what you guard

```go
func (t *Tally) Snapshot() map[string]int   // a copy
```

A map or a slice returned from behind a lock is a *reference* to the guarded
memory, and the caller keeps using it long after `Unlock`. Return a copy, or
return the one value the caller asked for. The lock protects the memory, not the
function call.

The mirror of that rule: **do not hold a lock across someone else's code**. If a
caller's function runs while you hold the lock, your program is serial again -
and if that function calls back into your type, it deadlocks, because
`sync.Mutex` is not reentrant. Relocking a mutex you already hold is a hang, not
an error.

## Task

Implement `Counter`, `Tally` and `CountMatching`.

## Hints

- Both structs need their fields written as well as their methods; the starter
  leaves them empty on purpose.
- `Tally`'s map is `nil` until the first `Record`. Reading a `nil` map is fine
  and gives 0, so only `Record` has to create it.
- `Snapshot` copies key by key. `out := t.counts` copies the reference, not the
  map.
- `CountMatching` is last concept's `WaitGroup` with this concept's `Counter`:
  start every goroutine, call `pred` **outside** the lock, `Add(1)` on a match,
  `wg.Wait()`, then read the value.
