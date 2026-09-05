# Strings and Formatting

You already know what a string *is*: an immutable, read-only slice of bytes
holding UTF-8. This lesson is about what you do with one, and almost all of it
lives in four packages - `strings`, `strconv`, `fmt` and `unicode`.

## There are no string methods

`string` is a built-in type, and built-in types cannot have methods. So there is
no `s.Split(",")`, no `s.Trim()`, no `s.ToUpper()`. Every operation is a plain
function in the `strings` package taking the string as its first argument:

```go
strings.ToLower("GO")                  // "go"
strings.ToUpper("go")                  // "GO"
strings.TrimSpace("  hi \n")           // "hi"
strings.Contains("gofinity", "fin")    // true
strings.HasPrefix("gofinity", "go")    // true
strings.Index("gofinity", "f")         // 2, or -1 when absent
strings.Replace("aaa", "a", "b", 2)    // "bba", the first two only
strings.ReplaceAll("aaa", "a", "b")    // "bbb"
strings.Repeat("ab", 3)                // "ababab"
strings.EqualFold("Go", "GO")          // true, case-insensitive comparison
```

Everything here returns a **new** string; nothing modifies the one you passed.
A function that "changes" a string returns the changed copy, and ignoring the
return value is one of the most common beginner bugs in Go.

Comparison needs no function at all. `==`, `<` and `>` work directly on strings
and compare bytes, which for ASCII is alphabetical order and for anything else
is code-point order.

## Split, Fields and Cut

Three ways to break a string apart, and choosing the wrong one is a bug rather
than a style choice.

`strings.Split(s, sep)` splits on an exact separator and **keeps every empty
field**:

```go
strings.Split("a,,b", ",")     // ["a", "", "b"]
strings.Split("", ",")         // [""] - one empty field, not zero
strings.Split("abc", "")       // ["a", "b", "c"] - splits into runes
```

`strings.Fields(s)` splits on runs of any whitespace and **drops the empties**:

```go
strings.Fields("  a   b\tc\n")   // ["a", "b", "c"]
strings.Fields("   ")            // [] - zero fields
```

When you want words, you want `Fields`. When you want columns, and an empty
column means something, you want `Split`.

`strings.Cut(s, sep)` splits on the **first** occurrence and tells you whether
it found one:

```go
key, value, found := strings.Cut("a=1=2", "=")   // "a", "1=2", true
```

That third result is the reason it exists. `Split` on a `"key=value"` line gives
you a slice you then have to check the length of; `Cut` gives you the two halves
and a boolean, which is exactly the shape an `if` wants.

Going back the other way is `strings.Join(parts, sep)`. `Fields` followed by
`Join` is the shortest way to normalise spacing in a string: taking it apart on
whitespace and putting it back with single spaces trims the ends and collapses
the middle in a single expression.

## Trimming

There are more trim functions than people expect, and they do different things:

```go
strings.TrimSpace(s)              // whitespace off both ends
strings.Trim(s, ".,!")            // any of those CHARACTERS off both ends
strings.TrimLeft(s, "0")          // any of those characters off the left
strings.TrimPrefix(s, "http://")  // that exact STRING, once, if present
strings.TrimSuffix(s, ".go")      // that exact string, once, if present
```

`Trim` takes a **cutset** - a set of characters, in any order - while
`TrimPrefix` takes a literal string. `strings.Trim("banana", "ban")` is `""`,
not `"ana"`, because `b`, `a` and `n` are all in the cutset. If you meant "strip
this exact prefix", `TrimPrefix` is the one, and it returns the string unchanged
when the prefix is not there, so it needs no `HasPrefix` guard in front of it.

## Parsing: strconv

Everything from outside the program - a flag, an environment variable, a form
field, a line of CSV - arrives as a string. `strconv` turns it into something
else, and every function returns the value **and an error**:

```go
n, err := strconv.Atoi("42")               // int
f, err := strconv.ParseFloat("1.5", 64)    // float64
b, err := strconv.ParseBool("true")        // bool
i, err := strconv.ParseInt("ff", 16, 64)   // int64, base 16
```

And back:

```go
strconv.Itoa(42)                          // "42"
strconv.FormatFloat(1.5, 'f', 2, 64)      // "1.50"
strconv.FormatBool(true)                  // "true"
```

Two habits go with this. First, these parsers are **strict**: no surrounding
whitespace, no trailing junk. `strconv.Atoi(" 42 ")` is an error, so a
`strings.TrimSpace` almost always comes first.

Second, the error is more than a message. It is a `*strconv.NumError` carrying
the function name, the input, and one of two sentinels:

