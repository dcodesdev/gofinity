# Values and variables

Every name in a Go program is introduced one of three ways, and the choice says
something. This lesson is about picking the right one, about the value a
variable holds before you assign anything to it, and about the one place Go
does have a little magic: `iota`.

## Three ways to name a value

### `const`

A constant is fixed at compile time. It cannot be reassigned, it cannot have
its address taken, and it costs nothing at run time because the compiler
substitutes the value at every use.

```go
const AppName = "Gofinity"
const MaxRetries = 3
```

Constants can only hold what the compiler can evaluate: numbers, strings,
booleans, runes, and expressions over those. `const Now = time.Now()` does not
compile, and that is the whole rule for when `const` is available to you.

### `var`

A variable, with or without an initial value.

```go
var attempts int          // 0
var name string = "Ada"   // explicit type, rarely needed
var limit = 10            // type inferred as int
```

`var` is the only form that works at package level, and the only form that can
say "give me the zero value" without naming a value.

### `:=`

Short declaration: declare and assign in one step, with the type inferred from
the right-hand side.

```go
spent := clamp(used)
lines := []string{"first", "second"}
```

It works **only inside a function**. Inside one, it is what you should reach
for by default. Save `var` for the zero-value case or when you need the type
spelled out for clarity.

One rule catches everyone once: `:=` requires at least one *new* name on the
left. `a, err := f()` followed by `b, err := g()` is fine, because `b` is new
and `err` is merely assigned.

And Go treats a declared-but-never-read local variable as a compile error, the
same way it treats an unused import. A variable you do not use is a mistake or
a leftover, and the compiler would rather tell you now.

## Typed and untyped constants

`const MaxRetries = 3` is an **untyped** constant. It has no type yet; it is
just the number three, and it takes whatever type its context needs:

```go
var i int = MaxRetries        // int
var f float64 = MaxRetries    // float64
var d time.Duration = MaxRetries * time.Second
```

Write `const MaxRetries int = 3` and you have pinned it: it is now an `int` and
the float line stops compiling. Untyped is the better default, because Go has
no implicit conversion between numeric *variables*, and an untyped constant is
how you avoid needing one.

Untyped constants are also arbitrary precision until they land somewhere.
`const Big = 1 << 62` is fine as a constant even where an `int32` variable
could not hold it; the error, if any, appears where you assign it.

## The zero value

Go has no uninitialised memory. A variable declared without a value gets its
type's zero value, guaranteed:

| Type | Zero value |
| --- | --- |
| `int`, `int64`, `uint`, ... | `0` |
| `float64` | `0` |
| `string` | `""` |
| `bool` | `false` |
| slice, map, pointer, function, channel, interface | `nil` |

This is not a detail, it is a design. It is why so much Go code has no
constructor: the zero value is meant to be immediately useful.

```go
var total int          // 0, ready to add to
var out []int          // nil, and append works anyway
out = append(out, 1)   // append allocates on first use
```

A `nil` slice behaves like an empty one for `len`, `range` and `append`. It
differs only when you compare it against `nil`. A `nil` **map** is the one to
watch: reading from it is fine and gives the zero value, but *writing* to it
panics. A map needs `make(map[string]int)` before it can hold anything.

```go
var counts map[string]int
fmt.Println(counts["nothing"])  // 0, fine
counts["a"] = 1                 // panic: assignment to entry in nil map
```

Reading a missing key giving the zero value is the same guarantee wearing a
different hat, and it is why counting words in Go needs no "if absent" branch.

## Enumerations with `iota`

Go has no `enum` keyword. An enumeration is a defined integer type plus a
`const` block, and `iota` supplies the numbers.

`iota` starts at `0` in each `const` block and increases by one per line. A
constant line with no expression repeats the previous line's expression, which
is why only the first line needs to say anything:

```go
type Weekday int

const (
	Sunday Weekday = iota // 0
	Monday                // 1
	Tuesday               // 2
	// ...
	Saturday              // 6
)
```

The first line does two jobs: it starts the counter and it gives every constant
in the block the type `Weekday`. Because `Weekday` is a defined type and not an
alias, `Monday + 1` is a `Weekday`, and a function taking a `Weekday` will not
accept a bare `int` without a conversion. That is the point of declaring the
type at all.

`iota` counts lines, not constants, so a skipped value is spelled with `_`:

```go
const (
	_  = iota             // skip 0
	KB = 1 << (10 * iota) // 1 << 10
	MB                    // 1 << 20
)
```

### Making it print

An integer enum prints as a number, which is useless in a log line. Give the
type a `String() string` method and `fmt` will use it everywhere, for `%v`,
`Println`, `Sprint`, all of it:

```go
func (d Weekday) String() string {
	if d < Sunday || d > Saturday {
		return fmt.Sprintf("Unknown(%d)", int(d))
	}
	return dayNames[d]
}
```

Two things there are worth keeping.

The `int(d)` conversion is not decoration. `fmt.Sprintf("Unknown(%d)", d)`
would ask `fmt` to format a `Weekday`, `fmt` would call `String` again, and the
program would recurse until the stack ran out. Converting to `int` first is how
you break the loop, and it is a trap every Go programmer meets exactly once.

The range check is not decoration either. Nothing stops a caller writing
`Weekday(9)`; a defined type constrains what compiles, not what a conversion
can produce. Anything indexing a table by the enum value has to guard first, or
it panics.

## Further reading

- [Declarations and scope](https://go.dev/ref/spec#Declarations_and_scope) - the
  spec on `var`, `const`, `:=`, and where each name is visible.
- [Constant declarations](https://go.dev/ref/spec#Constant_declarations) - `iota`
  and the rule that a line with no expression repeats the one above it.
- [Effective Go: names](https://go.dev/doc/effective_go#names) - why Go names are
  short, and how the case of the first letter carries the meaning.
- [The fmt package](https://pkg.go.dev/fmt) - the `Stringer` interface that makes
  an enum print as something other than a number.

## Practise

Three challenges, in order. The first fixes the three declaration forms in
place, the second is about what a variable holds before you touch it, and the
third builds a real enum with a `String` method and the recursion trap waiting
inside it.
