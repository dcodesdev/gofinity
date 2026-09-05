# Basic types

Go's built-in types are a short list, and the list is short on purpose. There is
no separate integer class hierarchy, no implicit widening, and no string type
that hides how it stores its characters. What you get instead is a set of types
whose sizes and behaviour you can state exactly, and a rule that you have to ask
before one becomes another.

## The list

**Integers** come signed and unsigned, in five widths each:

| Signed | Unsigned |
| --- | --- |
| `int8`, `int16`, `int32`, `int64` | `uint8`, `uint16`, `uint32`, `uint64` |
| `int` | `uint`, `uintptr` |

`int` is the one you want unless you have a reason. It is 64 bits on every
platform Gofinity runs on, it is what `len` returns, and it is what an untyped
integer constant becomes when nothing else decides. Reach for a sized type when
the size is part of the problem: a byte in a file, a field in a wire format, a
counter you know must fit in 16 bits.

Two of those names have aliases you will see far more often than the originals:

- `byte` is `uint8`. Use it when the value is data.
- `rune` is `int32`. Use it when the value is a Unicode code point.

They are the same types, not conversions of them, so `[]byte` and `[]uint8` are
interchangeable. The alias exists to say what you mean.

**Floats** are `float64` and `float32`. Use `float64`. `float32` halves the
memory and roughly halves the precision, which is a trade worth making in a
large numeric array and nowhere else.

**Booleans** are `bool`, `true` and `false`. Go will not convert a bool to a
number or a number to a bool, and `if n` where `n` is an `int` does not compile.
An `if` takes a `bool` and nothing else.

**Strings** are immutable sequences of bytes, covered below.

There are also `complex64` and `complex128`. They exist, they are rarely used,
and knowing they exist is enough.

## Every conversion is explicit

This is the rule that catches people arriving from other languages:

```go
var n int = 7
var f float64 = n           // does not compile
var f float64 = float64(n)  // this
```

There is no implicit conversion between numeric types in Go. Not even between
`int` and `int64`, which are the same size. `T(v)` is the only way, and you
write it every time.

It looks like ceremony until you notice what it buys. Two operations that
silently corrupt data in other languages are impossible to write by accident
here, because you cannot get from one type to another without saying so at the
exact place it happens.

### Truncation

Converting a float to an integer **truncates toward zero**. It never rounds.

```go
int(2.9)   // 2
int(-2.9)  // -2
```

If you wanted the nearest whole number, `math.Round` first and convert after.
`int(math.Round(2.9))` is `3`. Getting these two in the wrong order is a real
bug with a small blast radius, which is the worst kind: it is off by one, only
sometimes, and only for some inputs.

### Wrapping

Converting an integer to a narrower type keeps the low bits and discards the
rest. No panic, no error, no complaint.

```go
int8(127)  // 127
int8(128)  // -128
int8(200)  // -56
```

If the value has to survive the trip, you check the range before you convert.
`math.MinInt8` and `math.MaxInt8` (and the same for every other width) are there
for exactly this.

### Integer division

`/` between two integers is integer division. The fraction is not rounded, it is
discarded.

```go
7 / 2        // 3
sum / count  // an int, whatever you meant
```

The fix is to convert **before** dividing, not after:

```go
float64(sum) / float64(count)   // right
float64(sum / count)            // divides as ints, then converts the damage
```

## Untyped constants soften the rule

A literal in Go source has no type until it needs one. That is why this compiles
despite everything above:

```go
var f float64 = 2      // 2 is untyped, and becomes a float64 here
seconds := 90 * time.Second
```

An untyped constant takes the type of its context, and it carries arbitrary
precision until it lands. This is why the strict conversion rule is livable:
constants flow, variables do not.

It also explains a classic surprise:

```go
c * 9 / 5      // fine when c is a float64
9 / 5          // 1, because both constants become ints
```

Writing `9.0 / 5.0` puts the intent in the source instead of leaving it to be
inferred from whichever variable happens to be nearby.

## Strings are bytes

A `string` is an immutable slice of bytes, and Go source files are UTF-8, so a
string literal holds the UTF-8 encoding of what you typed.

```go
s := "héllo"
len(s)   // 6, not 5: é is two bytes
s[0]     // 104, a byte, the 'h'
s[1]     // 195, the FIRST byte of é, which is not a character
```

`len` counts bytes. Indexing gives a `byte`. Neither knows what a character is,
and for ASCII-only text that never matters, which is exactly why it surprises
you the first time it does.

To work in characters you work in runes, and there are two ways:

```go
for i, r := range s {   // r is a rune, i is r's BYTE offset
	fmt.Println(i, string(r))
}

runes := []rune(s)      // decode the whole string at once
len(runes)              // 5
```

Ranging over a string decodes UTF-8 as it goes. The index it hands back is where
the rune starts *in bytes*, not which rune it is, and reaching for that index as
if it were a character position is the single most common string bug in Go.

Reversing a string is where all of this shows up at once. Reverse the bytes and
`é` comes apart into two bytes in the wrong order, which is no longer valid
UTF-8. Convert to `[]rune`, reverse that, convert back.

A few more things worth knowing now:

- Strings are immutable. `s[0] = 'H'` does not compile. Build a new string, or
  work in `[]byte` or `[]rune` and convert at the end.
- `+` concatenates, and in a loop it allocates a new string every time.
  `strings.Join` or `strings.Builder` are what you want at any size.
- `string(65)` is `"A"`, not `"65"`. Converting a number to a string converts it
  to *the character with that code point*. Turning a number into its digits is
  `strconv.Itoa`, and you will meet it in the strings and formatting concept.

## Floats are approximations

`float64` is IEEE 754 binary floating point, and most decimal fractions have no
exact binary representation. `0.1 + 0.2` is not `0.3`, in Go or anywhere else.

Two consequences you should adopt as habits immediately:

- Do not compare computed floats with `==`. Compare the absolute difference
  against a small tolerance instead: `math.Abs(got-want) < 1e-9`.
- Do not store money in a float. Store the smallest unit as an integer: cents,
  not dollars.

`math` also gives you `math.Inf`, `math.NaN`, and the rule that `NaN != NaN`,
which is the one exception to the sentence "a value equals itself".

## Further reading

- [Types](https://go.dev/ref/spec#Types) - the predeclared types and what each
  one actually holds.
- [Conversions](https://go.dev/ref/spec#Conversions) - the exact rules for when
  `T(x)` is allowed, including the number to string case.
- [Constants](https://go.dev/ref/spec#Constants) - untyped constants, arbitrary
  precision, and where a default type comes from.
- [The math package](https://pkg.go.dev/math) - `Abs`, `Inf`, `NaN` and the
  rounding functions the float rules make you need.
- [Strings, bytes, runes and characters in Go](https://go.dev/blog/strings) -
  what a string holds, one byte at a time.

## Practise

Three challenges. The first is the conversion rules on their own, the second is
what a string actually holds, and the third puts integers, floats and rounding
in the same function and asks you to keep them straight.
