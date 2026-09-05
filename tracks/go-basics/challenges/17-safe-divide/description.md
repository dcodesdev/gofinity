# Safe Divide

Some failures in Go are not error values. Dividing by zero, indexing past the
end of a slice, calling a method on a nil pointer: those **panic**. A panic
unwinds the stack, running every deferred call on the way, and if nothing stops
it the program dies with a stack trace.

`recover` stops it:

```go
func safe() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
	}()
	risky()
	return nil
}
```

Three things make that work, and all three are easy to get wrong.

`recover` only does anything **inside a deferred function** of the call that is
panicking. Called anywhere else it returns `nil` and the panic carries on.

The value it returns is an `any` - whatever was handed to `panic`. A runtime
failure hands you a [`runtime.Error`](https://pkg.go.dev/runtime#Error); your
own code can hand over anything, and an `error` is usually the right choice
because the recoverer can pass it straight on.

And the recovery is only useful if the function can still report something,
which is why these functions use **named results**: the deferred closure sets
`err` after the `return` has already chosen its values.

Panicking is not Go's error handling. Return an `error` for anything a caller
could reasonably expect. Recover at a boundary, where letting one bad call kill
everything would be worse than reporting it.

## Task

Fill in the four functions in `main.go`.

1. `SafeDivide(a, b int) (int, error)` divides, and turns the divide-by-zero
   panic into `"safe divide: <what happened>"` with a result of `0`.
2. `MustPositive(n int) int` returns `n`, or panics with `ErrNotPositive` when
   `n` is zero or negative.
3. `SafeIndex(nums []int, i int) (int, error)` is the same shape as
   `SafeDivide`, for `"safe index: ..."`.
4. `Guard(body func()) error` runs `body` and converts a panic into an error:
   an `error` panic value comes back unchanged, anything else is formatted as
   `"panic: <value>"`.

## Hints

- The tests check that the message carries the *recovered* text, so an
  `if b == 0` guard that never divides will not pass. Let it panic and catch it.
- `if r := recover(); r != nil { ... }` is the whole idiom. Keep it in the
  deferred literal.
- Reset the value result as well as setting `err`. The `return a / b` never
  completed, but a partially-set result is exactly the plausible-looking answer
  a careless caller would use.
- `Guard` needs a type assertion, `r.(error)`, in its comma-ok form: `err, ok :=
  r.(error)`. Type assertions get a lesson of their own later; the comma-ok
  shape is the one you already know.
- [`errors.Is`](https://pkg.go.dev/errors#Is) in the tests compares identity, so
  panic with `ErrNotPositive` itself rather than a new error with the same
  text.
