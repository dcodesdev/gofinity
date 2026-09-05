# Channels

The previous lesson started work and waited for it, and shared memory only by
giving each goroutine an index nobody else touched. That covers a surprising
amount of real code. What it cannot do is let goroutines talk while they run:
stream results as they appear, stop early, wait on two things at once.

A channel is the tool for that. It is a typed pipe with synchronisation built
in, and once you know exactly what blocks and who closes, everything else in Go
concurrency is composition.

## Sending and receiving

```go
ch := make(chan int)
go func() { ch <- 42 }()
v := <-ch          // 42
```

The arrow points the way the value travels: `ch <- v` sends, `<-ch` receives.
The type is part of the channel - a `chan int` carries `int` and nothing else.

An **unbuffered** channel is a rendezvous. A send does not complete until a
receive is ready and a receive does not complete until a send is ready; neither
side gets ahead of the other. That is why the send above is inside a goroutine.
On the main goroutine it would block with nobody left to receive, and the
runtime would notice that every goroutine was asleep and end the program with a
deadlock message.

So the rule that catches everyone exactly once: **a channel needs two
goroutines.** A send and its receive in the same one, in that order, is a
deadlock you wrote on purpose without meaning to.

## Buffers

`make(chan int, 3)` adds a queue of three:

```go
ch := make(chan int, 3)
ch <- 1
ch <- 2   // neither blocks: there is room
```

- A send blocks only when the buffer is **full**.
- A receive blocks only when the buffer is **empty**.

`cap(ch)` is the size it was made with and never changes. `len(ch)` is how many
values are queued right now, and "right now" is the catch: with another
goroutine sending, the number is stale the instant you read it. `len(ch) <
cap(ch)` is a trustworthy "there is room" only when you are the sole sender.

A buffer is not a speed knob. It is for two things. **Decoupling** a bursty
producer from a bursty consumer, so a pause on one side does not stall the
other - and when the buffer fills the producer waits again, which is back
pressure working, not failing. And **counting**, which the semaphore below is
built on.

## Closing

The sender closes, and only the sender:

```go
close(ch)
```

Closing does not discard what is buffered. It says no more will arrive. A
receive from a closed and drained channel returns immediately with the zero
value, and the two-value form distinguishes the cases:

```go
v, ok := <-ch    // ok is false once the channel is closed and empty
```

`for v := range ch` is that loop written out: it receives until the channel
closes, then stops. If nobody closes, it waits for ever. Closing is not
housekeeping, it is how the receiver learns the stream ended.

Two things panic rather than fail quietly: **sending on a closed channel** and
**closing twice**. Both are ownership bugs, which is why the convention is
firm: whoever creates a channel and sends on it is the one who closes it. A
receiver that closes its input has broken somebody else's send.

Closing is also not needed for garbage collection. A channel nobody references
is collected whether or not it was closed. You close to communicate, not to
clean up.

## Direction in the type

```go
func Produce(n int) <-chan int
func Consume(ch <-chan int) int
func Fill(ch chan<- int)
```

`<-chan int` is receive-only, `chan<- int` is send-only. A bidirectional
`chan int` converts to either automatically, so this costs nothing at the call
site and makes ownership a compile error instead of a code review comment. A
caller holding `<-chan int` cannot send and cannot close. Use it every time.

## The generator shape

A function that produces a stream returns the receiving end and does the work
in a goroutine it starts itself:

```go
func Produce(n int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := range n {
			out <- i
		}
	}()
	return out
}
```

Read it twice, because nearly everything else is a variation. `Produce` returns
the channel, not the values, so it does not block. The goroutine owns `out`,
sends on it, and closes it with a `defer` on its first line so no path can
forget. The caller receives at its own pace and the unbuffered channel makes
the producer follow that pace exactly.

A **stage** is the same shape with an input:

```go
func Double(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			out <- v * 2
		}
	}()
	return out
}
```

Stages compose into pipelines: `Double(Produce(5))`. Note that a stage must
forward one value at a time. Draining the input into a slice first and only
then sending deadlocks against an unbuffered producer, and it also throws away
the streaming that was the point.

## A buffered channel as a semaphore

Capacity `n` is `n` permits. A send takes one, a receive gives it back:

```go
sem := make(chan struct{}, limit)
for _, job := range jobs {
	wg.Add(1)
	go func() {
		defer wg.Done()
		sem <- struct{}{}
		defer func() { <-sem }()
		work(job)
	}()
}
wg.Wait()
```

Every goroutine still starts at once, because goroutines are cheap. Only
`limit` of them are ever past the `sem <-` line. That is how you fan out over
ten thousand items without opening ten thousand connections. `chan struct{}` is
the idiom for a channel whose values carry no information: `struct{}` occupies
no memory.

## `select`

