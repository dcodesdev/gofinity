# Context Deadlines

A cancellation you have to remember to trigger is a cancellation you will
forget. Most work should stop on its own if it takes too long: a request that
has been running for thirty seconds is not going to become useful at
thirty-one.

[`context.WithTimeout`](https://pkg.go.dev/context#WithTimeout) and
[`context.WithDeadline`](https://pkg.go.dev/context#WithDeadline) are the same
tool with two faces. A **deadline** is a moment; a **timeout** is a duration
from now. Pick whichever the caller already has.

```go
ctx, cancel := context.WithTimeout(parent, 2*time.Second)
defer cancel()
```

`defer cancel()` is not optional. It releases the timer and unhooks the child
from its parent; skipping it leaks both until the parent dies, and
[`go vet`](https://pkg.go.dev/cmd/vet) reports it as "the cancel function is
not used".

## A child is never laxer than its parent

This is the property that makes contexts safe to pass around. Deriving a
context can only tighten it:

```go
parent, _ := context.WithTimeout(ctx, time.Second)
child, _ := context.WithTimeout(parent, time.Hour)   // still one second
```

`child` inherits the parent's deadline, so a helper that "gives itself an hour"
cannot outlive the request it was called for. There is no way to extend a
deadline, only to start a fresh context from `Background()` - which you should
do consciously and rarely, for work that genuinely outlives the request that
started it.

`ctx.Deadline()` reads it back: the moment, and whether there is one at all.
[`time.Until(deadline)`](https://pkg.go.dev/time#Until) turns that into a
budget you can pass to something else, or use to decide the remaining time is
not worth spending.

## Two errors, not one

Once a context is done, `ctx.Err()` is one of exactly two values:

- [`context.Canceled`](https://pkg.go.dev/context#Canceled) - somebody called
  `cancel`.
- [`context.DeadlineExceeded`](https://pkg.go.dev/context#DeadlineExceeded) -
  time ran out.

They mean completely different things to a caller. A cancellation is usually
normal: the user navigated away, a sibling request already failed. A timeout is
a signal about your system - something was too slow, and somebody may want to
retry or alert.

Keep them distinguishable all the way up. Wrap with `%w`, join with
[`errors.Join`](https://pkg.go.dev/errors#Join) when there are genuinely two
stories to tell, and test with [`errors.Is`](https://pkg.go.dev/errors#Is).
Never flatten them into a string.

```go
return fmt.Errorf("fetching user %d: %w", id, ctx.Err())
```

## Racing work you cannot interrupt

Plenty of functions know nothing about contexts: an old library, a CPU-bound
loop, a blocking system call. You cannot stop them, but you can stop *waiting*
for them:

```go
done := make(chan result, 1)  // buffered, and this is why
go func() { done <- do() }()

select {
case r := <-done:
	return r.value, r.err
case <-ctx.Done():
	return 0, ctx.Err()
}
```

The buffer is the whole trick. When the timeout wins, nobody will ever receive
from `done`; on an unbuffered channel that goroutine would be parked on its
send for the lifetime of the process. One slot of buffer lets it deliver a
result nobody wants and exit.

That work still runs to completion in the background - `select` bounds how long
you wait, not how long it takes. If the work holds something expensive, teach it
about the context instead.

## Task

Implement `Budget`, `Limit`, `Do`, `Retry` and `Classify`.

## Hints

- `Budget` is `ctx.Deadline()` plus `time.Until`. Do not clamp the result at
  zero: a negative budget is real information.
- `Limit` needs no comparison at all. Deriving already does the capping - work
  out which one-line call that is.
- `Do` is the buffered-channel race above.
- `Retry` checks `ctx.Err()` at the top of each turn, so an expired context
  costs no attempts, and reports `n - 1` attempts when it gives up on turn `n`.
- `errors.Join(last, ctx.Err())` keeps both errors findable, and it drops nils,
  so it works even when there is no last error to report.
- `Classify` is a `switch` with no expression, `errors.Is` in each case.
