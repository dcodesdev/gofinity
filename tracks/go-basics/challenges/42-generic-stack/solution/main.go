package main

import "fmt"

// Stack is a last-in-first-out stack of any element type. Its zero value -
// var s Stack[int] - is an empty, usable stack, because a nil slice appends
// fine. No method may panic on it.
type Stack[T any] struct {
	items []T
}

// NewStack returns an empty stack. It is a convenience, not a requirement: the
// zero value works just as well.
func NewStack[T any]() *Stack[T] {
	return &Stack[T]{}
}

// Push puts v on top of the stack. The receiver is a pointer because a value
// receiver would append to a copy.
func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

// Pop removes and returns the top element, and whether there was one. An empty
// stack gives the zero value of T and false.
//
// Clear the vacated slot before reslicing, or the popped element stays
// reachable through the backing array and cannot be collected.
func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	last := len(s.items) - 1
	v := s.items[last]
	var zero T
	s.items[last] = zero
	s.items = s.items[:last]
	return v, true
}

// Peek returns the top element without removing it, and whether there was one.
func (s *Stack[T]) Peek() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

// Len returns how many elements the stack holds.
func (s *Stack[T]) Len() int {
	return len(s.items)
}

// IsEmpty reports whether the stack holds nothing.
func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}

// Clone returns an independent copy: pushing to or popping from one must not
// change the other, so the elements have to be copied rather than shared.
func (s *Stack[T]) Clone() *Stack[T] {
	out := &Stack[T]{items: make([]T, len(s.items))}
	copy(out.items, s.items)
	return out
}

// MapStack returns a new stack holding f applied to every element of s, bottom
// to top. It is a function and not a method because a method's type parameters
// can only be the receiver's, and U has nowhere else to come from.
func MapStack[T, U any](s *Stack[T], f func(T) U) *Stack[U] {
	out := &Stack[U]{items: make([]U, 0, len(s.items))}
	for _, v := range s.items {
		out.items = append(out.items, f(v))
	}
	return out
}

// Pair is one key and one value. A generic struct: the type parameters are used
// as the field types.
type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

// Items returns the entries of m as pairs, in unspecified order.
//
// An empty or nil map gives an empty, non-nil slice.
func Items[K comparable, V any](m map[K]V) []Pair[K, V] {
	out := make([]Pair[K, V], 0, len(m))
	for k, v := range m {
		out = append(out, Pair[K, V]{Key: k, Value: v})
	}
	return out
}

// PairsToMap is the inverse of Items. When two pairs share a key, the later one
// wins.
//
// An empty or nil slice gives an empty, non-nil map.
func PairsToMap[K comparable, V any](pairs []Pair[K, V]) map[K]V {
	out := make(map[K]V, len(pairs))
	for _, p := range pairs {
		out[p.Key] = p.Value
	}
	return out
}

func main() {
	var s Stack[int]
	s.Push(1)
	s.Push(2)
	fmt.Println(s.Pop())
	fmt.Println(MapStack(&s, func(n int) string { return fmt.Sprint(n) }).Len())
}
