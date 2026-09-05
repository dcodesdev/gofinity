# Benchmark Basics

A benchmark is a function that looks almost like a test:

```go
func BenchmarkJoin(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SinkString = JoinFields(parts, ",")
	}
}
```

`go test -bench=.` runs it, and the framework chooses `b.N`: once with `N = 1`,
then with larger and larger values until the whole loop has run for about a
second. It divides the total by `N` and prints nanoseconds per operation. That
is why the loop is not optional - a body that ignores `b.N` does its work once
however many iterations were asked for, and the framework, seeing an
instantaneous benchmark, keeps asking for more until it gives up at a billion.

Everything about writing a good benchmark comes from that one fact: the loop
body is measured, `N` times, and nothing else should be.

## Setup does not belong in the measurement

```go
func BenchmarkJoin(b *testing.B) {
	parts := BuildParts(20000)   // slow, and not what is being measured
	b.ReportAllocs()
	b.ResetTimer()               // clock and counters back to zero
	for i := 0; i < b.N; i++ {
		SinkString = JoinFields(parts, ",")
	}
}
```

[`b.ResetTimer()`](https://pkg.go.dev/testing#B.ResetTimer) throws away
everything up to that line: the elapsed time *and* the allocation counters.
Without it the setup is amortised over `N` iterations, which sounds harmless
until you notice `N` is chosen from a measurement that included the setup, and
the allocation count is dominated by it. A benchmark that reports 700 allocs/op
for a function that does one is almost always a missing `ResetTimer`.

[`b.ReportAllocs()`](https://pkg.go.dev/testing#B.ReportAllocs) asks for the
`B/op` and `allocs/op` columns for this one benchmark, the same thing
`-benchmem` asks for globally. Allocation counts are the most stable number a
benchmark produces - they do not change when the machine is busy - so it is
worth turning on.

## Work the loop cannot skip

If the result of the body is thrown away, the compiler is free to throw the
body away with it, and you have benchmarked an empty loop. Assign it to a
package-level variable:

```go
var SinkString string
```

Package level matters: a local would be provably dead too.

## StopTimer for per-iteration setup

Some work destroys its input. `FilterInPlace` compacts a slice in place, so
running it twice on the same slice is not the same operation twice. Each
iteration needs a fresh slice, and building one is not the thing being measured:

```go
for i := 0; i < b.N; i++ {
	b.StopTimer()
	nums := BuildNumbers(20000)
	b.StartTimer()
	SinkInts = FilterInPlace(nums, keep)
}
```

[`StopTimer`](https://pkg.go.dev/testing#B.StopTimer) and
[`StartTimer`](https://pkg.go.dev/testing#B.StartTimer) pause and resume both
the clock and the allocation counters, so what is left is the filtering alone.
They are not free - do not wrap a nanosecond of work in them - but for a
rebuild that costs more than the call, they are the difference between
measuring the filter and measuring the builder.

## Reading the numbers

```
BenchmarkJoin-8   4180   265292 ns/op   163840 B/op   1 allocs/op
```

`4180` is `N`, `265292 ns/op` is the mean. The mean of a whole loop, not of one
call - so a benchmark that is 3ns/op is measuring something the optimiser
already deleted, and one whose `N` is exactly `1000000000` never looped.

Compare two runs with
[`benchstat`](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat); do not
eyeball two numbers and declare a 3% win. And measure before optimising:
`-benchmem` and a profile will tell you where the allocations are, and almost
every "faster Go" is really "fewer allocations".

## The three functions

Each one is a small lesson in that last sentence:

- **`JoinFields`** - `s += part` allocates a new string every time round the
  loop. A [`strings.Builder`](https://pkg.go.dev/strings#Builder) grown to the
  final size once allocates twice at most, whatever the length of the input.
  Two loops: one to add up the lengths, one to write.
- **`FilterInPlace`** - `xs[:0]` is an empty slice that still owns xs's array,
  so appending to it writes over elements that have already been read. Zero
  allocations, and a caller who keeps using the old slice gets what they
  deserve, which is why the doc comment says so.
- **`ParseSum`** - `string(data)` copies.
  [`strings.Split`](https://pkg.go.dev/strings#Split) allocates a slice and a
  header per field. Walking the bytes and folding digits as you go does
  neither.

## Task

Implement `JoinFields`, `FilterInPlace` and `ParseSum`, then write `BenchJoin`
and `BenchFilter` in the shapes above.

The tests measure the benchmarks you write - they run them with
[`testing.Benchmark`](https://pkg.go.dev/testing#Benchmark) and compare against
a correctly shaped reference - so a missing `ResetTimer` or a missing
`StopTimer` is a failure here, not just a misleading number.

## Hints

- `strings.Builder`'s `Grow` takes the number of **bytes**, so the separator's
  length counts `len(parts)-1` times.
- `FilterInPlace` is `out := xs[:0]` and an `append` inside an `if`. The read
  index is always ahead of the write index, so nothing is lost.
- `ParseSum` can walk once: track the start of the current field, and when you
  reach a comma or the end, fold that slice into a number. An empty field, a
  bare `-`, and any non-digit are errors.
- Return a package-level error value from `ParseSum` rather than building one
  with [`fmt.Errorf`](https://pkg.go.dev/fmt#Errorf); the tests do not read the
  message.
- `b.ReportAllocs()` and `b.ResetTimer()` go after the setup and before the
  loop. In `BenchFilter` there is no setup before the loop at all - it is all
  per-iteration, which is exactly why it is stopped and started instead.
