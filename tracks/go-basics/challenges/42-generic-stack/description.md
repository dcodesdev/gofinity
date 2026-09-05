# Generic Stack

Functions were the easy half. A **type** can take type parameters too, and that
is where generics stop being a convenience and start being a data structure.

```go
type Stack[T any] struct {
	items []T
}
```

`Stack` on its own is not a type any more, it is a generic type. `Stack[int]`
and `Stack[string]` are two distinct types, and neither is assignable to the
other. Every use of the name needs its argument: a field of type `Stack[T]`, a
parameter of type `*Stack[string]`, a `make(map[string]*Stack[int])`.

## Methods on a generic type

The receiver repeats the type parameters, declaring them for the method body:

```go
func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }
```

The `[T]` on the receiver is a **declaration**, not a use - the name is yours to
pick, though matching the type is the only sane choice. The constraint is not
repeated; it came with the type.

## Methods cannot add type parameters

This is the rule that surprises everybody:

```go
func (s *Stack[T]) Map[U any](f func(T) U) *Stack[U]  // does not compile
```

A method's type parameters can only be the receiver's. There is no
`Stack[int].Map` to a `Stack[string]`, because that `U` has nowhere to come
from. The workaround is not a workaround, it is the answer: make it a function.

```go
func MapStack[T, U any](s *Stack[T], f func(T) U) *Stack[U]
```

That is why [`slices.SortFunc`](https://pkg.go.dev/slices#SortFunc) and
[`maps.Keys`](https://pkg.go.dev/maps#Keys) are package-level functions rather
than methods, and why an idiomatic generic container has a small method set with
free functions around it.

## The zero value

`var s Stack[int]` is a usable empty stack, because its only field is a nil
slice and appending to one works. Design for that: a constructor should be a
convenience, not a requirement, and no method should panic on the zero value.
`Pop` on an empty stack returns `(zero, false)`, never a panic.

## Popping without leaking

```go
v := s.items[len(s.items)-1]
s.items = s.items[:len(s.items)-1]
```

Reslicing does not shrink the backing array, so that popped element is still
referenced and cannot be collected. When `T` holds a pointer, a long-lived stack
that grows and shrinks keeps every object it ever held alive. One line fixes it:

```go
var zero T
s.items[len(s.items)-1] = zero
```

## Task

Implement `Stack[T]` with `Push`, `Pop`, `Peek`, `Len`, `IsEmpty` and `Clone`,
the free functions `NewStack` and `MapStack`, and the `Pair[K, V]` type with
`Items` and `PairsToMap`.

## Hints

- Every method takes a pointer receiver. `Push` on a value receiver appends to a
  copy, and the caller sees nothing.
- `Pop` returns `(T, bool)`. Clear the vacated slot before reslicing.
- `Peek` reads the top without removing it, same `(T, bool)` shape.
- `Clone` returns a `*Stack[T]` whose slice is a copy: pushing to the clone must
  not touch the original. `append([]T(nil), s.items...)` or `make` plus `copy`.
- `MapStack` is a function precisely because `U` cannot be introduced by a
  method. It preserves order, bottom to top.
- `Items` returns `[]Pair[K, V]` in unspecified order - the test sorts it.
  `Pair` is a generic **struct**, so its fields are `Key K` and `Value V`.
- `PairsToMap` is the inverse. A later pair with the same key wins.
