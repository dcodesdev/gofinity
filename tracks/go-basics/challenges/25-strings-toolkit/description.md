# Strings Toolkit

Go has no string methods. `s.Split(",")` does not exist, because `string` is a
built-in type and built-in types have no methods. Everything you would reach for
lives in the [**`strings` package**](https://pkg.go.dev/strings) instead, as
plain functions that take the string as their first argument:

```go
strings.ToLower("GO")            // "go"
strings.Contains("gofinity", "fin")  // true
strings.Split("a,b,c", ",")      // []string{"a", "b", "c"}
strings.Join([]string{"a", "b"}, "-")  // "a-b"
strings.TrimSpace("  hi \n")     // "hi"
strings.ReplaceAll("aa", "a", "b")     // "bb"
strings.Repeat("ab", 3)          // "ababab"
```

Two of them are easy to confuse.
[`strings.Split`](https://pkg.go.dev/strings#Split) splits on an exact separator
and keeps every empty field it produces, so `Split("a,,b", ",")` gives three
elements with an empty one in the middle.
[`strings.Fields`](https://pkg.go.dev/strings#Fields) splits on *runs* of any
whitespace and drops the empties, so `Fields(" a   b ")` gives exactly two. When
you want words, you want `Fields`; when you want columns, you want `Split`.

`Fields` plus `Join` is also the shortest way to normalise spacing: taking a
string apart on whitespace and putting it back with single spaces trims the ends
and collapses the middle in one step.

## Task

Fill in the six functions in `main.go`.

1. `Normalize(s)` collapses whitespace runs to one space, trims the ends, and
   lowercases.
2. `CountWords(s)` counts whitespace-separated words.
3. `TitleWords(s)` uppercases the first letter of every word, leaves the rest of
   each word alone, and separates words with a single space.
4. `Initials(name)` returns the uppercased first letter of each word, joined
   with nothing.
5. `Redact(s, secret)` replaces every occurrence of `secret` with one `*` per
   byte of it. An empty `secret` changes nothing.
6. `SplitTrim(s, sep)` splits on `sep`, trims each field, drops the empty ones,
   and never returns `nil`.

## Hints

- `strings.Fields` handles tabs and newlines, not just spaces, which is why the
  tests use them.
- For the first letter, `word[:1]` is the first **byte** and will cut a
  multi-byte letter in half. Convert with `[]rune(word)`, uppercase
  `string(r[0])`, and add `string(r[1:])` back on.
- [`strings.ReplaceAll(s, "", x)`](https://pkg.go.dev/strings#ReplaceAll)
  inserts `x` between every rune rather than doing nothing, so `Redact` needs to
  check for an empty secret before it calls it.
- [`strings.Repeat("*", len(secret))`](https://pkg.go.dev/strings#Repeat) gives
  the mask. `len` is bytes here, and bytes is what the doc comment asks for.
- `make([]string, 0)` is an empty non-nil slice; `var out []string` is nil, and
  the test checks the difference.
