# Slice Basics

Go has two list types, and almost all of the confusion about the second one
comes from not having met the first.

An **array** has its length in its type. `[5]int` and `[6]int` are different
types, and an array is a **value**: assigning one, or passing one to a function,
copies every element. That makes arrays predictable and mostly unused.

A **slice** is a view onto an array. It is three words - a pointer to some
element of a backing array, a **length**, and a **capacity** - and it is what
you write in real Go:

```go
s := []int{1, 2, 3}     // a slice, no length in the type
a := [3]int{1, 2, 3}    // an array, length 3 forever
```

`len(s)` is how many elements you can index. `cap(s)` is how many the backing
array has room for from the start of the slice onward. They are usually
different, and the gap between them is what the next challenge is about.

Slicing takes a view of a view:

```go
s := []int{1, 2, 3, 4, 5}
s[1:3]   // [2 3] - low bound included, high bound excluded
s[:2]    // [1 2] - low defaults to 0
s[2:]    // [3 4 5] - high defaults to len(s)
```

The bounds are indices, not counts, and `s[low:high]` panics unless
`0 <= low <= high <= cap(s)`. Clamping the inputs before you slice is ordinary
Go, not defensive programming.

## Task

Fill in the seven functions in `main.go`.

1. `Describe(s)` returns `"len=3 cap=5"` for a slice of length 3 and capacity 5.
2. `First(s)` and `Last(s)` return an element and `true`, or `0, false` when
   there is nothing to return.
3. `Head(s, n)` returns the first `n` elements and `Tail(s, n)` the last `n`,
   both clamping `n` into range rather than panicking.
4. `SumArray(a)` adds up a `[5]int`, and `Doubled(a)` returns a `[5]int` with
   every element doubled, leaving the caller's array alone.

## Hints

- `Describe` is one [`fmt.Sprintf`](https://pkg.go.dev/fmt#Sprintf) over `len`
  and `cap`.
- `First` and `Last` are the comma-ok shape from the functions lesson. Guard on
  `len(s) == 0`, which is true for a nil slice too, so nil needs no separate
  case.
- The last element is `s[len(s)-1]`. There is no negative indexing in Go.
- Clamp `n` into `0..len(s)` first, then slice once: `s[:n]` for the head,
  `s[len(s)-n:]` for the tail.
- `Doubled` can write straight into its parameter. The array was copied when the
  call was made, so the loop is editing your own copy - and the test checks that.
