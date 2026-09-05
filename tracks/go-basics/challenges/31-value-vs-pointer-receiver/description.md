# Value versus Pointer Receiver

A method is a function with a receiver: an extra parameter written before the
name.

```go
func (c Counter) Value() int   { return c.N }
func (c *Counter) Add(delta int) { c.N += delta }
```

The receiver is an ordinary parameter, which means the ordinary rule applies. A
value receiver gets a **copy**, so writing to it changes nothing the caller can
see. A pointer receiver gets the address, so writing to it changes the caller's
value. That is the entire distinction, and every other rule about receivers
follows from it.

## Calling one

Go inserts the `&` and the `*` for you when it can:

```go
var c Counter
c.Add(3)      // shorthand for (&c).Add(3)

p := &Counter{}
p.Value()     // shorthand for (*p).Value()
```

This works because `c` is a variable, and a variable is **addressable**. Not
everything is. A map element, the result of a function call and a range copy are
not addressable, so a pointer method cannot be called on them at all. That is
why a loop that means to mutate has to reach the slice element:

```go
for _, c := range cs { c.Add(1) }   // adds to a copy, then throws it away
for i := range cs    { cs[i].Add(1) } // adds to the slice
```

The first loop compiles - `c` is a variable, so it is addressable - and does
nothing. It is the most common receiver mistake in Go.

## Method sets

The distinction shows up again when a type meets an interface. The method set of
`*Counter` contains both value-receiver and pointer-receiver methods; the method
set of `Counter` contains only the value-receiver ones. So `*Counter` satisfies
an interface with `Add` in it, and `Counter` does not:

```go
type Adder interface{ Add(int) }
var _ Adder = (*Counter)(nil)   // fine
var _ Adder = Counter{}         // does not compile
```

## Methods on any named type

Receivers are not a struct feature. Any type declared in this package can have
methods, including one defined over `float64`:

```go
type Temperature float64
func (t Temperature) Warmer(d float64) Temperature { return t + Temperature(d) }
```

A method named `String() string` is special only by convention: `fmt` looks for
it, so printing the value formats it without anyone asking.

## Task

Fill in the methods and functions in `main.go`. `Value`, `Plus`, `SumValues`,
`Warmer` and `String` read; `Add`, `Reset` and `AddEach` write.

## Hints

- `Plus` has a value receiver, so `c` is already a copy: write to it and return
  it, exactly like `AddPages` in the structs challenge.
- `AddEach` needs `for i := range cs` and `cs[i].Add(delta)`. Ranging over
  values gives you copies, and adding to a copy is a no-op.
- In `String`, format `float64(t)` rather than `t`. `%.1f` on a `Temperature`
  would call `String` again and recurse until the stack runs out.
- `Warmer` cannot write `t + d`: `d` is a `float64` and `t` is a `Temperature`.
  Go has no implicit conversion between named types, so convert one of them.
