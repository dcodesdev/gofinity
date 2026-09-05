# Temperature Conversion

Two scales, one formula each, and everything that goes wrong when integers and
floats meet:

```text
F = C * 9/5 + 32
C = (F - 32) * 5/9
```

Write `c * 9 / 5` in Go and it works, because `c` is a `float64` and the untyped
constants adopt its type. Write `9 / 5` on its own and you get `1`: two untyped
integer constants divide as integers. The habit that saves you is writing the
constants as `9.0` and `5.0` so the intent is on the page rather than inferred
from a neighbouring variable.

Rounding is the other half. `int(f)` truncates toward zero, so `int(70.7)` is
`70` and `int(-3.6)` is `-3`. When you want the *nearest* value you say so with
[`math.Round`](https://pkg.go.dev/math#Round), which rounds halves away from
zero, and only then convert.

## Task

Finish the five functions in `main.go`.

1. `CToF(c float64) float64` and `FToC(f float64) float64` - the two formulas.
   `-40` is the same in both scales, and the tests check that converting there
   and back lands where it started.
2. `RoundTenth(v float64) float64` rounds to one decimal place, halves away
   from zero: `1.25` becomes `1.3` and `-1.25` becomes `-1.3`.
3. `FahrenheitWhole(c float64) int` converts to Fahrenheit and rounds to the
   nearest whole degree. `21.5°C` is `70.7°F`, so the answer is `71`, not `70`.
4. `Report(c float64) string` renders one line with both scales to one decimal:
   `Report(21.5)` is `"21.5°C = 70.7°F"`.

## Hints

- `math.Round` is in the `math` package, which `main.go` does not import yet.
  Add it; an unused import is a compile error, which is why it was left out.
- Rounding to a decimal place is rounding after a shift:
  `math.Round(v*10) / 10`. The same trick with `100` gives you two places.
- `FahrenheitWhole` is `math.Round` **then** `int(...)`. Reverse those two and
  the conversion truncates before the rounding ever happens.
- `%.1f` in `fmt.Sprintf` prints one decimal place and rounds for you, so
  `Report` does not need `RoundTenth`. Rounding a value and rounding its
  *display* are different jobs; this challenge asks for both so the difference
  has somewhere to show.
- The degree sign in the expected output is `°`, one rune and two bytes in
  UTF-8. Paste it rather than typing a lookalike, and note that `len("°C")` is
  `3` - exactly what the previous challenge was about.
- The tests compare floats with a tolerance instead of `==`. `0.1 + 0.2` is not
  `0.3` in binary floating point, in Go or anywhere else, so an exact
  comparison on computed floats is a bug waiting for the right input.
