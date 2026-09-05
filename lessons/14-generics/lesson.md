# Generics

For most of its life Go had two answers to "write this function once for every
element type": copy it per type, or take `interface{}` and assert your way back
out. The first is a maintenance tax, the second throws away the type checking
that is the reason you are using Go. Go 1.18 added a third.

```go
func Map[T, U any](s []T, f func(T) U) []U {
	out := make([]U, len(s))
	for i, v := range s {
		out[i] = f(v)
	}
	return out
}
```

`[T, U any]` is the **type parameter list**. `T` and `U` are types the caller
supplies; `any` is the constraint on each. Inside the body they behave like
ordinary types whose names you happen not to know.

## Inference does the work

You rarely write the types at the call site:

```go
lengths := Map(words, func(s string) int { return len(s) })
```

`words` is a `[]string`, so `T` is `string`; the literal returns an `int`, so
`U` is `int`. Explicit instantiation is legal - `Map[string, int](words, f)` -
and occasionally necessary when nothing in the arguments pins a parameter down,
but needing it everywhere usually means the signature is wrong.

Instantiation makes a real, distinct thing: `Map[string, int]` and
`Map[int, int]` are two functions as far as the type checker is concerned, and
`Stack[int]` and `Stack[string]` are two types with no assignment between them.

## Constraints are the whole idea

A type parameter is only as capable as its constraint. Under `any` you can move
a `T` around and nothing else - you cannot add two of them, compare them with
`==`, or print them as a number. That is not a limitation to work around, it is
the guarantee: the body is checked **once**, against the constraint, so a call
that compiles cannot fail inside.

Three levels, in the order you reach for them:

```go
[T any]           // move it around
[T comparable]    // == and !=, and it can be a map key
[T Number]        // arithmetic, because Number lists the numeric types
```

## Type sets

A constraint is an interface, but a constraint interface may hold **types**
instead of methods:

```go
type Float interface {
	~float32 | ~float64
}
```

`|` is union. The operators you may use on a `T` are the ones valid for
**every** type in the set, which is why `+` works over a numeric set and `%`
does not: floats do not have it. An interface with a type set can only be used
as a constraint - `var f Float` does not compile - and that is the one place
interfaces mean something different from everywhere else in the language.

### Write the tilde

`~float32` means "float32 and any named type whose underlying type is float32".
Without it, the set is those two exact types, and this fails:

```go
type Celsius float64
Sum([]Celsius{20.5, 1.5})   // Celsius is not float64
```

Named types over primitives are ordinary Go - `type ID int`, `type Celsius
float64` - and omitting the tilde rejects all of them for no benefit. Write the
tilde unless you genuinely mean one exact type.

Constraints embed each other, so real ones are built up rather than repeated:

```go
type Integer interface{ Signed | Unsigned }
type Number  interface{ Integer | Float }
type Ordered interface{ Integer | Float | ~string }
```

`string` is in `Ordered` and not in `Number`, which is exactly why `Max` and
`Sum` take different constraints. The standard library ships `cmp.Ordered`;
writing these once by hand is how you learn what is in it.

## The zero value of a T

You cannot write `return nil` or `return 0` for a `T`, because `T` might be
neither. One spelling always works:

```go
var zero T
return zero, false
```

That is why generic lookups so often return `(T, bool)`: the comma-ok pattern is
the only honest way to say "nothing here" when you do not know what "nothing"
looks like.

## Generic types

Types take parameters too, and that is where this stops being a convenience:

```go
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }
```

The `[T]` on the receiver **declares** the parameter for the method body; the
constraint is not repeated, it came with the type. Every other use of the name
needs its argument: `*Stack[string]`, `map[string]*Stack[int]`.

Design so the zero value works. `var s Stack[int]` is an empty, usable stack
because its only field is a nil slice - a constructor should be a convenience,
not a requirement.

### Methods cannot add type parameters

This is the rule that catches everyone:

```go
func (s *Stack[T]) Map[U any](f func(T) U) *Stack[U]   // does not compile
```

A method's type parameters can only be the receiver's. `U` has nowhere to come
from. The fix is to make it a function:

```go
func MapStack[T, U any](s *Stack[T], f func(T) U) *Stack[U]
```

That single rule explains a lot of the standard library's shape: `slices.Sort`,
`maps.Keys` and friends are package-level functions rather than methods on a
container, and an idiomatic generic type has a small method set with free
functions around it.

## When not to

Generics are for code where the **logic** is identical and only the type
differs. They are not a design goal.

- One caller, one type? Write it for that type. `[T any]` with a single
  instantiation is a worse `func(int)`.
- Behaviour differs per type? That is an interface with a method, not a type
  set.
- Reaching for reflection to make it work? Stop; the constraint is wrong.

The honest rule from the Go team: write the concrete version first. When you
have written it three times and the bodies are the same, replace them.

## Further reading

- [Tutorial: getting started with generics](https://go.dev/doc/tutorial/generics):
  type parameters and constraints, built up one step at a time.
- [Type parameter declarations](https://go.dev/ref/spec#Type_parameter_declarations)
  and [general interfaces](https://go.dev/ref/spec#General_interfaces): the
  spec on type sets, unions and the tilde.
- [When to use generics](https://go.dev/blog/when-generics) - the Go team's own
  rule for when a type parameter is the wrong answer.
- [cmp](https://pkg.go.dev/cmp) - `Ordered` and `Compare`, the constraints you
  write by hand in this lesson, shipped.
- [slices](https://pkg.go.dev/slices) and [maps](https://pkg.go.dev/maps) - what
  a generic standard library looks like: free functions, not methods.

## Practise

Three challenges. The first writes `Map`, `Filter`, `Reduce`, `Keys` and
`First` - the `any` and `comparable` constraints, inference at the call site,
and `var zero T` for a miss. The second builds `Signed`, `Unsigned`, `Number`
and `Ordered` from type sets and uses them for `Sum`, `Average`, `Min`, `Max`
and `Clamp`, with named types in the tests that only compile if you wrote the
tilde. The third is a generic `Stack[T]`: methods on a parameterised type, a
`Pop` that clears the slot it vacates, and a `MapStack` that has to be a
function because a method cannot introduce a `U`.
