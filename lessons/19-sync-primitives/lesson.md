# Sync Primitives

Two lessons ago the advice was "do not communicate by sharing memory; share
memory by communicating", and a channel did the sharing for you. It is good
advice and it is not the whole story. The Go proverb has a second half almost
nobody quotes: *use whichever is more expressive*.

A counter is not more expressive as a channel. Neither is a cache, a set of
seen URLs, or a configuration loaded once at startup. Those are shared memory,
and Go ships three small tools that make shared memory safe. This is the
shortest concept in the track, and the one whose rules you will use daily.

## What a race actually is

```go
var n int
go func() { n++ }()
go func() { n++ }()
```

`n++` is not one operation. It reads `n`, adds one, and writes the result back.
Two goroutines can read the same value, both compute the same sum, and both
store it: two increments, one result.

That is a **data race** - two goroutines touching the same memory, at least one
of them writing, with nothing ordering them. The Go memory model does not say
you get a stale value. It says the program has **no defined behaviour at all**.
A racy program may produce a number that was never written, tear a slice header
into halves from two different slices, or crash inside the runtime.

Races are not reliably reproducible, so do not hunt them by staring. Go has a
detector built in:

```sh
go test -race ./...
go run -race .
```

It instruments memory access and reports the two goroutines and the two stack
traces the moment a race happens on a code path you exercised. It costs roughly
ten times the CPU and a lot of memory, which is why it is a test-and-staging
tool rather than a production one. Run your test suite under it in CI; it is the
single highest-value thing you can do to a concurrent Go program.

## sync.Mutex

A mutex is a lock. Between `Lock` and `Unlock`, exactly one goroutine runs that
code and everyone else waits.

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

func (c *Counter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}
```

Five rules carry the whole concept.

**The mutex sits above what it guards.** Go has no syntax that ties a lock to a
field, so the placement is the documentation: the mutex, then the fields it
protects, and nothing else in between. If a struct has two unrelated groups of
fields, it usually wants two mutexes.

**`defer mu.Unlock()` on the line after `Lock`.** An early return or a panic
between them leaves the lock held for ever and parks every other goroutine
behind it. The defer costs a few nanoseconds; getting this wrong costs the
process. Unlock without defer only when you have measured that you need to,
which is the double-check in the next section.

**Reads lock too.** `Value` needs the lock as much as `Add`. A read racing a
write is a race, however plausible the answer looks. "It is only an int" is
wrong: an unsynchronised read may see a value the compiler cached in a register
from before the write ever happened.

**Pointer receivers, always.** A value receiver copies the struct, mutex
included, and a copy of a mutex guards nothing - the two copies lock two
different locks. `go vet` reports this as "passes lock by value", and it is one
of the vet checks `go test` runs for you by default.

**Never hold a lock across code you do not control.** A caller's callback, a
network call, another method of your own that locks again. `sync.Mutex` is not
reentrant: locking a mutex your goroutine already holds is a permanent hang, not
an error. And what you return from behind a lock matters as much as what you
call: handing back the guarded map or slice gives the caller a reference that
outlives the `Unlock`. Return a copy.

The zero value of a mutex is unlocked, so a struct containing one needs no
constructor. `var c Counter` is ready. A guarded map field is still `nil`
though, so create it on first write - under the lock, like everything else.

## sync.RWMutex

A cache is almost all reads, and readers do not interfere with each other.
`sync.RWMutex` splits the lock in two:

```go
c.mu.RLock()      // a read lock: any number of holders at once
defer c.mu.RUnlock()

c.mu.Lock()       // a write lock: one holder, and no readers
defer c.mu.Unlock()
```

Many readers hold `RLock` together. A `Lock` waits for every one of them to let
go, and while a writer is waiting new readers queue behind it - which is what
stops a stream of readers starving the writer for ever.

`RLock` pairs with `RUnlock` and `Lock` with `Unlock`. Mixing them panics or
hangs. And an `RWMutex` is not reentrant in either direction: taking the read
lock twice in one goroutine can deadlock behind a waiting writer, and taking the
write lock while holding the read lock always does.

It is not automatically faster. An `RWMutex` has more bookkeeping than a
`Mutex`, so for a lock held for a handful of instructions the plain one wins.
Reach for `RWMutex` when you have measured many concurrent readers holding the
lock for long enough to matter.

## The double check

Read first, write only if you must, and the gap in the middle is the lesson:

```go
c.mu.RLock()
v, ok := c.entries[key]
c.mu.RUnlock()
if ok {
	return v
}

