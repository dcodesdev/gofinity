# The Shape Interface

An interface in Go is a list of method signatures:

```go
type Shape interface {
	Area() float64
	Perimeter() float64
}
```

A type satisfies it by having those methods. There is no `implements`, no
registration, and nothing in `Rect` mentions `Shape`. The compiler checks the
match where the value is used - assigned to a `Shape` variable, passed to a
`Shape` parameter, appended to a `[]Shape` - and that is the only place the
relationship exists.

That inversion matters: the interface belongs to the code that *consumes* it,
not to the types that satisfy it. You can write an interface for a type you
did not write, in your own package, after the fact.

## Two values, one variable

An interface value holds two things: the concrete type, and the value of that
type. `var s Shape = Rect{2, 3}` records "this is a `Rect`" beside the `Rect`
itself. Calling `s.Area()` looks up `Rect`'s `Area` and runs it, so the same
line of code does different work depending on what is in there.

The zero value of an interface is `nil`: no type, no value. Calling a method on
it panics, so a function that returns "nothing" as an interface returns `nil`
and callers check for it.

## Asking what is inside

A type assertion recovers the concrete value, and the comma-ok form asks rather
than insists:

```go
r, ok := s.(Rect)      // ok is false if s does not hold a Rect
```

The same syntax works with another **interface** on the right, which asks
whether the value has that method set as well:

```go
if n, ok := s.(Named); ok {
	fmt.Println(n.Name())
}
```

Without `, ok`, a failed assertion panics. Use the two-value form unless you can
prove the assertion holds.

## Small interfaces

The interfaces you meet in the standard library are one or two methods:
[`io.Reader`](https://pkg.go.dev/io#Reader),
[`io.Writer`](https://pkg.go.dev/io#Writer),
[`fmt.Stringer`](https://pkg.go.dev/fmt#Stringer), `error`. Small interfaces are
easy to satisfy by accident, which is exactly the point - a type written years
ago fits a function written today. Start with the smallest set of methods your
function actually calls.

## Task

Implement `Area` and `Perimeter` on `Rect` and `Circle`, `Name` on `Circle`,
then the four functions that work through the interface: `TotalArea`,
`Largest`, `Describe` and `CountAtLeast`.

## Hints

- [`math.Pi`](https://pkg.go.dev/math#pkg-constants) is the constant;
  `Circle{R: 2}` has an area of `4 * math.Pi`.
- `Largest` starts with `var best Shape`. The zero value of an interface is
  `nil`, so an empty slice returns `nil` without a special case.
- `Describe` needs `if n, ok := s.(Named); ok` - a plain `s.(Named)` would
  panic on a `Rect`, which has no `Name` method.
- `%.2f` formats both numbers. `Circle{R: 1}` describes as
  `circle: area 3.14, perimeter 6.28`.
- `CountAtLeast` is inclusive: an area exactly equal to `min` counts.