A receive waits on one channel. `select` waits on several and takes whichever
is ready first:

```go
select {
case v := <-a:
	use(v)
case v := <-b:
	use(v)
}
```

Exactly one case runs. If several are ready it picks **at random**, which is
deliberate: no channel can starve the others by being written first. If none
are ready it blocks until one is.

Sends are cases too - `case ch <- v:` - and that is the arm most people never
reach for, although it is the whole of graceful shutdown.

**`default` makes it non-blocking.** If no case is ready, `select` takes the
default instead of waiting:

```go
select {
case v := <-ch:
	return v, true
default:
	return 0, false
}
```

That is the honest version of "is there anything in there" that `len(ch)` only
pretends to be. Note that a closed channel **is** ready - immediately, for
ever, with the zero value - so it takes its own case and never the default.

**A timeout is just another channel.** `time.After(d)` returns a channel that
receives after `d`:

```go
select {
case v := <-ch:
	return v, true
case <-time.After(100 * time.Millisecond):
	return 0, false
}
```

Go needs no timeout syntax because the clock is a channel. `time.After` holds
its timer until it fires, so in a tight loop prefer `time.NewTimer` and `Stop`;
for a one-shot wait it is fine.

## Cancellation, and the leak it prevents

A goroutine blocked for ever on a send never returns and its stack is never
freed. No crash, no error, just a process that grows: a **goroutine leak**.
Anything that produces has to be tellable to stop.

The tool is a channel carrying no value that is only ever closed:

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

The send and the `done` receive must be arms of the **same** `select`. Written
as a bare send followed by a check, the goroutine parks on a send nobody is
waiting for and never looks at `done` again.

Cancellation is signalled by **closing**, never by sending, because a close
makes the receive ready for every watcher at once. One `close` stops a thousand
goroutines; a send stops one of them, chosen at random.

Two more facts worth memorising here. A receive from a **nil** channel blocks
for ever, and so does a send - which sounds useless until you notice it lets
you disable a `select` case by setting its channel to `nil`. And this whole
`done` pattern is what `context.Context` packages up, with deadlines and values
attached, which is the next concept.

## Fan-in

Merging many channels into one is the `WaitGroup` you already know:

```go
out := make(chan int)
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
	close(out)
}()
return out
```

The closer has to be its own goroutine. `wg.Wait()` on the calling one would
block until every input drained, and nothing is draining `out` yet because
`Merge` has not returned it. Closing before the last sender finished would
panic, which is why it waits at all.

## Channels or a mutex

The slogan is "share memory by communicating", and it is good advice pointed at
the wrong half of the time. Channels are for **handing a value over**: a
pipeline, a queue of work, a signal, a result. A `sync.Mutex` is for
**protecting a piece of state that stays where it is**: a counter, a cache, a
map several goroutines read and write.

Reaching for a channel to guard a shared counter produces something slower and
harder to read than three lines with a mutex. That is the next concept, and it
is a shorter one.

## The checklist

- Unbuffered is a rendezvous; buffered blocks only when full, or empty.
- A channel needs two goroutines. One is a deadlock.
- `cap` is fixed, `len` is a snapshot and lies the moment anyone else sends.
- The sender closes, exactly once. Sending on a closed channel panics.
- `range` ends at the close. No close, no end.
- Return `<-chan T`, take `chan<- T`: ownership in the type.
- A stage forwards one value at a time and closes its own output.
- Capacity `n` is `n` permits.
- `select` picks at random among ready cases; `default` makes it non-blocking.
- A timeout is a case, because `time.After` is a channel.
- Cancel by closing a `done` channel, in the same `select` as the send.

## Further reading

- [Channel types](https://go.dev/ref/spec#Channel_types) - direction, capacity,
  and the exact rules for send, receive and `close`.
- [Select statements](https://go.dev/ref/spec#Select_statements) - why a ready
  case is chosen at random and what `default` changes.
- [Pipelines and cancellation](https://go.dev/blog/pipelines) - stages,
  closing downstream, and the `done` channel this lesson ends on.
- [Concurrency](https://go.dev/doc/effective_go#channels) - Effective Go on
  channels as both values and semaphores.
- [time](https://pkg.go.dev/time) - `After`, `NewTimer` and `Tick`, the
  channels a timeout case is built from.

## Practise

Three challenges. The first is the vocabulary: a generator, a `range`-based
collector, and a pipeline stage that has to forward one value at a time or
deadlock. The second is buffering - a channel prefilled without any goroutine
at all, `len` against `cap` in the one situation where that is honest, and a
semaphore that runs a bounded number of calls at once and proves it. The third
is `select` in all three forms: the `default`, the timeout, a fan-in that
closes only after its last input, and a generator that has to notice a closed
`done` rather than leaking the goroutine that would otherwise wait for ever.
