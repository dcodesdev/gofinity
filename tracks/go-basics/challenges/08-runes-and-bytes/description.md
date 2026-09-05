# Runes and Bytes

A Go string is a read-only slice of **bytes**, and its bytes are UTF-8. That one
sentence explains most of what surprises people about strings in Go.

`len(s)` counts bytes. Indexing, `s[0]`, gives you a single `byte`. Neither one
knows anything about characters, so for `"héllo"` you get `6` and the first
byte of `h`, and for `"日本語"` you get `9`.

A **rune** is a Unicode code point: the thing you probably mean when you say
"character". It is an alias for `int32`. Two operations decode UTF-8 for you:

```go
for i, r := range s { ... }  // i is a BYTE offset, r is a rune
runes := []rune(s)           // decodes the whole string at once
```

Ranging over a string steps one rune at a time, but the index it hands you is
where that rune *starts in bytes*, not which rune it is. That gap is where the
bugs live.

## Task

Finish the four functions in `main.go`.

1. `ByteLen(s string) int` - how many bytes.
2. `RuneLen(s string) int` - how many runes.
3. `RuneAt(s string, i int) (rune, bool)` - the rune at position `i` counted in
   runes. `RuneAt("héllo", 1)` is `'é'`, not the second byte of `é`. Out of
   range, including a negative `i`, returns `false`.
4. `ReverseRunes(s string) string` - reverse the string without cutting a
   multi-byte rune in half. `"héllo"` reverses to `"olléh"`, and reversing
   twice must give the original back.

## Hints

- `for range s` with no variables at all is legal and is all `RuneLen` needs.
  [`utf8.RuneCountInString(s)`](https://pkg.go.dev/unicode/utf8#RuneCountInString)
  from `unicode/utf8` is the standard-library way to say the same thing.
- In `RuneAt`, ignore the index `range` gives you and keep your own counter, or
  convert with `[]rune(s)` and index that. The byte offset is not the rune
  position: in `"héllo"` the `l` after `é` starts at byte 3.
- Reversing bytes is the classic wrong answer. `s[len(s)-1-i]` walks backwards
  through bytes, and the two bytes of `é` come out swapped, which is no longer
  valid UTF-8. Convert to `[]rune` first.
- Swapping two elements needs no temporary variable:

  ```go
  runes[i], runes[j] = runes[j], runes[i]
  ```

  and a `for` can advance two things at once with
  `for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1`.
- `string(runes)` converts a `[]rune` back to a string. `string(r)` on a single
  rune works too, but note that `string(65)` is `"A"` and not `"65"`: turning a
  number into its digits is [`strconv`](https://pkg.go.dev/strconv), which you
  will meet later.
