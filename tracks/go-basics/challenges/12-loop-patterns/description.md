# Loop Patterns

Go has one loop keyword. There is no `while`, no `do`, no `foreach`, and no
comprehension. `for` covers all of it by dropping clauses:

```go
for i := 0; i < n; i++ { }   // counter
for n != 1 { }               // while
for { }                      // forever, until break
for i, v := range items { }  // each element
```

`break` ends the innermost loop and `continue` skips to its next iteration. When
you need to leave an *outer* loop from inside an inner one, label it and say
which: `break search`. That label is not a `goto`; it names a loop, and it is
the readable alternative to a `found` flag threaded through two conditions.

`range` over a slice gives you index and value. Take only the index with
`for i := range items`, and drop the index with `for _, v := range items`. The
value is a **copy**, so assigning to `v` changes nothing.

## Task

Finish the six functions in `main.go`. Each one is a different shape of loop.

1. `SumTo(n int) int` adds 1 through `n` with the three-clause form. `n` below 1
   sums to `0`.
2. `CollatzSteps(n int) int` counts steps to reach 1: halve an even number,
   triple an odd one and add 1. `CollatzSteps(1)` is `0`, `CollatzSteps(27)` is
   `111`, and anything below 1 is `-1`.
3. `OddSquares(n int) []int` returns the squares of the odd numbers up to `n`,
   using `continue` to skip the even ones.
4. `NextPowerOfTwo(n int) int` returns the smallest power of two at least as
   large as `n`, with a `for` that has no clauses and a `break`.
5. `FirstIndex(items []string, target string) int` returns the index of the
   first match, or `-1`, and stops as soon as it finds one.
6. `FindCell(grid [][]int, target int) (int, int)` scans a grid row by row and
   returns the row and column of the first matching cell, or `-1, -1`. Rows may
   have different lengths, including zero.

## Hints

- A `for i := 1; i <= n; i++` never runs when `n` is `0`, so the "n below 1"
  cases need no extra `if`. Check that before writing one.
- The Collatz loop needs its guard first: `n < 1` returns `-1` before the loop,
  because the loop would otherwise never reach 1.
- `continue` runs the post statement on its way round, so the counter still
  advances. A `continue` in a condition-only `for` that forgets to advance the
  variable itself is an infinite loop, and in the runner that is a timeout
  rather than an error message.
- Appending to a `nil` slice is fine, so `var out []int` is a complete
  initialisation. `OddSquares(0)` then returns `nil`, which has length zero and
  is what the tests expect.
- `FirstIndex` can just `return i` from inside the range; that is a `break` that
  also hands back the answer.
- `FindCell` is the one that wants a label. Write `search:` on the line above
  the outer `for`, then `break search` from the inner loop. Do it with a plain
  `break` instead and the outer loop keeps going, which quietly returns the
  *last* match rather than the first.
- Ranging over an empty or `nil` slice runs the body zero times, so the empty
  row in the test grid needs no special case.
