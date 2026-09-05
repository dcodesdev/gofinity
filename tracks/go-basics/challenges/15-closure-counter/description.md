# Closure Counter

In Go a function is a value. It has a type, it can be stored in a variable,
passed as an argument, returned from another function and kept in a slice:

```go
var op func(int) int = func(n int) int { return n * 2 }
op(21)   // 42
```

A function literal written inside another function can refer to that function's
variables, and those variables **outlive the call**. That is a closure:

```go
func Counter() func() int {
	n := 0
	return func() int {
		n++
		return n
	}
}

next := Counter()
next()   // 1
next()   // 2
```

`n` is not copied into the returned function; the function holds a reference to
it. Two things follow, and both matter. Each call to `Counter()` creates a
*new* `n`, so two counters count independently. And any closure that captures
the same variable sees the same value, so two functions returned together share
their state on purpose.

## Task

Fill in the seven functions in `main.go`.

1. `Counter() func() int` counts 1, 2, 3 on successive calls.
2. `Accumulator() func(int) int` keeps a running total and returns it.
3. `Multiplier(n int) func(int) int` multiplies by the captured `n`.
4. `Apply(nums []int, f func(int) int) []int` maps `f` over a slice into a new
   one, leaving the input alone.
5. `Compose(f, g func(int) int) func(int) int` applies `g` first, then `f`.
6. `Countdown(from int) func() (int, bool)` yields `from` down to `1`, then
   `0, false` forever.
7. `Multipliers(ns []int) []func(int) int` returns one multiplier per element.

## Hints

- The state goes *outside* the returned literal. `n := 0` inside the inner
  function resets on every call and the counter is stuck at 1.
- `Accumulator` needs no `if`: the total starts at its zero value, and the tests
  include a step that brings it back to `0`.
- `Apply` builds a new slice with `append` onto a `var out []int`. Writing back
  into `nums[i]` would change the caller's data, and one test checks for that.
- Read `Compose(f, g)` right to left, like the maths: `f(g(x))`. Swapping the
  two gives `Compose(double, increment)(3) = 7` instead of `8`.
- `Countdown` has to stay exhausted. Return `0, false` without touching the
  counter once it has gone below 1, or the fifth call starts counting into the
  negatives.
- `Multipliers` looks like the classic loop-variable trap. Since Go 1.22 the
  range variable is a new variable each iteration, so capturing it directly is
  correct now; in older Go every closure shared one variable and they all
  multiplied by the last element. Calling `Multiplier(n)` is unambiguous under
  either rule, which is why it is worth preferring.
