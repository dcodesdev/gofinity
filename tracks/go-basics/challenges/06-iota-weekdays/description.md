# Weekdays with iota

An enumeration in Go is a named integer type plus a block of constants. There
is no `enum` keyword, and none is needed: `iota` and a defined type do the job.

`iota` is a counter that resets to `0` at the start of every `const` block and
increases by one per line. Inside a block, a constant line with no expression
repeats the previous one, so this declares three constants and not one:

```go
type Level int

const (
	Debug Level = iota // 0
	Info               // 1
	Warn               // 2
)
```

Giving the type a `String() string` method makes it print as a name instead of
a number, everywhere: `fmt` looks for that method on any value it formats.

## Task

Finish `main.go`.

1. Replace the placeholder constants with **one `iota` block** declaring
   `Sunday` through `Saturday` as `Weekday`, numbered `0` to `6`.
2. `String() string` returns the day's English name. A value outside `0..6` is
   not a day, and formats as `Unknown(<n>)`: `Weekday(9).String()` is
   `"Unknown(9)"`, and `Weekday(-1)` is `"Unknown(-1)"`.
3. `IsWeekend() bool` is true for Saturday and Sunday only.
4. `Next() Weekday` returns the following day and wraps `Saturday` back to
   `Sunday`. A value outside `0..6` has no following day and comes back
   unchanged.

The tests check the numbers, the names, the wrap, and that `fmt.Sprintf("%v",
Friday)` prints `Friday` rather than `5`.

## Hints

- `Sunday Weekday = iota` on the first line is what gives the whole block its
  type. The lines after it need no expression at all.
- A method is a function with a receiver in front of the name:

  ```go
  func (d Weekday) IsWeekend() bool { ... }
  ```

  `d` is the value the method was called on, exactly like `this` elsewhere, and
  you pick its name. You will meet receivers properly in the methods concept;
  here you only need this one form.
- The names are easiest as a package-level array indexed by the day:

  ```go
  var dayNames = [...]string{"Sunday", "Monday", ...}
  ```

  `[...]` sizes the array from the literal. Check the range **before** you
  index it: `dayNames[9]` panics, and the tests pass `Weekday(9)` on purpose.
- Write `int(d)` when formatting the unknown case. `fmt.Sprintf("Unknown(%d)",
  d)` would call `String` again and recurse until the stack runs out. That is
  the classic `Stringer` trap, and worth meeting once.
- `d + 1` is a `Weekday`, not an `int`: arithmetic on a defined type stays in
  that type. That is exactly why the enum is a type of its own rather than a
  bare `int`.
