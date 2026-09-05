# Channel Basics

A channel is a typed pipe between goroutines. One goroutine sends, another
receives, and the channel takes care of the handoff:

```go
ch := make(chan int)   // unbuffered
go func() { ch <- 42 }()
v := <-ch              // 42
```

The arrow always points in the direction the value travels: `ch <- v` sends,
`<-ch` receives.

## An unbuffered channel is a rendezvous

`make(chan int)` with no capacity means a send does not complete until a
receive is ready, and a receive does not complete until a send is ready.
Neither side runs ahead of the other. That is why the send above is inside a
goroutine: on the main goroutine it would block forever, with nobody left to
receive, and Go would call that a deadlock and kill the program.

So the rule that catches everyone once: **a channel needs two goroutines.**
If the send and the receive are in the same one, in that order, you have
written a deadlock.

## Closing, and ranging

The sender - and only the sender - closes:

```go
close(ch)
```

Closing does not discard buffered values; it says "no more will arrive".
A receive from a closed, drained channel returns immediately with the zero
value, and the two-value form tells you which happened:

```go
v, ok := <-ch   // ok is false once the channel is closed and empty
```

`range` over a channel is that loop written out for you. It receives until the
channel is closed and then stops:

```go
for v := range ch {
	fmt.Println(v)
}
```

If nobody ever closes, `range` waits forever. Closing is not optional
housekeeping: it is how the receiver learns the stream ended.

Two things that panic rather than fail quietly: sending on a closed channel,
and closing a channel twice. Both are bugs about *ownership*, which is why the
convention is that whoever created the channel and sends on it is the one who
closes it.

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

The return type `<-chan int` is a **receive-only** channel: the caller can read
and cannot send or close. That is the type system enforcing ownership, and it
costs nothing, so use it every time.

Note the `defer close(out)` on the first line of the goroutine, and note that
`Produce` returns immediately - it returns the channel, not the values. The
first send blocks until the caller receives, and that is fine, because the
caller is a different goroutine.

## Task

Implement `Produce`, `Collect` and `Double`.

## Hints

- `Produce` is the generator above.
- `Collect` is a `range` over the channel appending into a slice. Return an
  empty, non-nil slice when the channel closes with nothing in it.
- `Double` is a **pipeline stage**: it takes a receive-only channel and returns
  another, forwarding each value doubled. It must return the output channel
  straight away and do the forwarding in a goroutine, one value at a time -
  reading the whole input before sending anything deadlocks against an
  unbuffered producer.
- Every one of the three closes exactly the channel it created, and none of
  them closes a channel it was given.
