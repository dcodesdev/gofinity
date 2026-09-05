# Generic Map and Filter

Before Go 1.18 there were two ways to write a function that doubles every
element of a slice: write it once per element type, or take `[]any` and pay for
a type assertion on every element. Neither is good. Type parameters are the
third way.

```go
func Map[T, U any](s []T, f func(T) U) []U
```

`[T, U any]` is the type parameter list. `T` and `U` are types the caller
supplies, `any` is the constraint on each - here, no constraint at all. Inside
the body `T` and `U` behave like ordinary types you happen not to know the name
of.

## Inference

You almost never write the types at the call site:

```go
lengths := Map(words, func(s string) int { return len(s) })
```

The compiler reads `words` as `[]string`, so `T` is `string`, and the function
literal returns `int`, so `U` is `int`. Explicit instantiation - `Map[string,
int](...)` - is legal and occasionally necessary, but if you find yourself
writing it everywhere, the signature is usually the problem.

## The zero value of a type parameter

You cannot write `return nil` or `return 0` for a `T`, because `T` might be
neither. There is one spelling that always works:

```go
var zero T
return zero, false
```

That is the idiom for "no result" in a generic function, and it is why the
comma-ok pattern shows up so often here.

## What you can do with a T

Only what the constraint allows. Under `any` you can assign it, pass it, put it
in a slice, and compare it to nothing at all - not even `==`, because not every
Go type is comparable. Adding `comparable` as the constraint buys you `==` and
`!=`, which is exactly what a `Contains` or a map key needs. The next challenge
is about constraints that buy you arithmetic.

## Allocating the result

`make([]U, 0, len(s))` for a filter, `make([]U, len(s))` for a map. A nil slice
appends fine, so the length is a performance choice, not a correctness one - but
these functions return their result to a caller who may compare it, so be
deliberate: this file's functions all return an **empty non-nil** slice for an
empty input.

## Task

Implement `Map`, `Filter`, `Reduce`, `Keys`, `Contains`, `IndexOf` and `First`.
The signatures are given; the type parameter lists on some of them are not.

## Hints

- `Map` and `Filter` differ in one line: one appends `f(v)`, the other appends
  `v` when `keep(v)` is true.
- `Reduce` threads an accumulator: start at `init`, and `acc = f(acc, v)` for
  each element. Its two type parameters are the element type and the
  accumulator type, and they are not the same thing.
- `Keys` needs `K comparable`, because a map key is comparable by definition.
  The order of a map range is random, so the test sorts what you return.
- `Contains` and `IndexOf` compare elements, so their `T` is `comparable`, not
  `any`. `IndexOf` returns -1 for a miss.
- `First` returns `(T, bool)`. Use `var zero T` for the miss.
