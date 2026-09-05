# Variadic Sum

A **variadic** parameter accepts any number of arguments, including none. It is
written `...T` and has to be the last parameter:

```go
func Sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

Sum()           // 0
Sum(1, 2, 3)    // 6
```

Inside the function `nums` is an ordinary `[]int`. With no arguments it is
`nil`, which has length zero and ranges zero times, so the empty case usually
needs no code at all.

To pass a slice you already have, **spread** it with `...`:

```go
values := []int{1, 2, 3}
Sum(values...)      // 6
Sum(values)         // does not compile: []int is not an int
```

The two forms are not the same underneath. `Sum(1, 2, 3)` builds a fresh slice;
`Sum(values...)` passes `values` itself, so the parameter shares the caller's
backing array. A variadic function that writes to its parameter writes to the
caller's slice.

When you need "at least one", put a required parameter in front of the variadic:
`func Largest(first int, rest ...int) int` cannot be called with nothing, and the
compiler enforces it rather than a runtime check.

## Task

Fill in the six functions in `main.go`.

1. `Sum(nums ...int) int` adds its arguments; no arguments is `0`.
2. `Largest(first int, rest ...int) int` returns the biggest value.
3. `SumAll(groups ...[]int) int` adds every number in every group, by calling
   `Sum` with each group spread.
4. `Average(nums ...int) (float64, bool)` returns the mean, and `false` when
   there are no arguments.
5. `Describe(format string, args ...any) string` forwards to
   [`fmt.Sprintf`](https://pkg.go.dev/fmt#Sprintf).
6. `SumEvens(nums ...int) int` adds only the even arguments, leaving the
   caller's slice untouched.

## Hints

- `Sum()` needs no `if`. The parameter is `nil`, `range` over `nil` runs zero
  times, and the total stays `0`.
- `SumAll` is a variadic *of slices*. The outer call spreads groups in;
  `Sum(group...)` spreads each group back out.
- `Average` divides by `len(nums)`, which is why it has to check for zero first:
  integer or not, dividing by zero here is a panic.
- `float64(Sum(nums...) / len(nums))` is already wrong - it does the integer
  division and then converts the wrong answer. Convert both operands first.
- `Describe` must end in `args...`, not `args`. Passing the slice itself makes
  it one argument of type `[]any`, and `%s has %d` then prints `%!s(...)`
  noise instead of the values.
- The last test calls `SumEvens(nums...)` and checks `nums` afterwards. As long
  as you only read the parameter you are fine; a solution that sorts or
  compacts it in place will fail there and not in the sum.
