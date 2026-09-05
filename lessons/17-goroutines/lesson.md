# Goroutines

Concurrency is the part of Go people came for, and the part most likely to be
learned in the wrong order. The keyword is trivial. What takes practice is
knowing who is allowed to touch which piece of memory, and that turns out to be
a question you can answer without any concurrency vocabulary at all.

So this lesson does the whole of the first half - starting work and waiting for
it - and deliberately leaves channels for the next one. You can write a great
deal of useful parallel code with nothing but `go`, `sync.WaitGroup` and a
slice, and code written that way is easier to read than the channel version of
the same thing.

## `go` starts, and returns

```go
go doWork()
```

That is it. `doWork` now runs independently, and the `go` statement returns
immediately - before `doWork` has necessarily executed a single line. The
arguments are evaluated **now**, on the calling goroutine; the call happens
later, on a new one.

A goroutine is not a thread. It starts with a small stack, a couple of
kilobytes, that grows as it needs to, and the Go runtime multiplexes many of
them onto a handful of operating-system threads. Starting ten thousand is
routine. Starting ten thousand threads is not, which is most of why Go looks
the way it does.

Two consequences fall straight out of "returns immediately":

**Nobody waits.** When `main` returns, the process exits. Goroutines still
running are dropped mid-statement, with no unwinding and no `defer`. A program
whose answer is computed in a goroutine and never waited for is a program that
prints nothing.

**You do not know the order.** Not "it is usually this order". You do not know.
Two goroutines started one line apart can run in either order, on either
thread, interleaved at any instruction. Any correctness you want has to be
arranged, not observed.

## `sync.WaitGroup`

A `WaitGroup` is a counter with a queue of sleepers attached. `Add` raises it,
`Done` lowers it, `Wait` blocks until it is zero.

```go
var wg sync.WaitGroup
for _, task := range tasks {
	wg.Add(1)
	go func() {
		defer wg.Done()
		task()
	}()
}
wg.Wait()
```

Three rules, and they are the same rule from three angles:

- **`Add` before `go`, on the calling goroutine.** Move it inside and `Wait`
  can see a counter of zero and return before anything has started.
- **`Done` exactly once, via `defer`.** An early `return` or a `panic` in the
  middle still decrements, so `Wait` still returns instead of hanging.
- **Never copy a `WaitGroup`.** The copy has its own counter and its `Done`
  never reaches the original. Close over it, or pass a `*sync.WaitGroup`.

Since Go 1.22 each loop iteration gets its own copy of the loop variables, so
the classic "every goroutine saw the last element" bug is gone. Older code
still opens its loop bodies with `task := task`; that line is now a no-op, and
seeing it tells you the code predates 1.22.

## The rule that matters: who owns what

Goroutines share memory. Two of them writing the same variable at the same time
is a **data race**, and a data race is not a program that gets the wrong answer
sometimes. It is a program with no defined behaviour at all - the compiler is
allowed to assume it does not happen, so the result can be a torn value, a
stale read, or something no source-level reasoning predicts.

`total += v` looks atomic and is not. It is a load, an add and a store, and two
goroutines interleaving those lose updates.

There are locks for this, and they are the next concept but one. Reach for them
second. The first move is **not to share**:

```go
out := make([]int, len(in))     // every index exists before anything starts
for i, v := range in {
	wg.Add(1)
	go func() {
		defer wg.Done()
		out[i] = f(v)          // only this goroutine ever touches index i
	}()
}
wg.Wait()
return out                     // safe to read: Wait is the barrier
```

Nothing here is locked, because nothing here is shared. Each goroutine writes
one element that no other goroutine writes. The slice header is only read.
`make` rather than `append` is load-bearing: `append` from several goroutines
races on the header, which is both easier to write and harder to spot.

`Wait` is doing two jobs. The obvious one is "do not return early". The other
is that it is the point at which the writes inside the goroutines are
guaranteed to be visible to the code after it. Reading `out` before `Wait` is
the same race in disguise.

Writing the result at the input's index buys something else for free: **the
output is in input order**, whatever order the goroutines actually finished in.
Ordering is usually treated as something concurrency takes away from you. It
only takes it away if you append.

