# String Builder

Strings in Go are **immutable**. Nothing can change one in place, so every `+`
and every `+=` allocates a new string and copies both halves into it. In a loop
that is quadratic:

```go
s := ""
for i := range 10000 {
	s += "x"      // 10000 allocations, ~50 million bytes copied
}
```

The tenth iteration copies nine characters, the ten-thousandth copies 9999. The
work grows with the square of the length, and none of it is your program's
actual job.

[`strings.Builder`](https://pkg.go.dev/strings#Builder) is the answer. It writes
into a growable byte buffer and hands the buffer over as a string at the end,
without copying it again:

```go
var b strings.Builder
b.Grow(10000)              // reserve once, optional but free to get right
for range 10000 {
	b.WriteString("x")
}
s := b.String()
```

The zero value is ready to use, so `var b strings.Builder` needs no `make` and
no `&`. Its methods are `WriteString`, `WriteByte`, `WriteRune` and `Write`, and
`Len` reports how much has been written so far. None of the writes can fail, so
their errors are always `nil` and idiomatic code ignores them.

Because a `Builder` implements [`io.Writer`](https://pkg.go.dev/io#Writer),
[`fmt.Fprintf(&b, ...)`](https://pkg.go.dev/fmt#Fprintf) writes formatted output
straight into it - no intermediate `Sprintf` string. Note the `&`: it needs the
pointer.

That pointer is not a detail. A `Builder` **must not be copied after first
use**. It keeps a pointer to itself to detect exactly that, and a copy panics
with "illegal use of non-zero Builder copied by value". So pass `*strings.Builder`
to a function, never `strings.Builder`, and never assign one to another
variable.

Finally, the formatting verbs are worth knowing: `%q` quotes and escapes a
string, `%02x` renders a byte as two hex digits, `%-10s` left-aligns in a
ten-column field, and `%5d` right-aligns a number in five.

## Task

Fill in the six functions in `main.go`. Every one of them should build its
result with a `strings.Builder`.

1. `Repeat(s, n)` writes `s` out `n` times, reserving the space with `Grow`
   first - there is a test that counts allocations.
2. `JoinLines(lines)` writes each line followed by `"\n"`.
3. `AppendTo(b, items...)` appends into a builder it is given, separating with
   `", "` and joining onto whatever is already there.
4. `QuoteList(items)` renders `["a", "b"]`.
5. `HexDump(data)` renders `"47 6f"`.
6. `Table(names, counts)` renders `%-10s%5d` rows.

## Hints

- `b.Grow(len(s) * n)` before the loop is what makes `Repeat` allocate once.
- `AppendTo` takes a `*strings.Builder` because a `Builder` cannot be copied,
  and because it has to add to the caller's one rather than to a copy of it.
  `b.Len() > 0` is how it tells "first item" from "not the first" across calls.
- For a separator between items and not after them, the usual shape is
  `if i > 0 { write the separator }` at the top of the loop.
- `fmt.Fprintf(&b, "%q", item)` - remember the `&`, or it will not compile.
- `%02x` on a `byte` gives the leading zero; `%x` alone does not, so `0x07`
  would come out as `7`.
- `min(a, b)` is a builtin, which is all `Table` needs to stop at the shorter
  slice.
