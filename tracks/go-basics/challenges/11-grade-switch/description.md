# Grade Switch

Go's `switch` is the same keyword you have seen elsewhere with two of its worst
habits removed. Cases do **not** fall through, so no `break` is needed and
forgetting one cannot cost you an afternoon. And a case can list several values,
which turns a stack of near-identical branches into one line.

```go
switch day {
case "sat", "sun":
	return "weekend"
default:
	return "unknown"
}
```

Leave the expression out and the switch compares against `true`, so every case
is a condition of its own. That is the form you want for ranges:

```go
switch {
case score >= 90:
	return "A"
case score >= 80:
	return "B"
}
```

Falling through is still possible when you actually want it, with an explicit
`fallthrough` as the last statement in a case. It jumps to the next case's body
without testing its condition.

## Task

Finish the four functions in `main.go`.

1. `Grade(score int) string` maps a percentage to a letter: 90+ is `"A"`, then
   `"B"`, `"C"`, `"D"` every ten points, and below 60 is `"F"`. A score outside
   `0..100` is not a score at all: return `"?"`.
2. `DayKind(day string) string` answers `"weekend"` for `"sat"` and `"sun"`,
   `"weekday"` for the other five abbreviations, and `"unknown"` for anything
   else, including `"Mon"` with its capital.
3. `Perks(tier string) []string` lists a membership tier's perks, most
   exclusive first. Each tier includes everything below it: gold gets
   `["lounge", "priority", "points"]`, silver drops the lounge, bronze is just
   points. An unrecognised tier gets nothing.
4. `Season(month int) string` names the meteorological season of a month number,
   with December, January and February together in `"winter"`, and anything
   outside `1..12` `"unknown"`.

## Hints

- `Grade` wants the expressionless form, and the out-of-range case has to come
  first. Put it last and `101` matches `score >= 90` on the way past.
- The cases in an expressionless switch are tested top to bottom, exactly like
  an if/else if chain, so `score >= 80` only ever sees scores under 90.
- `DayKind` and `Season` want the value form with several values per case. No
  `break` anywhere; a Go case ends where the next one begins.
- `Perks` is the `fallthrough` exercise. Three cases, each appending its own
  perk and then falling through to the one below, and no `default` at all so an
  unknown tier leaves the slice untouched.
- `fallthrough` must be the last statement in its case, and it cannot appear in
  the final case. It ignores the next case's condition entirely, which is why it
  is rare and why Go makes you write it.
- Appending to a `var perks []string` that is still `nil` works and returns a
  slice of one. A `nil` slice has length zero and ranges over nothing, so an
  unknown tier needs no special handling.
