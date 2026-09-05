# Slices and Arrays

Go has one list type you will write every day and one you will barely write at
all. Learning the rare one first is the fastest way to understand the common
one, because the common one is built out of it.

## Arrays

An array carries its length in its **type**:

```go
var a [5]int
b := [3]string{"x", "y", "z"}
c := [...]int{1, 2, 3}   // length 3, counted for you
```

`[5]int` and `[6]int` are different types. A function taking a `[5]int` cannot
be handed a `[6]int`, and there is no way to write "an array of any length".

More importantly, an array is a **value**. Assigning one copies every element,
and so does passing one to a function:

```go
a := [3]int{1, 2, 3}
b := a
b[0] = 99
fmt.Println(a, b)   // [1 2 3] [99 2 3]
```

Arrays are comparable with `==` when their element type is, which slices are
not. Beyond fixed-size buffers, hash digests and the occasional lookup table,
you will not reach for them. What they are is the thing a slice points at.

## Slices

A slice is written with no length in the type:

```go
s := []int{1, 2, 3}
```

Under it is a header of exactly three words:

- a **pointer** to an element of some backing array,
- a **length**, how many elements you can index,
- a **capacity**, how many the array has room for from that pointer onward.

`len(s)` and `cap(s)` read the second and third. The zero value is `nil`: a nil
pointer, length 0, capacity 0. A nil slice is a perfectly good empty slice - you
can range it, `len` it and `append` to it - which is why `var out []int` is the
normal way to start accumulating.

`make` builds one with a length, and optionally a capacity beyond it:

```go
make([]int, 3)      // len 3, cap 3, all zero
make([]int, 0, 8)   // len 0, cap 8 - empty, but with room prepared
```

## Slicing

`s[low:high]` takes a view. The low bound is included, the high bound is not,
both are indices rather than counts, and either can be left off:

```go
s := []int{1, 2, 3, 4, 5}
s[1:3]   // [2 3]
s[:2]    // [1 2]
s[2:]    // [3 4 5]
s[:]     // all of it
```

It panics unless `0 <= low <= high <= cap(s)`. Note `cap`, not `len`: you may
reslice **past** the current length, back into capacity the slice already owns.
That is legal and occasionally useful, and it is the first hint that a slice can
see more memory than its length admits.

Slicing allocates nothing. The result is a new header pointing into the same
array, which is why a view is free at any size.

## Append

`append` lives in the gap between length and capacity:

```go
s := make([]int, 0, 4)
s = append(s, 1)   // room to spare: writes into the array, len becomes 1
```

When there is capacity left, `append` writes into the existing array and returns
a slice one longer. When there is not, it allocates a **bigger array**, copies
everything over, and returns a slice pointing at the new one. The old array
stays exactly as it was, still seen by any slice that was looking at it.

So `append` returns a value, and using it is not optional:

```go
s = append(s, 1)   // right
append(s, 1)       // does not compile: the result is unused
```

You can append several at once, and spread a slice with `...`:

```go
s = append(s, 1, 2, 3)
s = append(s, other...)
```

The growth factor is deliberately unspecified. Do not write code that depends on
what `cap` will be after a particular append. Do use the capacity argument to
`make` when you know the answer's size: `make([]T, 0, n)` then `append` avoids
every intermediate array.

## Copy

`copy(dst, src)` copies element by element, stops at the shorter of the two, and
returns how many it moved. It never resizes anything:

```go
dst := make([]int, 2)
n := copy(dst, []int{1, 2, 3})   // n is 2, dst is [1 2]
```

The classic mistake is copying into a slice with capacity but no length.
`copy(make([]int, 0, 10), src)` copies **nothing**, because `dst` has length 0.
`copy` respects length; `append` uses capacity.

Together, `make` plus `copy` is an independent duplicate, and `make` with a
capacity plus `append` is a result built to a known size. (`slices.Clone` in the
standard library is the first of those, written for you.)

## Aliasing, and the bug it causes

Because a slice value is just that three-word header, copying a slice copies the
header and not the array. Two slices can be entirely separate values that
describe the same memory.

Writing to an element goes through the pointer, so it is visible to everyone:

```go
func Scale(s []int, f int) {
	for i := range s {
		s[i] *= f
	}
}
```

No pointer syntax, no return value, and the caller sees the change. That is
normal Go, and it is why `sort.Ints` and friends need no result.

Changing the **length**, though, is not visible, because the length lives in the
caller's own copy of the header. That asymmetry is the whole rule: a function
may edit elements in place, but if it grows or shrinks a slice it must return
the new header.

Now the case where both meet:

```go
s := []int{1, 2, 3, 4, 5}
view := s[1:3]            // len 2, cap 4 - it can still reach s[3] and s[4]
view = append(view, 99)   // writes into s[3]
fmt.Println(s)            // [1 2 3 99 5]
```

Nothing here is a bug in Go. `view` had spare capacity, so `append` used it, and
that capacity is the middle of `s`. The same thing bites in the other direction:

```go
func Concat(a, b []int) []int {
	return append(a, b...)   // reuses a's array whenever it has room
}
```

Sometimes that returns a fresh slice and sometimes it returns a view of `a`,
depending on a capacity the caller never thinks about. Code whose aliasing
depends on the input is code that works in tests and corrupts data in
production.

Two fixes, and which one you want depends on what you meant.

**Allocate deliberately** when the result should be independent:

```go
out := make([]int, 0, len(a)+len(b))
out = append(out, a...)
out = append(out, b...)
```

**Cap the view** when you are handing out a window into something you own. The
three-index slice `s[low:high:max]` sets the capacity as well as the length:

```go
view := s[1:3:3]          // len 2, cap 2
view = append(view, 99)   // no room, so a new array; s is untouched
```

That form exists for exactly this: a method returning part of its own buffer
should return `b.buf[i:j:j]`, so a caller who appends cannot reach back into the
rest of the buffer.

The last consequence is lifetime. A one-element slice of a huge array keeps that
whole array alive, because the pointer still points into it. If you are holding
a small piece of something large for a long time, copy it out.

## Further reading

- [Go slices: usage and internals](https://go.dev/blog/slices-intro) - the
  pointer, length and capacity behind the slice header, with pictures.
- [The mechanics of append](https://go.dev/blog/slices) - why `append` returns
  a slice, and when it reuses the backing array instead of allocating.
- [Slice types](https://go.dev/ref/spec#Slice_types) - the spec on slicing bounds,
  `len`, `cap` and the three-index form.
- [Appending to and copying slices](https://go.dev/ref/spec#Appending_and_copying_slices):
  the exact contract for `append` and `copy`.
- [slices](https://pkg.go.dev/slices) - the standard library helpers, including
  `Clone`, `Insert`, `Delete` and `Compact`.

## Practise

Three challenges. The first is the shape of the thing: `len` against `cap`,
slicing with clamped bounds, and an array proving it copies. The second is
`append` and `copy` - returning the grown slice, cloning independently, building
insert and remove, and a `Concat` whose test only passes if you allocate rather
than reuse. The third is aliasing head on: in place versus copied, two views of
one array, a window it is safe to append to, and an in-place dedup that returns
a prefix of the slice it was given.
