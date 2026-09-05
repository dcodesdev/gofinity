# Buffered Channels

`make(chan int, 3)` gives the channel room for three values:

```go
ch := make(chan int, 3)
ch <- 1
ch <- 2   // neither of these blocks: there is space
```

An unbuffered channel is a rendezvous - the sender waits for the receiver. A
buffered channel is a queue with a fixed size, and the two rules follow from
that:

- **A send blocks only when the buffer is full.**
- **A receive blocks only when the buffer is empty.**

A closed channel is still drainable: closing does not throw away what is
already in the buffer. `range` keeps handing out the queued values and stops
after the last one.

## `len` and `cap`

Channels answer both:

```go
ch := make(chan int, 3)
ch <- 7
len(ch)   // 1 - values waiting in the buffer
cap(ch)   // 3 - the size it was made with
```

`cap` never changes. `len` is a snapshot, and that word matters: with another
goroutine sending or receiving, the value is already stale by the time you look
at it. `len(ch) < cap(ch)` is a reliable "there is room" only when **you are
the only sender**, which is exactly the case in this challenge and almost never
the case in real concurrent code. The general answer is `select` with a
`default`, and that is the next lesson.

## What a buffer is actually for

Not speed. Two things:

**Decoupling a producer from a consumer** that are both bursty, so a brief
pause on one side does not stall the other. The buffer absorbs the jitter, and
once it is full the producer waits again - back pressure is the feature, not
the flaw.

**Counting.** A buffered channel of capacity `n` is a semaphore: `n` permits,
handed out by a send and returned by a receive.

```go
sem := make(chan struct{}, limit)
for _, job := range jobs {
	wg.Add(1)
	go func() {
		defer wg.Done()
		sem <- struct{}{}        // take a permit, blocking while all are out
		defer func() { <-sem }() // give it back
		work(job)
	}()
}
wg.Wait()
```

Every goroutine still starts at once - goroutines are cheap - but only `limit`
of them are ever past the `sem <-` line. That is how you fan out over ten
thousand items without opening ten thousand connections. `chan struct{}` is the
idiom for a channel whose values carry no information: a `struct{}` occupies no
memory at all.

## Task

Implement `Buffered`, `FillUpTo` and `MapLimited`.

## Hints

- `Buffered` needs no goroutine at all. Make the channel with capacity
  `len(values)`, send them all - none of those sends can block - then close it
  and return the receiving end.
- `FillUpTo` is the one place `len(ch) < cap(ch)` is honest, because the test
  is the only other party. Return how many values you managed to send.
- `MapLimited` is the semaphore above with the result slice from the previous
  concept: preallocate with `make`, let each goroutine write its own index, and
  `Wait` before returning. A `limit` below 1 means 1.
- `defer func() { <-sem }()` inside the goroutine, not after `wg.Done()` - the
  permit must come back even if `work` panics.
