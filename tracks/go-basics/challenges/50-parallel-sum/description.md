# Parallel Sum

Summing a slice is one line. Summing it across four goroutines is the smallest
honest example of the pattern behind almost all parallel work: **split, work,
combine**.

```
[ 0 1 2 3 4 5 6 7 8 9 ]      split into contiguous chunks
  \___/ \___/ \__/ \__/      one goroutine per chunk
    6     15    13   17      each writes its own partial
              51             the caller adds the partials up
```

Only the middle step is concurrent. The split happens before any goroutine
starts and the combine happens after all of them have finished, so the shared
state is a slice of partials in which every goroutine owns exactly one cell.
Nothing is locked because nothing is shared.

## Splitting evenly

`len(nums) / workers` is not the answer on its own, because the division has a
remainder and dropping it drops elements. Distribute it: with `n` items and `w`
workers, the first `n % w` chunks get one extra item each.

With `n = 10` and `w = 4`: `10/4` is `2` remainder `2`, so the sizes are
`3, 3, 2, 2` and the ranges are `[0,3) [3,6) [6,8) [8,10)`. Every element lands
in exactly one chunk and no two chunks differ in size by more than one, which
is what "evenly" means here.

Two edge cases are worth deciding once rather than at every call site:

- **More workers than items.** Never produce an empty chunk. Ten workers over
  three items is three chunks of one, not three chunks and seven empties.
- **`workers` of zero or less.** Treat it as one. A caller passing
  [`runtime.NumCPU()`](https://pkg.go.dev/runtime#NumCPU) on a strange machine
  should get a correct answer, not a panic.

Doing this as its own function is not ceremony. Range arithmetic is where the
off-by-one lives, and a pure function that returns ranges can be tested without
starting a single goroutine.

## Partials, not a shared total

The tempting version is wrong:

```go
total := 0
for _, r := range ranges {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, v := range nums[r[0]:r[1]] {
			total += v          // data race: read, add, write, from N goroutines
		}
	}()
}
```

`total += v` is three operations, and two goroutines interleaving them lose
updates. Worse, it is *undefined*, not merely lossy: the [race
detector](https://go.dev/doc/articles/race_detector) exists because "it seemed
to work" is not evidence.

The fix costs one slice:

```go
partials := make([]int, len(ranges))
// ... goroutine i writes only partials[i] ...
wg.Wait()
total := 0
for _, p := range partials {
	total += p
}
```

Same answer, no sharing, and the result does not depend on the order the
goroutines ran in. Integer addition is associative, so the total is exactly the
sequential total - every element counted once, in some order that does not
matter.

## Is it faster?

For summing a small slice: no. Starting a goroutine costs more than adding a
hundred integers, and the parallel version will lose to the one-line loop until
the per-element work is real. Parallelism is a tool for expensive elements, not
for many cheap ones. Learn the shape here, where it is easy to check, and spend
it where the work justifies it.

## Task

Implement `Chunks`, `ParallelSum` and `ParallelCount`.

## Hints

- `Chunks(n, workers)` returns `[][2]int` of half-open `[start, end)` ranges
  covering `[0, n)` in order. `Chunks(0, 4)` is empty.
- Track a running `start`, and give the first `n % workers` chunks one extra
  element.
- Both parallel functions are the same six lines: `Chunks`, a `partials` slice,
  a `WaitGroup`, one goroutine per range, `Wait`, then a sequential fold.
- `nums[r[0]:r[1]]` is the goroutine's slice. Reslicing does not copy, and
  reading the same backing array from many goroutines is fine - it is writing
  that needs care.
