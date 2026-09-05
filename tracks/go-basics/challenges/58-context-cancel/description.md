# Context Cancellation

A goroutine you cannot stop is a leak. The last three concepts started
goroutines and waited for them to finish on their own terms; real programs need
the other direction too - a request is abandoned, a user closes the tab, the
server is shutting down, and every goroutine working on that job should notice
and stop.

[`context.Context`](https://pkg.go.dev/context#Context) is how Go says "stop".
One value, passed down through every call, carrying a single broadcast: *this
work is no longer wanted*.

## The shape of it

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
```

[`context.Background()`](https://pkg.go.dev/context#Background) is the empty
root - no deadline, never cancelled, the context you start from in `main` or a
test. [`WithCancel`](https://pkg.go.dev/context#WithCancel) **derives** a child
from it and hands you the switch. Contexts form a tree: cancelling a parent
cancels every descendant, and cancelling a child leaves the parent alone.

Three things a context gives you, and only three:

- `ctx.Done()` - a channel that is **closed** when the work should stop. A
  closed channel is receivable for ever by every goroutine at once, which is
  exactly what a broadcast needs.
- `ctx.Err()` - `nil` while the context is live, then
  [`context.Canceled`](https://pkg.go.dev/context#Canceled) or
  [`context.DeadlineExceeded`](https://pkg.go.dev/context#DeadlineExceeded)
  once it is not.
- `ctx.Value(key)` - request-scoped data, and the subject of the third
  challenge in this concept.

## Listening for it

Cancellation is a channel, so it composes with everything you learned about
`select`:

```go
select {
case v := <-work:
	return v, nil
case <-ctx.Done():
	return 0, ctx.Err()
}
```

Two rules from that snippet are worth stating on their own.

**Return `ctx.Err()`, not an error of your own.** `errors.Is(err,
context.Canceled)` is how a caller several frames up tells "the user went away"
from "the database is broken", and a hand-written
[`errors.New("cancelled")`](https://pkg.go.dev/errors#New) throws that away.

**A send needs the same treatment as a receive.** `out <- v` blocks until
somebody receives; if the consumer has already given up, that goroutine is
parked for the lifetime of the process. Wrap it in the same select.

## Checking without waiting

`select` with a `default` never blocks: if no case is ready, the default runs
immediately.

```go
select {
case <-ctx.Done():
	return ctx.Err()
default:
}
```

That is the check to put at the top of each iteration of a long loop, so a
cancelled job stops at the next step instead of at the end.

## Conventions

- `ctx` is the **first** parameter, named `ctx`, and typed `context.Context` -
  not stored in a struct.
- The function that creates a context calls `cancel`, always, usually with
  `defer`. Skipping it leaks the timer and the child bookkeeping until the
  parent dies. `go vet` calls this out as "the cancel function is not used".
- Never pass a `nil` context.
  [`context.TODO()`](https://pkg.go.dev/context#TODO) is the placeholder for
  "there will be one here later".

## Task

Implement `Wait`, `Cancelled`, `CountTicks` and `Produce`.

## Hints

- `Wait` is one `select` with two cases, and the cancelled case returns
  `ctx.Err()`.
- `Cancelled` is the same `select` with a `default` and no work in it.
- `CountTicks` needs the two-value receive: a closed `ticks` is always ready,
  so `v := <-ticks` alone would spin for ever counting zeros.
- `Produce` starts a goroutine, `defer close(out)` inside it, and sends each
  value through a select against `ctx.Done()`. Return the channel straight
  away - the sends happen after you have returned.
