# Struct Equality

Two structs of the same type can be compared with `==` when **every field is
comparable**, and the comparison is field by field:

```go
type Point struct{ X, Y int }

Point{1, 2} == Point{1, 2}   // true
```

No method, no interface, no `Equals`. That is the language rule, and it makes a
comparable struct usable as a **map key** and as a `switch` case.

## Which types are comparable

| Comparable | Not comparable |
| --- | --- |
| numbers, strings, bools | slices |
| pointers, channels | maps |
| interfaces (see below) | funcs |
| arrays of comparable types | structs or arrays containing any of the above |

The rule is recursive: a struct is comparable exactly when all its fields are,
so one slice field anywhere inside makes the whole thing uncomparable.

```go
type Config struct {
	Name  string
	Hosts []string
}

a == b   // compile error: invalid operation, Config cannot be compared
```

This is a **compile-time** error, which is the good case. Compare that with a
map key: `map[Config]int` also fails to compile, so the mistake is caught where
you make it.

Note that `[3]int` is comparable but `[]int` is not. An array has its length in
its type and its elements in the value; a slice is a header pointing somewhere
else, and Go refuses to guess whether you meant the header or the contents.

## Comparing the rest

When `==` is off the table there are two routes.

**By hand**, which is what production code usually does:

```go
func EqualConfigs(a, b Config) bool {
	return a.Name == b.Name && slices.Equal(a.Hosts, b.Hosts)
}
```

It compiles to straight-line code, it is obvious, and you get to decide the
edge cases - in particular whether a nil slice equals an empty one.

**[`reflect.DeepEqual`](https://pkg.go.dev/reflect#DeepEqual)**, which works on
anything:

```go
reflect.DeepEqual(a, b)
```

It is slow, it is unchecked by the compiler, and it is stricter than people
expect: `DeepEqual([]string(nil), []string{})` is **false**. In tests that is
usually a false alarm rather than a real difference, which is why
[`slices.Equal`](https://pkg.go.dev/slices#Equal) and
[`maps.Equal`](https://pkg.go.dev/maps#Equal) are the better tools when the
shape is known.

## Two traps

**Pointer fields compare addresses.** A struct holding a `*Node` is comparable,
because pointers are, but `==` asks "the same object?", not "the same
contents?":

```go
a := Node{Next: &Node{Label: "x"}}
b := Node{Next: &Node{Label: "x"}}
a == b   // false - two different allocations
```

**Comparing interfaces can panic.** `==` on two `any` values compiles, always,
because interfaces are comparable as a type. At run time it compares the dynamic
types first, and only if they match does it compare the values - and if that
dynamic type turns out to be uncomparable, the program panics with "comparing
uncomparable type []string". So the one place the compiler cannot help you is
exactly the place a runtime failure is waiting.
[`reflect.TypeOf(v).Comparable()`](https://pkg.go.dev/reflect#TypeOf) is how to
ask first.

Finally, field **order and type** are part of the type. `struct{A, B int}` and
`struct{B, A int}` are different types and cannot be compared at all, and two
structs with the same fields but different names only compare after an explicit
conversion.

## Task

Fill in the seven functions in `main.go`. `Point`, `Config` and `Node` are
declared for you, and each one is there to exercise a different rule.

## Hints

- `CountUnique` wants a `map[Point]struct{}` or `map[Point]bool`. It is the map
  key that proves `Point` is comparable.
- `EqualConfigs` cannot use `==` on the whole struct. `a.Where == b.Where` on
  the `Point` field is fine, though - only `Hosts` is the problem, and
  `slices.Equal` handles it.
- `slices.Equal(nil, []string{})` is `true` and `reflect.DeepEqual` on the same
  pair is `false`. The two tests want that difference, so do not try to make
  them agree.
- `SameNode` really is just `a == b`. The test is there to show you what that
  means for the pointer field.
- `reflect.TypeOf(nil)` returns a nil `reflect.Type`, so `CanCompare` has to
  check for that before calling `Comparable` on it - otherwise it panics on the
  nil case rather than returning false.
- `Dedup` should start from `[]Point{}` so an empty result is non-nil.
