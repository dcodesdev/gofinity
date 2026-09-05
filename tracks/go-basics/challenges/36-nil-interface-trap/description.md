# The Nil Interface Trap

An interface value is a pair: a type, and a value of that type. It is `nil` only
when **both halves** are empty. Put a nil pointer into an interface and you get
a non-nil interface holding a nil pointer, which compares `!= nil` while
carrying nothing:

```go
var f *Failure         // f == nil
var err error = f      // err != nil
```

This is the single most reported "bug" in Go, and it is not a bug. `err` really
does know something: it knows the value is a `*Failure`. It just happens to be
the nil one.

It bites through return values:

```go
func Do() error {
	var f *Failure     // nil
	if somethingWentWrong() {
		f = &Failure{...}
	}
	return f           // always non-nil as an error
}
```

Every caller's `if err != nil` fires, on every call, forever.

## The fixes

**Return a bare `nil`.** Not a typed nil variable:

```go
func Do() error {
	if !somethingWentWrong() {
		return nil
	}
	return &Failure{...}
}
```

**Never declare the concrete pointer as the return variable.** If a helper
hands you a `*Failure`, guard the conversion at the boundary:

```go
func Wrap(f *Failure) error {
	if f == nil {
		return nil
	}
	return f
}
```

**Keep concrete types out of interface-typed variables.** The trap needs an
assignment from a concrete pointer to an interface. Where there is none, there
is no trap.

## Detecting it

A type assertion recovers the pointer, and once you have the pointer you can
compare it to `nil` on its own terms:

```go
f, ok := err.(*Failure)
if ok && f == nil { /* the trap */ }
```

[`reflect.ValueOf(err).IsNil()`](https://pkg.go.dev/reflect#ValueOf) does the
same for any pointer type, and if you find yourself needing it, the real answer
is usually to fix the function that produced the value.

## Nil receivers are fine

Calling a method on a nil pointer is legal. The method runs with a nil receiver
and only a dereference panics, so a nil-safe `Error()` is easy to write and
worth writing:

```go
func (f *Failure) Error() string {
	if f == nil {
		return "<no failure>"
	}
	return f.Msg
}
```

That is why the trap is quiet: the interface is non-nil, the method works, and
nothing crashes. It just reports a failure that never happened.

## Task

`Broken` is written for you and is wrong on purpose: read it, and read the test
that pins its behaviour down. Then write `Error`, `Fixed`, `Wrap`,
`HoldsNilPointer`, `Compact` and `FirstError`.

## Hints

- `Error` must check `f == nil` before touching `f.Msg`. The method call itself
  is safe; the dereference is not.
- `Fixed(false)` returns the literal `nil`. Assigning a nil `*Failure` to the
  result and returning that is exactly `Broken`.
- `HoldsNilPointer` needs both halves: `err != nil` first, then
  `f, ok := err.(*Failure)` and `f == nil`.
- `Compact` should build a new slice with `make([]error, 0, len(errs))`.
  Returning the input, filtered in place, would change the caller's slice.
- `FirstError` returns a bare `nil` when it finds nothing. Returning a
  `*Failure` variable that stayed nil recreates the trap in a new place.
