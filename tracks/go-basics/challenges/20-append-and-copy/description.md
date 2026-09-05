# Append and Copy

A slice has a length and a capacity, and `append` is the built-in that lives in
the gap between them.

```go
s := make([]int, 0, 4)   // len 0, cap 4
s = append(s, 1)         // len 1, cap 4 - room to spare, no allocation
```

While there is capacity left, `append` writes into the existing backing array
and hands back a slice one longer. When there is not, it allocates a **bigger
array**, copies the elements across, and returns a slice pointing at the new
one. The old array is left behind, still referenced by whatever slices were
looking at it.

That is why `append` returns a value and why you must use it:

```go
s = append(s, 1)   // right
append(s, 1)       // wrong, and it does not compile
```

The growth rule itself is deliberately unspecified. You do not get to depend on
capacity doubling, and code that reads `cap()` to decide something is code that
will break. What you *can* depend on is that a slice with spare capacity keeps
its array and one without gets a new one, which is the whole source of the
aliasing surprises in the next challenge.

`copy(dst, src)` is the other half. It copies element by element and stops at
the shorter of the two, returning how many it copied. It never resizes
anything, so `copy` into a slice of length 0 copies nothing however large its
capacity.

Together they give the two idioms you will write forever: `make` plus `copy` for
an independent duplicate, and `make` with a capacity plus `append` for building
a result whose size you already know.

## Task

Fill in the six functions in `main.go`.

1. `AppendAll(dst, values...)` appends and returns the result.
2. `CloneInts(s)` returns an independent copy. `nil` clones to `nil`; an empty
   slice clones to an empty non-nil slice.
3. `CopyInto(dst, src)` copies what fits and returns the count.
4. `Insert(s, i, v)` and `RemoveAt(s, i)` return new slices, leaving `s` alone
   and clamping an out-of-range `i` instead of panicking.
5. `Concat(a, b)` returns `a` followed by `b` in a backing array of its own.

## Hints

- `append(dst, values...)` spreads a slice into a variadic parameter, the same
  `...` from the variadics challenge.
- `CloneInts` is `make` then `copy`. `make([]int, len(s))` gives a slice of the
  right length; `copy` fills it.
  [`slices.Clone`](https://pkg.go.dev/slices#Clone) exists and does the same
  thing, but write it out once first.
- `CopyInto` really is one line: `copy` already stops at the shorter slice and
  already returns the count.
- Build `Insert` and `RemoveAt` into a fresh slice with the capacity you know
  you need, then append the pieces: `s[:i]`, the new element, `s[i:]`.
- `Concat` is the trap. `append(a, b...)` reuses `a`'s array whenever it has the
  room, so the result would be a view of `a` and writing to it would corrupt
  `a`. Allocate with `make([]int, 0, len(a)+len(b))` and append into that.
