# Goroutines and WaitGroup

A goroutine is a function running independently of the one that started it:

```go
go doWork()
```

`go` returns immediately. `doWork` may not have started yet, may be halfway
through, may already be finished - you do not know, and you are not allowed to
guess. Goroutines are cheap: a few kilobytes of stack that grows as needed, so
starting thousands is ordinary.

## `main` does not wait

Nothing waits for a goroutine. When `main` returns the whole program exits and
every goroutine still running is discarded mid-step, with no unwinding and no
`defer`. A goroutine whose result nobody waits for is a goroutine whose result
you may never see.

So the first thing to learn is not how to start work. It is how to wait for it.

## `sync.WaitGroup`

A [`WaitGroup`](https://pkg.go.dev/sync#WaitGroup) is a counter with a queue
attached:

```go
var wg sync.WaitGroup
for _, task := range tasks {
	wg.Add(1)          // before the go statement, never inside it
	go func() {
		defer wg.Done() // exactly once, whatever happens
		task()
	}()
}
wg.Wait()              // returns when the counter reaches zero
```

Three rules, and all three are the same rule seen from different sides:

- **`Add` before `go`.** Inside the goroutine it is a race: `Wait` can see a
  counter of zero and return before the goroutine has run a line.
- **`Done` exactly once,** which is what `defer` buys you - a `return` or a
  `panic` in the middle still decrements.
- **Do not copy a `WaitGroup`.** Copying it copies the counter, so the copy's
  `Done` never reaches the original. Pass a `*sync.WaitGroup`, or close over it
  as the loop above does.

Since Go 1.22 each iteration of a `for` loop gets its own copy of the loop
variables, so the old "every goroutine sees the last value" bug is gone. Older
code you read will still have `task := task` at the top of the loop body; it is
now a no-op.

## Where the results go

Goroutines share memory, and two of them writing to the same variable at once
is a **data race**: not a slow program, an undefined one. The simplest fix is
not to share. Give each goroutine its own cell:

```go
out := make([]int, len(in))
for i, v := range in {
	wg.Add(1)
	go func() {
		defer wg.Done()
		out[i] = f(v)     // only this goroutine ever touches index i
	}()
}
wg.Wait()
return out
```

Every goroutine writes to a different element of the same slice, so no two of
them touch the same memory. The slice header itself is only read. After `Wait`
returns, every write has happened and the caller can read the whole thing -
`Wait` is the point where the results become visible, and reading `out` before
it is the same race in a different disguise.

Note also that `out` is preallocated with `make`. `append` from several
goroutines would be a race on the slice header, which is a much easier mistake
to make and a much harder one to see.

## Task

Implement `RunAll`, `Squares` and `Gather`.

## Hints

- `RunAll` is the loop from above with nothing else in it.
- `Squares(n)` returns `[]int{0, 1, 4, 9, ...}` of length `n`, one goroutine per
  element. `n` of 0 returns an empty, non-nil slice - `make([]int, 0)` does that.
- `Gather` is `Squares` generalised: run every function concurrently and return
  their results in the **input** order, not the order they finished in. Indexing
  the output slice gives you that for free.
- If a test hangs rather than fails, a `wg.Add` is missing or a `wg.Done` is not
  deferred.
