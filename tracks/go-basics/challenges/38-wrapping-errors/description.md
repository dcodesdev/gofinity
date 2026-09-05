# Wrapping Errors

An error that travels three functions up and still says `not found` is useless:
what was not found, and who was looking? Wrapping adds that context on the way
out without destroying the error underneath.

## %w

[`fmt.Errorf`](https://pkg.go.dev/fmt#Errorf) has one verb the others do not:

```go
fmt.Errorf("score %q: %w", name, err)
```

`%w` formats `err` like `%v` **and** records it, so the new error keeps a
pointer to the old one. The result is a chain:

```
score "bob": parse "not-a-number": invalid
└─ parse "not-a-number": invalid
   └─ invalid
```

`%v` produces the identical text and no chain at all. That is the whole
difference, and it is invisible in the output, which is why it is worth being
deliberate about.

## The house style for the message

Lowercase, no trailing punctuation, and the context you are adding followed by
`: %w` at the end. Each layer prepends its own piece, so the finished message
reads outside-in like a path. Do not write "failed to" or "error while" - the
reader already knows it is an error, and by the third layer the message is
mostly apology.

## errors.Is

Once an error is wrapped it equals nothing:

```go
err == ErrNotFound          // false, it is a *fmt.wrapError now
errors.Is(err, ErrNotFound) // true
```

[`errors.Is`](https://pkg.go.dev/errors#Is) walks the chain, comparing at every
level. That is why the previous lesson said to use it from the start: it is the
only comparison that survives a caller deciding to add context later.

## errors.Unwrap

One step down the chain, or `nil` at the bottom:

```go
inner := errors.Unwrap(err)
```

You rarely call [`errors.Unwrap`](https://pkg.go.dev/errors#Unwrap) directly -
`errors.Is` and [`errors.As`](https://pkg.go.dev/errors#As) are the interface
you want - but it is what they are built on, and walking it by hand once makes
the chain concrete.

Wrapping is opt-in in the other direction too. An error you wrap with `%w`
becomes part of your API: callers can now match against whatever is inside it.
When the inner error is an implementation detail you may change, format it with
`%v` on purpose and the chain stops there.

## Multiple errors

`%w` can appear more than once, and
[`errors.Join(err1, err2)`](https://pkg.go.dev/errors#Join) builds a value
holding several. `errors.Is` searches all the branches. That is the tool for
"validate everything and report all the problems", rather than stopping at the
first.

## Task

Write `parseScore`, `Score`, `Reason`, `Unwrapped`, `Depth` and `Total`.
`fetch` and the two sentinels are given. Every message in the tests is exact,
including the quotes, and every chain has to stay walkable.

## Hints

- `%q` prints a Go-quoted string: `parse "not-a-number"`. `%s` would drop the
  quotes and fail the comparison.
- `Score` wraps both failures the same way, so the two `if err != nil` blocks
  are identical apart from where the error came from.
- `parseScore` needs no `strconv`: reject an empty string, then walk the runes
  and reject anything outside `'0'`-`'9'`, building `n = n*10 + int(r-'0')` as
  you go. `-1` fails on the `-`, which is what the test expects.
- `Reason` is a `switch` with no expression: `case err == nil`, then two
  `errors.Is` cases, then a default.
- `Unwrapped` loops until `errors.Unwrap` returns nil and gives back the last
  non-nil error it saw. `errors.Unwrap(nil)` is nil, not a panic, so the nil
  case falls out for free.
- `Total` wraps what `Score` already wrapped, which is why its chain is two
  deep, not one.
