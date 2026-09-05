# Slice Aliasing

A slice is a pointer, a length and a capacity. Copy a slice - assign it, pass it
to a function, return it - and you copy those three words. You do not copy the
array they point at.

So two slices can be different values that describe the **same memory**:

```go
s := []int{1, 2, 3, 4}
view := s[1:3]
view[0] = 99
fmt.Println(s)   // [1 99 3 4]
```

Nothing was passed by reference and nothing was pointed at explicitly. `view` is
its own variable; it just happens to address elements that `s` also addresses.
This is what people mean when they say slices are "reference types", and it is
both the reason Go can pass a million-element slice for free and the reason a
function that quietly edits its argument is a bug that takes an afternoon to
find.

Two rules keep it straight.

**Writing to an element is visible everywhere.** `s[i] = v` goes through the
pointer, so every slice sharing that array sees it. That is how a function
sorts, scales or fills a slice in place with no return value at all.

**Changing the length is not.** `append` and reslicing build a new header. The
caller keeps the old one, with the old length, which is why any function that
grows a slice has to return it.

The nasty case is where the two meet:

```go
s := []int{1, 2, 3, 4, 5}
view := s[1:3]            // len 2, cap 4 - it can still see s[3] and s[4]
view = append(view, 99)   // writes into s[3]!
fmt.Println(s)            // [1 2 3 99 5]
```

`view` had spare capacity, so `append` had somewhere to put the element: the
middle of `s`. The fix is the **three-index slice**, `s[low:high:max]`, which
sets the capacity as well as the length. `s[1:3:3]` has no room left, so its
first `append` allocates instead of overwriting.

Aliasing is not something to avoid. It is something to be deliberate about: hand
out a view when you mean a view, and a copy when you mean a copy.

## Task

Fill in the five functions in `main.go`.

1. `Scale(s, factor)` multiplies in place and returns nothing.
2. `ScaledCopy(s, factor)` returns a new slice and leaves `s` alone.
3. `SplitAt(s, i)` returns two **views** of `s`, clamping `i` into range. The
   tests check that writing through them changes `s`.
4. `Window(s, lo, hi)` returns a view that is safe to `append` to: growing it
   must never overwrite an element of `s`.
5. `Dedup(s)` collapses runs of equal neighbours in a sorted slice **in place**
   and returns the shortened prefix.

## Hints

- `Scale` is `for i := range s { s[i] *= factor }`. Ranging with `for _, v :=
  range s` gives you a copy of each element, so writing to `v` does nothing.
- `ScaledCopy` needs its own array: `make([]int, len(s))`, then fill it.
- `SplitAt` is `s[:i], s[i:]` once `i` is clamped. Two slices, one array, no
  copying.
- `Window` is the three-index form: `s[lo:hi:hi]`. Clamp `lo` into `0..len(s)`
  first, then `hi` into `lo..len(s)`, or the slice expression panics.
- `Dedup` is the classic two-index sweep: keep a write position `n`, range over
  `s[1:]`, and when the element differs from `s[n-1]`, write it to `s[n]` and
  advance. Return `s[:n]`. The read is always ahead of the write, so nothing is
  clobbered before it is read.
