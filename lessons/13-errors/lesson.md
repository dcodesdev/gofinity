# Errors

The last lesson ended on `error`, the one-method interface, and the trap where a
nil pointer inside one is not nil. This lesson is what that interface is
actually for.

Go has no exceptions. There is no `throw`, no `catch`, and no invisible unwind.
A function that can fail says so in its type:

```go
func Divide(a, b int) (int, error)
```

and the caller decides, in the open, every time:

```go
q, err := Divide(84, 7)
if err != nil {
	return err
}
```

People complain about that block. It is three lines, it is everywhere, and it is
the price of a language where you can read a function top to bottom and know
which lines can fail. An exception can be thrown from anywhere and caught
nowhere; a returned error is in the signature.

## Errors are values

That is the whole idea, and everything else is a consequence. An error is an
ordinary value of an ordinary interface type:

```go
type error interface{ Error() string }
```

So you can store one in a struct, put it in a slice, send it down a channel,
compare it, or write a function that takes one. Nothing about error handling is
special machinery. Later in this track you will pass errors between goroutines
over channels, and nothing new will be needed to do it.

## The pairing rule

When the error is non-nil, the other return values are the zero value and mean
nothing. When it is nil, they are real.

```go
return 0, err        // yes
return partial, err  // no
```

Callers depend on this so completely that breaking it produces bugs nobody
thinks to look for. Return the zero value with the error.

The one famous exception is `io.Reader`, which may return both bytes read and an
error, and every one of its docs says so loudly. If you need to break the rule,
you also need to document it.

## Creating one

```go
errors.New("empty input")
fmt.Errorf("divide %d by zero", a)
```

`errors.New` for a fixed message, `fmt.Errorf` when values belong in it. Reach
for `fmt.Errorf` by default: an error that does not say which input failed sends
the reader to a debugger.

House style, and it is near-universal in Go code: **lowercase, no trailing
punctuation, no "failed to"**. Errors get concatenated into each other, and a
capital letter or a period in the middle of a chain reads as a mistake.

## Sentinels

Some failures are not a message, they are a condition the caller wants to branch
on: not found, end of input, already exists. Those get a package-level value:

```go
var ErrNotFound = errors.New("not found")
```

Once, at package level, exported, named `Err...`. Once is what makes it work:
`errors.New` returns a fresh pointer each call, so two errors with identical
text are still different values, and a sentinel built inside a function is
comparable to nothing.

`io.EOF` is the one everybody meets first, and it is a good model: a condition
so ordinary that treating it as a message would be wrong.

A sentinel is API. Once callers branch on it you cannot rename or remove it
without breaking them, so promote a condition to a sentinel only when there is
something useful to do about it. "The database is unreachable" is a message.
"This key does not exist" is a sentinel.

## Adding context: %w

An error that travels three functions up and still says `not found` is useless:
what was not found, and who was looking? `fmt.Errorf` has one verb the others do
not:

```go
fmt.Errorf("score %q: %w", name, err)
```

`%w` formats the error like `%v` **and** records it, so the result keeps a
pointer to the original. Layers stack:

```
score "bob": parse "not-a-number": invalid
```

Each function adds its own piece and a `: %w` at the end, so the finished
message reads outside-in like a path. `%v` produces byte-identical text and no
chain at all - the difference is invisible in the output, which is why it is
worth being deliberate about.

Wrapping is a choice in both directions. What you wrap with `%w` becomes part of
your API, because callers can now match against it. When the inner error is an
implementation detail you might change - a driver's error, a library's - use
`%v` on purpose and let the chain stop there.

## errors.Is

A wrapped error equals nothing:

```go
err == ErrNotFound          // false, it is a *fmt.wrapError now
errors.Is(err, ErrNotFound) // true
```

`errors.Is` walks the chain, comparing at every level. Use it from the first
line you write, even when nothing is wrapped yet, because it is the only
comparison that survives somebody adding context later.