```go
_, err := strconv.Atoi("abc")
errors.Is(err, strconv.ErrSyntax)   // true - not a number at all
_, err = strconv.Atoi("999999999999999999999")
errors.Is(err, strconv.ErrRange)    // true - a number, but too big to hold
```

"Not a number" and "too large" usually deserve different messages to a user, so
pass a `strconv` error along unchanged rather than flattening it into a string
of your own.

Where a default exists, wrap the two-result call into a one-result one at the
place that knows the default:

```go
func ParseIntOr(s string, fallback int) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return fallback
	}
	return n
}
```

Note also that a **conversion is not a parse**. `string(65)` is `"A"`, the
character with that code point, not `"65"`; `strconv.Itoa(65)` is `"65"`. `go
vet` warns about the first one because it is almost never what was meant.

## Building: strings.Builder

Strings are immutable, so `+` allocates a new one and copies both halves in.
Inside a loop that is quadratic:

```go
s := ""
for i := range 10000 {
	s += "x"       // 10000 allocations, ~50 million bytes copied
}
```

The tenth iteration copies nine bytes, the ten-thousandth copies 9999. The work
grows with the square of the length and none of it is useful.

`strings.Builder` writes into a growable buffer and hands it over as a string at
the end, without a final copy:

```go
var b strings.Builder
b.Grow(10000)
for range 10000 {
	b.WriteString("x")
}
s := b.String()
```

The zero value is ready to use, so no `make` and no `&` on the declaration.
`Grow(n)` reserves capacity up front when you know roughly how much you will
write, turning many reallocations into one. The methods are `WriteString`,
`WriteByte`, `WriteRune` and `Write`, plus `Len` for how much has been written
and `Reset` to start over. None of the writes can fail, so their error results
are always `nil` and idiomatic code ignores them.

A `Builder` **must not be copied after first use**. It holds a pointer to itself
precisely to catch that, and using a copy panics with "illegal use of non-zero
Builder copied by value". So pass `*strings.Builder` to a function, never
`strings.Builder`, and do not assign one builder to another variable.

For the simple case of joining a slice you already have, `strings.Join` is
shorter and does the same thing internally. Reach for a `Builder` when the
pieces come from a loop, from several sources, or from formatting.

## Formatting: fmt

`fmt.Sprintf` returns a formatted string, `Printf` writes it to stdout, and
`Fprintf` writes it to any `io.Writer` - including a `strings.Builder`, which is
why formatted output can go straight into one:

```go
fmt.Fprintf(&b, "%-10s%5d\n", name, count)
```

The `&` is required: the writer methods are on `*strings.Builder`.

The verbs worth memorising:

| Verb | Does |
| --- | --- |
| `%v` | the default form of any value |
| `%+v` | a struct with its field names |
| `%#v` | Go syntax, the form you could paste back into source |
| `%T` | the type, not the value |
| `%d` `%f` `%t` `%s` | integer, float, bool, string |
| `%q` | a **quoted, escaped** string - `"a\nb"` rather than two lines |
| `%x` `%02x` | hex; the `02` pads a byte to two digits |
| `%%` | a literal percent sign |

Width and alignment go between the `%` and the verb: `%5d` right-aligns in five
columns, `%-10s` left-aligns in ten, `%6.2f` gives six columns with two decimal
places. A value wider than the field is never truncated, it just overflows,
which is what makes fixed-width tables drift when one row is long.

`%q` is the one to reach for in error messages and logs. `got abc` is ambiguous
about trailing spaces; `got "abc "` is not.

Finally, `fmt.Errorf` formats an error the same way, and `%w` inside it wraps
another error so `errors.Is` can still see through. That is the subject of a
later lesson, but it is the same formatting machinery.

## Further reading

- [Strings, bytes, runes and characters in Go](https://go.dev/blog/strings) - what
  a string actually holds, and how `range` decodes it.
- [strings](https://pkg.go.dev/strings) - the package this lesson is mostly about,
  including `Builder` and its documented no-copy rule.
- [strconv](https://pkg.go.dev/strconv) - parsing and formatting numbers, and the
  `ErrSyntax` and `ErrRange` sentinels its errors wrap.
- [fmt](https://pkg.go.dev/fmt) - the verb table, in the package documentation
  rather than in anyone's cheat sheet.
- [unicode](https://pkg.go.dev/unicode) - the rune classification functions the
  `strings` predicates take.

## Practise

Three challenges. The first is the `strings` toolkit itself: normalising with
`Fields` and `Join`, splitting and trimming, and the `ReplaceAll` trap where an
empty search string inserts the replacement between every rune. The second is
`strconv`, including passing its errors through untouched so `errors.Is` can
still tell `ErrSyntax` from `ErrRange`. The third is `strings.Builder` and the
`fmt` verbs, with an allocation-counting test that fails if you reach for `+=`
in a loop.
