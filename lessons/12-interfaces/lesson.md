# Interfaces

Methods attached behaviour to one type. Interfaces are how code stops caring
which type it got. Go's version is unusual in one respect, and everything worth
knowing follows from it: **satisfaction is implicit**.

## No implements keyword

An interface is a list of method signatures:

```go
type Shape interface {
	Area() float64
	Perimeter() float64
}
```

A type satisfies it by having those methods. `Rect` does not mention `Shape`,
`Shape` does not mention `Rect`, and there is no registration step. The
compiler checks the match at the point of use - an assignment, an argument, an
append to a `[]Shape` - and that is the only place the relationship exists at
all.

The consequence is that **the interface belongs to the consumer**. In a
language with `implements`, the author of a type decides which interfaces it
fits, and must have thought of yours. In Go you write the interface in the
package that needs it, listing only the methods you actually call, and every
type that already has them fits. That includes types from libraries written
before your interface existed.

So the Go idiom runs backwards from what other languages teach: do not start
with an interface and write types to fit it. Write the concrete types, and
introduce an interface when a second implementation, or a test double, gives
you something to abstract over. "Accept interfaces, return structs" is the
short form - a function's parameters say what it needs, and its return type
stays concrete so callers keep every method.

## Small is the point

The interfaces that carry the standard library are one method:

```go
type Stringer interface{ String() string }
type Reader   interface{ Read(p []byte) (n int, err error) }
type Writer   interface{ Write(p []byte) (n int, err error) }
type error    interface{ Error() string }
```

`io.Reader` is satisfied by files, network connections, `strings.Reader`,
`bytes.Buffer`, gzip streams, HTTP bodies, and anything you write this
afternoon. That reach comes from having exactly one method. A four-method
interface is satisfied by whatever you designed for it and nothing else.

Larger interfaces get built by embedding rather than by growing:

```go
type ReadWriter interface {
	Reader
	Writer
}
```

## What an interface value is

An interface value is a pair: a **type** and a **value** of that type.

```go
var s Shape = Rect{W: 2, H: 3}   // type Rect, value {2 3}
```

The call `s.Area()` looks up `Rect`'s `Area` and runs it, so the same line does
different work depending on what is inside. The zero value of an interface is
`nil`: no type, no value. Calling a method on it panics.

Two interface values are equal when both halves are equal - and comparing them
**panics** if the concrete type is not comparable, which is how a `==` between
two interfaces holding slices blows up at run time rather than at compile time.

## Getting the concrete value back

A type assertion, in the comma-ok form that asks instead of insisting:

```go
r, ok := s.(Rect)      // ok is false when s holds something else
```

Without `, ok` a failed assertion panics. The same syntax with an interface on
the right asks whether the value has that method set as well, which is how you
probe for an optional capability:

```go
if n, ok := s.(Named); ok {
	fmt.Println(n.Name())
}
```

For several possibilities, a type switch:

```go
switch v := v.(type) {
case nil:
	// the interface held nothing
case int:
	// v is an int
case fmt.Stringer:
	// anything with a String method
default:
	// v is still the interface
}
```

Two rules earn their keep here. **Order matters**: an interface case matches
everything satisfying it, so concrete cases go first. And **a named type is its
own type**: `type Celsius float64` matches `case Celsius`, never `case
float64`. That is the same no-implicit-conversion rule you met with numbers,
showing up in a new place.

When you find yourself type-switching over *your own* types, that is usually a
method wanting to be written. A switch over types from elsewhere - `any` out of
a JSON decode, an error from a library - is fine and normal.

## Method sets, again

The rule from the last lesson is what you will actually hit:

- the method set of `*T` holds every method
- the method set of `T` holds only the value-receiver ones

```go
type Adder interface{ Add(int) }

var _ Adder = (*Counter)(nil)   // fine
var _ Adder = Counter{}         // does not compile
```

The error message - "does not implement Adder (method Add has pointer
receiver)" - is worth recognising on sight. The fix is almost always `&`, not a
change of receiver.

`var _ Shape = Rect{}` at package level is the standard way to assert
satisfaction where you define the type, so the failure lands in the file that
caused it rather than in some distant caller. It costs nothing at run time.

## The empty interface

`interface{}` has no methods, so **everything** satisfies it. `any` is an alias
for it, added in Go 1.18, and is what you should write.

`any` says nothing, and that is its problem: getting a usable value back out
takes an assertion or a type switch, checked at run time. It is right for
genuinely heterogeneous data - `fmt.Println`'s arguments, a decoded JSON
document, a cache of unrelated things - and wrong as a way to avoid deciding on
a type. Since generics, most of the old `any` signatures have a better version
with a type parameter, which keeps the checking at compile time. Generics are
two lessons away.

## The nil interface trap

The one Go surprise that everybody meets. An interface is nil only when **both
halves** are nil. A nil pointer stored in an interface is a non-nil interface:

```go
var f *Failure         // f == nil
var err error = f      // err != nil
```

In a return statement it becomes a bug that fires on every call:

```go
func Do() error {
	var f *Failure
	if wentWrong() {
		f = &Failure{...}
	}
	return f          // never nil as an error
}
```

Every caller's `if err != nil` is true, always. And it is quiet, because a
method on a nil pointer runs perfectly well - only a dereference panics - so
nothing crashes to tell you.

Three habits keep it away:

1. Return a **bare `nil`**, never a typed nil variable.
2. Guard at the boundary where a concrete pointer becomes an interface:
   `if f == nil { return nil }`.
3. Do not declare interface-typed variables to hold concrete pointers on their
   way somewhere.

To detect one after the fact, assert back to the pointer and compare:
`f, ok := err.(*Failure); ok && f == nil`.

## Further reading

- [Interface types](https://go.dev/ref/spec#Interface_types) - implicit
  satisfaction, embedding, and what a type set is doing in there.
- [Effective Go: interfaces and methods](https://go.dev/doc/effective_go#interfaces):
  why the interface belongs to the consumer, and why one method goes far.
- [Type assertions](https://go.dev/ref/spec#Type_assertions) and
  [type switches](https://go.dev/ref/spec#Type_switches): the comma-ok form, and
  the ordering rule when an interface case is in the list.
- [FAQ: why is my nil error value not nil?](https://go.dev/doc/faq#nil_error) -
  the trap at the end of this lesson, in four paragraphs.
- [io](https://pkg.go.dev/io) - `Reader`, `Writer` and the interfaces built by
  embedding them, which is the design in this lesson used in anger.

## Practise

Three challenges. The first is satisfaction itself: two shapes, one interface
neither of them names, a slice holding both, and a probe for a second interface
with `s.(Named)`. The second is the type switch - eight kinds of value out of
one `any`, with a `fmt.Stringer` case that has to sit after the concrete ones
and a named `float64` that proves why. The third is the nil interface trap,
written both ways: the broken function you can run, the fixed one beside it,
and the guards that keep a nil pointer from ever becoming a non-nil error.