c.mu.Lock()
defer c.mu.Unlock()
if v, ok := c.entries[key]; ok {   // look again
	return v
}
v = compute()
c.entries[key] = v
return v
```

Between `RUnlock` and `Lock` this goroutine holds **nothing**. Any number of
others can run in that window, and one of them may store the very key you are
about to compute. Without the second look, both of you compute it and one result
is thrown away - or leaked, if `compute` opened something.

There is no lock upgrade in Go, and its absence is deliberate: two goroutines
both upgrading would each wait for the other to release. Release, retake, look
again.

## sync.Once

Some work must happen exactly once. A `bool` guard is the same broken race as
`n++`, and a mutex around the flag still lets the second caller return before
the first has finished.

```go
var once sync.Once
once.Do(f)
```

`f` runs on the first `Do` and never again. Every other caller **blocks until
that first `f` has returned**, which is the half people forget: `Do` is not
"skip if busy", it is "the work is complete by the time you get your turn". So
whatever `f` wrote is safe to read the moment `Do` returns, with no lock of your
own.

The zero value is ready, and a `Once` must never be **copied** after use - keep
it behind a pointer or captured in a closure, never in a map value or a value
receiver. A panic inside `f` still counts as done, so `Once` is not a retry.

Two shapes are worth memorising. The lazy value:

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

and the idempotent shutdown, because closing an already closed channel panics
and there is no way to ask a channel whether it is closed:

```go
func (c *Closer) Close() {
	c.closeOnce.Do(func() { close(c.done) })
}
```

The standard library has `sync.OnceValue`, `sync.OnceValues` and `sync.OnceFunc`
since Go 1.21, and they are what you should reach for in real code. Write them
once yourself first.

### Once per key

Combine the two and you get the pattern behind every real cache: one lock for
the map, and a separate `Once` for each entry.

```go
mu.Lock()
e, ok := entries[key]
if !ok {
	e = &entry{}          // a pointer: a Once must not be copied
	entries[key] = e
}
mu.Unlock()

e.once.Do(func() { e.value = f(key) })   // outside the map's lock
```

The map's lock is held only while the map is touched. The waiting happens on the
entry, so two callers of the same key share one computation while two callers of
different keys never wait for each other at all. That is `singleflight` in eight
lines.

## The rest of the package

`sync.WaitGroup` you already know from the goroutines lesson. Three more, in one
line each:

- **`sync/atomic`** - `atomic.Int64`, `atomic.Bool`, `atomic.Pointer[T]` and
  friends are single values updated without a lock. Faster than a mutex for one
  counter or one flag, and no help at all the moment two fields must change
  together.
- **`sync.Map`** - a map tuned for two specific shapes: keys written once and
  read many times, or disjoint key sets per goroutine. It is not a faster map in
  general, it loses all type safety, and a plain map behind an `RWMutex` beats it
  most of the time. Measure before choosing it.
- **`sync.Cond`** - wait until a condition changes. Almost always a channel does
  the job more clearly; reach for it last.

## Choosing between a channel and a mutex

Both are correct. The question is which one says what you mean.

Use a **channel** when the value is moving: a pipeline stage, a result being
handed to whoever asked, a cancellation broadcast, work distributed to a pool.

Use a **mutex** when the value is staying put and several goroutines need to
look at it: a counter, a cache, a registry, a set of what has been seen.

The tell is in how it reads. A counter behind a channel needs a goroutine that
owns it and a request-reply protocol to ask it anything - a dozen lines and a
lifetime to manage where three lines with a mutex would do. A pipeline behind a
mutex needs condition variables and hand-rolled queues. Write the one that is
shorter to explain.

## The checklist

- `n++` is three steps. Two goroutines and one unguarded variable is a race, and
  a race has no defined behaviour.
- `go test -race` finds them; run it in CI.
- The mutex goes directly above the fields it guards.
- `defer mu.Unlock()` on the next line.
- Reads lock too.
- Pointer receivers: a copied mutex guards nothing, and `go vet` says so.
- Never hold a lock across a caller's function, and never return the memory you
  guard - return a copy.
- Mutexes are not reentrant. Relocking is a hang, not an error.
- `RWMutex` for many concurrent readers; plain `Mutex` until you have measured.
- No lock upgrades. Release, retake, and look again.
- `Once.Do` runs once and every other caller waits for it to finish.
- A `Once` must not be copied; store `*entry`, not `entry`.
- Channels move values; mutexes guard values that stay.

## Further reading

- [sync](https://pkg.go.dev/sync) - `Mutex`, `RWMutex`, `Once`, `WaitGroup` and
  `Map`, each with the copying rule stated on it.
- [sync/atomic](https://pkg.go.dev/sync/atomic) - the typed atomics, for the
  counter that needs no lock at all.
- [The Go Memory Model](https://go.dev/ref/mem) - the document that defines
  what "a race has no defined behaviour" means, precisely.
- [Data Race Detector](https://go.dev/doc/articles/race_detector) - running
  `go test -race`, and reading the report it prints.
- [Share by communicating](https://go.dev/doc/effective_go#sharing) - Effective
  Go on choosing between a channel and a mutex.

## Practise

Three challenges. The first is the mutex itself: a counter that loses no
increments under fifty goroutines, a tally whose map is created lazily under the
lock and snapshotted as a copy, and a fan-out that calls a predicate outside the
lock. The second splits the lock - a cache where reads run together, `ForEach`
holds the read lock across the walk, and `GetOrCompute` does the double check.
The third is `sync.Once` in its three shapes: a lazy value, a `Close` that any
number of goroutines may call, and a memoizer with a `Once` per key that
computes different arguments at the same time and the same argument only once.
