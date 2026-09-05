# Select and Timeout

A receive waits on one channel. `select` waits on several at once and runs
whichever is ready first:

```go
select {
case v := <-a:
	use(v)
case v := <-b:
	use(v)
}
```

Exactly one case runs. If several are ready `select` picks one **at random**,
which is deliberate: it means no channel can starve the others just by being
listed first. If none are ready it blocks until one is.

Sends are cases too - `case ch <- v:` - and that is the piece most people
never reach for. A `select` can wait to hand a value out and to be told to stop
at the same time, which is the whole of graceful shutdown.

## `default` makes it non-blocking

```go
select {
case v := <-ch:
	return v, true
default:
	return 0, false   // nothing was ready, and we did not wait
}
```

With a `default`, `select` never blocks: if no case is ready it takes the
default and moves on. This is the honest version of "is there anything in the
channel", the one `len(ch)` only pretends to be.

Note the difference between a `default` and a closed channel. A receive from a
closed channel **is** ready - immediately, forever, with the zero value - so a
closed channel takes its own case and never the default.

## Timeouts

[`time.After(d)`](https://pkg.go.dev/time#After) returns a channel that
receives a value after `d`. As a `select` case it is a deadline:

```go
select {
case v := <-ch:
	return v, true
case <-time.After(100 * time.Millisecond):
	return 0, false
}
```

The clock is just another channel, which is why Go needs no special timeout
syntax. `time.After` leaks its timer until it fires, so in a hot loop use
[`time.NewTimer`](https://pkg.go.dev/time#NewTimer) and `Stop` it; for a
one-shot wait like the one above it is fine.

## The done channel

The reason a producer must be able to stop is that a goroutine blocked forever
on a send is a goroutine that never returns, and its stack is never freed. That
is a **goroutine leak**: no crash, no error, just a program that grows.

The fix is a second channel that carries no value and is only ever closed:

```go
func Ticker(done <-chan struct{}) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 0; ; i++ {
			select {
			case out <- i:
			case <-done:
				return
			}
		}
	}()
	return out
}
```

Closing `done` makes every receive on it ready at once, so one `close` stops
any number of goroutines. That is why cancellation is signalled by closing a
channel and never by sending on it: a send reaches one receiver, a close
reaches all of them. This shape is exactly what
[`context.Context`](https://pkg.go.dev/context#Context) wraps up, and that is
the next concept.

## Fan-in

Merging several channels into one is the same `WaitGroup` you already know,
with a goroutine per input and a closer goroutine at the end:

```go
var wg sync.WaitGroup
for _, ch := range chs {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for v := range ch {
			out <- v
		}
	}()
}
go func() {
	wg.Wait()
	close(out)      // after the last input goroutine, not before
}()
```

The closer has to be its own goroutine: `wg.Wait()` on the calling one would
block until the inputs drained, and nothing is draining `out` yet.

## Task

Implement `TryRecv`, `RecvTimeout`, `Merge` and `GenerateUntil`.

## Hints

- `TryRecv` is the `default` form and `RecvTimeout` is the `time.After` form.
  Both are four lines.
- `Merge` is the fan-in above. Merging no channels returns a channel that is
  already closed - `out` is created, no input goroutine is ever started, so the
  `WaitGroup` is at zero and the closer fires straight away.
- `GenerateUntil` is `Ticker`. The send and the `done` receive have to be arms
  of the **same** `select`, or it blocks on a send nobody is waiting for and
  never looks at `done` again.
- None of the four closes a channel it was given.
