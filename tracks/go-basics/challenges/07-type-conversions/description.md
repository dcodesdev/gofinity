# Type Conversions

Go has no implicit numeric conversion. An `int` is not an `int64`, and neither
is a `float64`. If you want a value in another type you say so, every time:

```go
var n int = 7
var f float64 = float64(n)  // float64(n), never just n
```

That looks like noise until you meet the bug it prevents. In a language that
converts silently, `sum / count` quietly gives you an integer division, or a
number too large for its new type wraps around without a word. Go makes both of
those visible: the first will not compile, and the second is spelled out in the
conversion you had to write.

Two things a conversion does that are worth knowing before you write one:

- `int(f)` on a float **truncates toward zero**. It does not round.
  `int(2.9)` is `2`, and `int(-2.9)` is `-2`.
- `int8(n)` on a value too big for `int8` **wraps**, silently, using the low
  bits. `int8(200)` is `-56`. Nothing panics. If the value has to survive the
  trip, you check the range yourself.

## Task

Finish the three functions in `main.go`.

1. `Truncate(f float64) int` drops the fractional part toward zero.
2. `Average(nums []int) float64` returns the mean. An empty slice averages to
   `0`. Watch the division: the whole point of this one is that the sum and the
   count are both `int`.
3. `FitsInt8(n int) (int8, bool)` returns `int8(n)` and whether that conversion
   kept the value. Out of range, return the wrapped result Go gives you and
   `false` - the caller wants to see what happened, not a zero.

## Hints

- [`math.MinInt8` and `math.MaxInt8`](https://pkg.go.dev/math#pkg-constants)
  are the bounds, and they read better than `-128` and `127`. `main.go` does
  not import `math` yet, so add it: an import you do not use is a compile error
  in Go, which is why it was left out.
- Convert **before** you divide, not after. `float64(sum / len(nums))` does the
  integer division first and then converts a result that already lost its
  fraction. `float64(sum) / float64(len(nums))` is the one you want.
- `Average(nil)` must return `0` rather than dividing by zero. Integer division
  by zero panics; float division by zero would give you `NaN`. Neither is an
  average.
- The tests compare floats with a tolerance rather than `==`, because binary
  floating point rarely lands on an exact decimal. Do the same in your own code
  when you get there.
