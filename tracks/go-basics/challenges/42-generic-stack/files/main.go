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
	// TODO
	return nil
}

// Push puts v on top of the stack. The receiver is a pointer because a value
// receiver would append to a copy.
func (s *Stack[T]) Push(v T) {
	// TODO
}

// Pop removes and returns the top element, and whether there was one. An empty
// stack gives the zero value of T and false.
//
// Clear the vacated slot before reslicing, or the popped element stays
// reachable through the backing array and cannot be collected.
func (s *Stack[T]) Pop() (T, bool) {
	// TODO
	var zero T
	return zero, false
}

// Peek returns the top element without removing it, and whether there was one.
func (s *Stack[T]) Peek() (T, bool) {
	// TODO
	var zero T
	return zero, false
}

// Len returns how many elements the stack holds.
func (s *Stack[T]) Len() int {
	// TODO
	return 0
}

// IsEmpty reports whether the stack holds nothing.
func (s *Stack[T]) IsEmpty() bool {
	// TODO
	return true
}

// Clone returns an independent copy: pushing to or popping from one must not
// change the other, so the elements have to be copied rather than shared.
func (s *Stack[T]) Clone() *Stack[T] {
	// TODO
	return nil
}

// MapStack returns a new stack holding f applied to every element of s, bottom
// to top. It is a function and not a method because a method's type parameters
// can only be the receiver's, and U has nowhere else to come from.
func MapStack[T, U any](s *Stack[T], f func(T) U) *Stack[U] {
	// TODO
	return nil
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
	// TODO
	return nil
}

// PairsToMap is the inverse of Items. When two pairs share a key, the later one
// wins.
//
// An empty or nil slice gives an empty, non-nil map.
func PairsToMap[K comparable, V any](pairs []Pair[K, V]) map[K]V {
	// TODO
	return nil
}

func main() {
	var s Stack[int]
	s.Push(1)
	s.Push(2)
	fmt.Println(s.Pop())
	fmt.Println(MapStack(&s, func(n int) string { return fmt.Sprint(n) }).Len())
}
