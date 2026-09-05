# Number Constraints

`any` lets you move a value around. It does not let you add two of them. To
write `Sum`, the constraint has to promise arithmetic, and the way a constraint
promises anything in Go is by listing the types it accepts.

## Type sets

An interface used as a constraint may hold, instead of methods, a **type set**:

```go
type Float interface {
	~float32 | ~float64
}
```

Read it as "any type whose underlying type is `float32` or `float64`". `|` is
union. An interface with a type set can only be used as a constraint - you
cannot declare a variable of type `Float` - and that is the one place the
compiler treats interfaces differently from every other use.

Inside a generic function, the operators available on `T` are the ones available
on **every** type in the set. `+` works for a set of numeric types because it
works for all of them. `/` works too. `%` does not, because floats do not have
it, so a `Sum` over `Number` compiles and a `Mod` over `Number` does not.

## The tilde

`~float32` means "float32 and every named type defined from it". Without the
tilde, `float32 | float64` accepts literally those two types and nothing else -
so this would not compile:

```go
type Celsius float64
Sum([]Celsius{1, 2})   // Celsius is not float64
```

Named types over primitives are ordinary Go: `type ID int`, `type Celsius
float64`. Leaving the tilde off makes your function reject all of them for no
reason. **Write the tilde.** The only time you leave it off is when you really
do mean one exact type.

## Composing constraints

Constraints embed each other, so you build them up rather than repeating the
union everywhere:

```go
type Integer interface{ Signed | Unsigned }
type Number  interface{ Integer | Float }
```

`Ordered` - everything `<` works on - is `Integer | Float | ~string`. `string`
belongs there and not in `Number`, which is exactly why `Max` and `Sum` want
different constraints: `Max` over strings is meaningful, `Sum` over strings
would mean concatenation and is somebody else's function.

The standard library ships [`cmp.Ordered`](https://pkg.go.dev/cmp#Ordered), and
[`x/exp/constraints`](https://pkg.go.dev/golang.org/x/exp/constraints) has the
rest. Write them by hand once, here, so you know what is in them.

## Converting a T

`float64(v)` is legal when every type in `T`'s set converts to `float64`, which
is true for `Number`. That is how `Average` returns a `float64` from a `[]int`:
convert each element, not the sum, or a `[]int8` overflows before you divide.

## Task

Declare `Signed`, `Unsigned`, `Integer`, `Float`, `Number` and `Ordered`, then
implement `Sum`, `Average`, `Min`, `Max`, `Clamp`, `Abs` and `SumValues`.

## Hints

- Every union member gets a `~`. The test defines named types over `int` and
  `float64` and will not compile without it.
- `Unsigned` needs `uintptr` as well as `uint`, `uint8`, `uint16`, `uint32` and
  `uint64`.
- `Sum` starts from `var total T` - the zero value idiom again.
- `Average` returns `(float64, bool)`; an empty slice is `(0, false)`. Convert
  each element as you add it.
- `Min` and `Max` return `(T, bool)`. Seed the running answer with the first
  element, not with a zero value, or a slice of negatives gives you 0.
- `Clamp` is two comparisons. When `hi < lo` the range is empty and there is no
  sensible answer, so return `lo`.
- `Abs` is over `Signed | Float`, not `Number`: negating an unsigned value is
  not a thing you want.
- `SumValues` takes a `map[K]V` and needs two type parameters with two different
  constraints, one `comparable` and one `Number`.
