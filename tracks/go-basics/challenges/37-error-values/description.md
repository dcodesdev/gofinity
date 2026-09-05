# Error Values

Go has no exceptions. A function that can fail says so in its signature, by
returning an `error` as its last result:

```go
func Divide(a, b int) (int, error)
```

[`error`](https://pkg.go.dev/builtin#error) is an ordinary interface with one
method:

```go
type error interface{ Error() string }
```

Nothing is thrown, nothing unwinds, and nothing happens behind your back. The
call site decides, every time:

```go
q, err := Divide(84, 7)
if err != nil {
	return err
}
```

That block is the most-typed thing in Go, and it is the point: every failure is
visible in the code that has to handle it.

## The pairing rule

When the error is non-nil, the other results are the **zero value** and mean
nothing. When it is nil, they are real. Callers rely on that, so a function
that returns a half-computed number beside an error is lying:

```go
return 0, err        // yes
return partial, err  // no
```

## Making an error

Two builders cover almost everything:

```go
errors.New("empty input")                 // a fixed message
fmt.Errorf("divide %d by zero", a)        // a message with values in it
```

[`fmt.Errorf`](https://pkg.go.dev/fmt#Errorf) is the one you reach for inside a
function, because an error that does not say *which* input failed sends the
reader back to the debugger.

## Sentinels

Some failures are not a message, they are a **condition** the caller wants to
branch on: no such key, end of input, already exists. Those get a package-level
value:

```go
var ErrNotFound = errors.New("not found")
```

Declared once, at package level, exported, named `Err...`. Once is what makes it
work: [`errors.New`](https://pkg.go.dev/errors#New) returns a new pointer each
call, so two errors with identical text are still different values. A sentinel
built inside the function is comparable to nothing.

Callers compare with [`errors.Is`](https://pkg.go.dev/errors#Is):

```go
if errors.Is(err, ErrNotFound) {
	// ...
}
```

`err == ErrNotFound` also works today, but `errors.Is` keeps working when
somebody later wraps the error on its way up. Use it from the start; the next
challenge is about the wrapping.

A sentinel is API. Once callers branch on it you cannot rename or remove it
without breaking them, so only promote a condition to a sentinel when there is
something useful to do about it.

## Passing one up

Most functions do not handle an error, they return it:

```go
q, err := Divide(total, d)
if err != nil {
	return 0, err
}
```

Returning `err` unchanged is fine when your function adds no context worth
having. When it does, wrap it - which is the next challenge.

## Task

Add the missing sentinel, then write `Divide`, `First`, `Lookup`,
`SumQuotients` and `Describe`. The starter does not compile until `ErrNotFound`
exists.

## Hints

- `ErrNotFound` goes beside `ErrEmpty`, at package level, with `errors.New`.
- `First` returns `ErrEmpty` itself. `errors.New("empty input")` inside the
  function has the same text and fails the test, which is the lesson.
- `Lookup` needs the comma-ok form, `v, ok := m[key]`. A missing key and a
  stored `0` look identical otherwise, and one of the cases checks exactly that.
  A nil map is safe to read from, so it needs no special case.
- `SumQuotients` returns `0, err` on the first failure, not the running total.
- `Describe` checks the error first and only formats the value when there is
  none.