## Splitting the work

One goroutine per element is fine when there are hundreds of elements and each
one is expensive. It is a bad trade for a million cheap ones, where the
bookkeeping costs more than the work. The usual shape is a fixed number of
workers, and there are two ways to divide the items between them.

**Contiguous chunks.** Worker `w` takes a block: `[0,3) [3,6) [6,8) [8,10)`.
The arithmetic is the interesting part - `n/workers` has a remainder, and
dropping it drops elements, so the first `n%workers` chunks take one extra
item each. Good when every element costs about the same. Good for cache
locality too, since each worker walks a contiguous span.

**Striding.** Worker `w` takes `w, w+workers, w+2*workers, ...`:

```
items    a  b  c  d  e  f  g  h  i
worker   0  1  2  0  1  2  0  1  2
```

Better when costs vary and cluster, because a run of expensive items next to
each other is shared out instead of landing on one worker. It is not load
balancing - a worker that draws three slow items still finishes last - but it
is one line of arithmetic, and it is usually enough.

Both preserve the ownership property. Index `i` belongs to exactly one worker
either way, so the safety argument does not change; only which worker does.

And in both, the concurrent part is the middle: **split, work, combine**. The
split happens before any goroutine starts, the combine after every one has
finished. That is why there is no lock anywhere in this lesson.

## Errors, and determinism

When the per-item work can fail, several workers can fail at once, and
returning whichever error arrives first makes the function non-deterministic:
the same input gives different errors on different runs, and the test that
checks it either turns flaky or gets watered down to `err != nil`.

Give errors the same treatment as results - one cell per item - and pick after
`Wait`:

```go
for _, err := range errs {
	if err != nil {
		return nil, err     // the lowest failing index, every time
	}
}
```

Deterministic for one extra slice. What it does **not** do is stop the other
workers: every item is still processed even if item 0 failed immediately.
Cancelling in-flight work needs a way to tell goroutines to give up, which is a
channel, and then a `context`. Those are the next two concepts, in that order.

## When it is actually faster

Often it is not. Starting a goroutine costs more than adding a hundred
integers, and a parallel sum of a small slice will lose to the one-line loop.
Parallelism pays when the *per-element* work is real - a parse, a hash, a
request - or when the slice is very large. `runtime.NumCPU()` is the usual
worker count, and more workers than cores buys nothing for CPU-bound work.

Measure before and after. Concurrency is a tool for a specific problem, not a
general improvement, and the version with no goroutines in it is easier to read
and impossible to race.

## The checklist

- `go` returns immediately; nothing waits for the goroutine.
- `Add` before `go`, `defer wg.Done()` inside, `Wait` on the caller.
- Two goroutines writing one variable is undefined behaviour, not a slow path.
- Prefer not sharing at all: preallocate, one cell per goroutine.
- Write results at the input's index and ordering comes free.
- The concurrent part is only the middle: split, work, combine.

## Further reading

- [Goroutines](https://go.dev/doc/effective_go#goroutines) - Effective Go on
  what a goroutine costs and why there can be thousands of them.
- [Go statements](https://go.dev/ref/spec#Go_statements) - the two lines of
  spec behind `go f(x)`, including when the arguments are evaluated.
- [sync](https://pkg.go.dev/sync) - `WaitGroup`, with the rule about calling
  `Add` before the goroutine starts.
- [Data Race Detector](https://go.dev/doc/articles/race_detector) - what
  `-race` does, what it can and cannot see, and how to run it.
- [runtime](https://pkg.go.dev/runtime) - `NumCPU` and `GOMAXPROCS`, for the
  worker count question.

## Practise

Three challenges. The first is the shape itself - `RunAll`, a slice of squares
computed one goroutine per element, and a `Gather` that keeps input order when
the last function finishes first. The second splits a slice into contiguous
chunks and sums it in parallel, with the chunk arithmetic as its own pure
function so the off-by-one is testable without a goroutine in sight. The third
strides instead of chunking, generically, and adds the deterministic error
rule: the lowest failing index wins, on every run.