`errors.Unwrap(err)` takes one step down, or returns nil at the bottom. You will
rarely call it: it is the mechanism `errors.Is` is built on, not the interface
you want.

## Custom types, and errors.As

A sentinel says what went wrong. Sometimes the caller also needs to know where,
or which field, or how long to wait - and a string cannot be asked that. When
the caller has to act on the detail, the error becomes a struct:

```go
type ParseError struct {
	Line int
	Key  string
	Err  error
}

func (e *ParseError) Error() string { ... }
func (e *ParseError) Unwrap() error { return e.Err }
```

`Unwrap` is what keeps the sentinel inside reachable, so you get both:
`errors.Is` for the condition, `errors.As` for the fields.

```go
var pe *ParseError
if errors.As(err, &pe) {
	fmt.Println(pe.Line)
}
```

`errors.As` walks the chain looking for a value of a given concrete type and
assigns it. The argument is a pointer to the variable you want filled, so it is
`&pe` where `pe` is itself a pointer.

Two rules keep custom error types out of trouble:

- **Pointer receiver, always.** `*T`'s method set has everything, `T`'s does
  not, and `errors.As` matches on the exact concrete type - so a package that
  returns `ParseError` sometimes and `*ParseError` other times forces every
  caller to check twice.
- **Return a bare `nil` on success.** The pointer brings last lesson's trap with
  it: `var e *ParseError; return e` is a non-nil error.

And the negative rule: a custom type is API surface, and every exported field is
a promise. If nobody branches on the detail, `fmt.Errorf` with `%w` is the whole
answer.

## Several at once

Validation rarely wants to stop at the first problem. `errors.Join` makes one
error out of many:

```go
return errors.Join(errs...)
```

Its message is each error on its own line, and its chain is a **tree**: instead
of `Unwrap() error` it has `Unwrap() []error`. `errors.Is` and `errors.As`
search every branch. `errors.Join` of nothing, or only nils, is nil.

To walk that tree yourself you handle both shapes, and an anonymous interface in
a type switch is the idiomatic way to say "whatever has this method":

```go
switch u := err.(type) {
case interface{ Unwrap() []error }:
	for _, child := range u.Unwrap() { walk(child) }
case interface{ Unwrap() error }:
	walk(u.Unwrap())
}
```

## Handle once

The most common real mistake is not a missing check, it is a doubled one: log
the error *and* return it, at every level, so one failure produces six lines of
log saying the same thing. Handle an error exactly once. Everywhere else, add
context and return it. The place that finally decides what to do - the HTTP
handler, `main`, the retry loop - is the place that logs.

And `if err != nil { return err }` with nothing added is fine. Wrap when you
have context worth having, not as a reflex.

## Further reading

- [errors](https://pkg.go.dev/errors) - `New`, `Is`, `As`, `Join` and `Unwrap`,
  with the exact rules each one follows down a chain.
- [Errors are values](https://go.dev/blog/errors-are-values) - the short post
  the whole approach is named after.
- [Error handling and Go](https://go.dev/blog/error-handling-and-go) - the
  interface, sentinels, and custom error types.
- [Working with errors in Go 1.13](https://go.dev/blog/go1.13-errors) - where
  `%w`, `errors.Is` and `errors.As` came from, and when to use each.
- [fmt](https://pkg.go.dev/fmt) - `Errorf`, and the one paragraph defining what
  `%w` does that `%v` does not.

## Practise

Three challenges. The first is the shape itself: sentinels declared once,
`fmt.Errorf` for the rest, the zero value beside every error, and a `Lookup`
that has to use the comma-ok form so a stored zero is not mistaken for a miss.
The second is wrapping: `%w`, `errors.Is` on a chain two deep, and walking the
chain by hand with `errors.Unwrap` to see what it really is. The third builds a
`*ParseError` with an `Unwrap`, finds it with `errors.As`, collects a file's
worth of problems with `errors.Join`, and walks the resulting tree over both
`Unwrap` shapes.
