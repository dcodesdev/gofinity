# Multiple Returns

A Go function can return more than one value, and that single fact removes a
whole category of design from the language. There are no out parameters, no
"return -1 for not found and hope nobody stores it", and no exceptions - a
function that might not have an answer returns the answer *and* a `bool` saying
whether it means anything.

```go
func Divide(a, b int) (int, bool) {
	if b == 0 {
		return 0, false
	}
	return a / b, true
}

q, ok := Divide(7, 2)
```

That shape is called **comma-ok**, and you will see it everywhere: map lookups,
type assertions, channel receives,
[`strings.Cut`](https://pkg.go.dev/strings#Cut). When `ok` is false the other
value is the zero value, never a half-answer.

Results can also be **named**, which declares them as ordinary variables, zeroed
on entry:

```go
func MinMax(nums []int) (lo, hi int, ok bool) {
	if len(nums) == 0 {
		return   // 0, 0, false
	}
	...
}
```

A bare `return` hands back whatever the named results currently hold. It is
worth using for an early exit like that one; a bare `return` at the end of a
long function, where the reader has to scroll to find what is being returned,
is not.

## Task

Fill in the five functions in `main.go`.

1. `Divide(a, b int) (int, bool)` returns `a / b`, or `0, false` when `b` is
   zero.
2. `MinMax(nums []int) (lo, hi int, ok bool)` returns the smallest and largest
   values. An empty or `nil` slice is `0, 0, false`. It must not reorder its
   argument.
3. `MaxOnly(nums []int) int` returns the largest value, or `0` when there is
   none, by calling `MinMax` and discarding the minimum.
4. `SplitName(full string) (first, last string)` splits at the **first** space:
   `"Grace Brewster Hopper"` is `"Grace"`, `"Brewster Hopper"`. A name with no
   space is all first name.
5. `Stats(nums []int) (count, sum int, mean float64)` counts, sums and averages.
   An empty slice returns three zeros and divides nothing.

## Hints

- Guard before you divide. `Divide` and `Stats` both have to notice the empty
  case *first*, because Go panics on integer division by zero rather than
  giving you an infinity.
- `mean` is a `float64` but `sum` and `count` are `int`s. `sum / count` is
  integer division, and converting afterwards - `float64(sum / count)` - has
  already thrown the fraction away. Convert both operands, then divide.
- `MinMax` starts `lo` and `hi` at `nums[0]`, not at `0`. Starting at zero
  reports `0` as the minimum of `[3 1 4]`.
- Do not sort the slice to find the ends.
  [`sort.Ints`](https://pkg.go.dev/sort#Ints) would reorder the caller's data,
  and one test checks that you left it alone.
- `_` is the blank identifier: it discards a result. `MaxOnly` wants
  `_, hi, ok := MinMax(nums)`, because Go refuses to compile a normal variable
  that is never read.
- `strings.Cut(s, sep)` returns `before, after, found`. It is exactly
  `SplitName`, including the no-separator case: `before` is the whole string and
  `after` is empty.
