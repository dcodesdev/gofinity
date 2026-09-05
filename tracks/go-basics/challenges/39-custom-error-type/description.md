# Custom Error Type

A sentinel says *what went wrong*. Sometimes the caller also needs to know
*where*, or *which field*, or *how long to wait* - and a string cannot be asked
that. When the caller has to act on the detail, the error becomes a struct.

```go
type ParseError struct {
	Line int
	Key  string
	Err  error
}

func (e *ParseError) Error() string { ... }
func (e *ParseError) Unwrap() error { return e.Err }
```

[`error`](https://pkg.go.dev/builtin#error) is an interface with one method, so
anything with `Error() string` satisfies it. Add `Unwrap` and the sentinel
inside stays reachable, which means you get both:
[`errors.Is(err, ErrNotNumeric)`](https://pkg.go.dev/errors#Is) for the
condition and [`errors.As(err, &pe)`](https://pkg.go.dev/errors#As) for the
fields.

## Pointer receiver

Write the methods on `*ParseError`, not `ParseError`. Two reasons: the method
set of `T` does not include pointer-receiver methods, so mixing the two makes
"does not implement error" appear at a distance; and `errors.As` compares
concrete types, so a package that sometimes returns `ParseError` and sometimes
`*ParseError` forces every caller to check for both. Pick the pointer and be
consistent.

The pointer brings the nil interface trap with it, so return a **bare `nil`** on
success. `var e *ParseError; return e` is a non-nil error.

## errors.As

`errors.Is` asks "is this condition in the chain". `errors.As` asks "is a value
of this type in the chain", and hands it to you:

```go
var pe *ParseError
if errors.As(err, &pe) {
	fmt.Println(pe.Line)
}
```

The second argument is a pointer to the variable you want filled, so it is
`&pe` where `pe` is already a pointer type. It walks the chain like `errors.Is`
and stops at the first match.

## errors.Join

Validation rarely wants to stop at the first problem.
[`errors.Join`](https://pkg.go.dev/errors#Join) builds one error out of
several:

```go
return errors.Join(errs...)
```

Its message is each error on its own line. Its chain is a **tree**: instead of
`Unwrap() error` it has `Unwrap() []error`, and `errors.Is` and `errors.As`
search every branch. `errors.Join` of nothing, or only nils, is nil, so the
common `if len(errs) > 0` guard is belt and braces.

Walking that tree by hand means handling both shapes:

```go
switch u := err.(type) {
case interface{ Unwrap() []error }:
	for _, child := range u.Unwrap() { walk(child) }
case interface{ Unwrap() error }:
	walk(u.Unwrap())
}
```

An anonymous interface in a `case` is unusual-looking and exactly right here:
the thing you care about is the method, not the type that has it.

## When not to

A custom type is API surface, and every exported field is a promise. If nobody
branches on the detail, [`fmt.Errorf`](https://pkg.go.dev/fmt#Errorf) with `%w`
is the whole answer.

## Task

Write `Error`, `Unwrap`, `ParseLine`, `ParseAll`, `LineOf`, `Lines` and
`Explain`. The struct and the three sentinels are given.

## Hints

- `Error` has two shapes: with a key and without. `%q` puts the quotes on the
  key, `%v` prints the wrapped sentinel.
- [`strings.Cut(line, "=")`](https://pkg.go.dev/strings#Cut) gives you the key,
  the value and whether the separator was there. You will need to add the
  import.
- A missing `=` **and** an empty key are both `ErrMalformed`. Only a present key
  with an empty value is `ErrMissingValue`.
- The digit check is the one from the last challenge: reject anything outside
  `'0'`-`'9'`, so `-1` is `ErrNotNumeric`.
- `ParseAll` numbers lines from 1, skips the empty ones, and appends to an
  `[]error` instead of returning. On failure return `nil` for the map.
- `LineOf` is four lines with `errors.As`. `Lines` cannot use it - `errors.As`
  stops at the first match - so write a recursive walk over the two `Unwrap`
  shapes, checking `err.(*ParseError)` first at each node.
- `Explain` checks the sentinels in the documented order, because a join can
  match more than one.
