# Fan Out Work

The previous challenge split a slice into contiguous blocks. That is the right
split when every element costs the same. It is the wrong split when they do
not: give worker 0 the first quarter and worker 3 the last, and if the
expensive items happen to sit at the front, worker 0 is still going long after
the other three have finished. The whole run takes as long as its unluckiest
chunk.

**Striding** spreads that luck out. Worker `w` of `n` takes indices
`w, w+n, w+2n, ...`:

```
items    a  b  c  d  e  f  g  h  i
worker   0  1  2  0  1  2  0  1  2
```

Worker 0 gets `a, d, g`; worker 1 gets `b, e, h`; worker 2 gets `c, f, i`. A run
of expensive items next to each other is now shared out rather than dumped on
one goroutine. It is not load balancing - a worker that draws three slow items
still finishes last - but it is one line of arithmetic, and for work whose cost
varies unpredictably it is usually enough.

```go
for i := worker; i < len(items); i += workers {
	out[i] = f(items[i])
}
```

The ownership property that makes the previous challenge safe still holds:
index `i` is touched by exactly one worker, because `i % workers` has exactly
one value. Striding changes *which* worker, not *how many*.

## Order is not finishing order

`out[i] = f(items[i])` writes the result where the input was, so the output is
in input order however the scheduler shuffled the goroutines. That property is
worth defending in a test, because it is the first thing to break when someone
"optimises" the function into appending.

## Errors from a concurrent run

When `f` can fail, several workers can fail at once, and "the error" stops being
obvious. Returning whichever one arrives first makes the function
non-deterministic: the same input gives different errors on different runs, and
a test either becomes flaky or gets weakened to `err != nil`.

Pick the rule instead: report the error of the **lowest failing index**. Collect
every error into its own cell, exactly like the results, and scan the slice in
order once every goroutine has finished:

```go
for _, err := range errs {
	if err != nil {
		return nil, err
	}
}
```

Deterministic, and it costs one extra slice. Note also what this does *not* do:
it does not cancel the remaining work. Every item is processed even when item 0
failed on the first goroutine to run. Stopping early needs a way to tell the
other goroutines to give up, which is a channel and then a
[`context`](https://pkg.go.dev/context) - the next two concepts.

## Task

Implement `Assignments`, `FanOut` and `FanOutErr`.

## Hints

- `Assignments(items, workers)` returns the index list for each worker, so
  `Assignments(9, 3)` is `[[0 3 6] [1 4 7] [2 5 8]]`. Same edge cases as before:
  `workers < 1` means one, and never return a worker with nothing to do.
- `FanOut` and `FanOutErr` are generic over the input and output types. Write
  the loop with the stride directly - you do not have to call `Assignments`,
  though you may.
- On failure `FanOutErr` returns `nil` results and the error from the lowest
  index, not a half-filled slice.
- `f` is called exactly once per element, on all three functions, even when
  another element has already failed.
