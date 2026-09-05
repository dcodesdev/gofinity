package main

import (
	"sort"
	"strings"
	"testing"
)

func TestZeroValueStackIsUsable(t *testing.T) {
	var s Stack[int]
	if !s.IsEmpty() {
		t.Error("the zero value of Stack[int] is not empty, want empty")
	}
	if s.Len() != 0 {
		t.Errorf("zero value Len() = %d, want 0", s.Len())
	}
	if _, ok := s.Peek(); ok {
		t.Error("Peek on the zero value returned ok = true, want false")
	}
	if _, ok := s.Pop(); ok {
		t.Error("Pop on the zero value returned ok = true, want false")
	}
	s.Push(1)
	if s.Len() != 1 {
		t.Errorf("after one Push on the zero value, Len() = %d, want 1", s.Len())
	}
}

func TestNewStack(t *testing.T) {
	s := NewStack[string]()
	if s == nil {
		t.Fatal("NewStack returned nil, want an empty stack")
	}
	if !s.IsEmpty() || s.Len() != 0 {
		t.Errorf("NewStack is not empty: Len() = %d", s.Len())
	}
}

func TestPushPopIsLastInFirstOut(t *testing.T) {
	s := NewStack[int]()
	s.Push(1)
	s.Push(2)
	s.Push(3)

	if s.Len() != 3 {
		t.Fatalf("Len() = %d after three pushes, want 3", s.Len())
	}
	for _, want := range []int{3, 2, 1} {
		got, ok := s.Pop()
		if !ok {
			t.Fatalf("Pop returned ok = false, want %d", want)
		}
		if got != want {
			t.Errorf("Pop = %d, want %d - a stack is last in, first out", got, want)
		}
	}
	if !s.IsEmpty() {
		t.Errorf("stack is not empty after popping everything: Len() = %d", s.Len())
	}
}

func TestPushIsVisibleToTheCaller(t *testing.T) {
	// A value receiver on Push would append to a copy and this would still be 0.
	s := NewStack[string]()
	s.Push("go")
	if got, ok := s.Peek(); !ok || got != "go" {
		t.Errorf("Peek after Push = (%q, %t), want (\"go\", true) - Push needs a pointer receiver", got, ok)
	}
}

func TestPeekDoesNotRemove(t *testing.T) {
	s := NewStack[int]()
	s.Push(10)
	s.Push(20)

	got, ok := s.Peek()
	if !ok || got != 20 {
		t.Fatalf("Peek = (%d, %t), want (20, true)", got, ok)
	}
	if s.Len() != 2 {
		t.Errorf("Len() = %d after Peek, want 2 - Peek must not remove", s.Len())
	}
	again, _ := s.Peek()
	if again != 20 {
		t.Errorf("second Peek = %d, want 20", again)
	}
}

func TestPopOnEmptyReturnsTheZeroValue(t *testing.T) {
	s := NewStack[string]()
	got, ok := s.Pop()
	if ok {
		t.Error("Pop on an empty stack returned ok = true, want false")
	}
	if got != "" {
		t.Errorf("Pop on an empty stack = %q, want the zero value", got)
	}
}

func TestPopClearsTheVacatedSlot(t *testing.T) {
	// The backing array outlives the reslice. Leaving the popped element in it
	// keeps whatever it points at alive for as long as the stack does.
	s := NewStack[*string]()
	a, b := "a", "b"
	s.Push(&a)
	s.Push(&b)

	if _, ok := s.Pop(); !ok {
		t.Fatal("Pop returned ok = false, want true")
	}
	backing := s.items[:cap(s.items)]
	if backing[1] != nil {
		t.Error("the popped slot in the backing array still holds the element, want it cleared to the zero value")
	}
}

func TestStacksOfDifferentTypesAreIndependentTypes(t *testing.T) {
	ints := NewStack[int]()
	strs := NewStack[string]()
	ints.Push(1)
	strs.Push("one")

	if ints.Len() != 1 || strs.Len() != 1 {
		t.Fatalf("Len() = %d and %d, want 1 and 1", ints.Len(), strs.Len())
	}
	if got, _ := strs.Peek(); got != "one" {
		t.Errorf("Stack[string] top = %q, want \"one\"", got)
	}
}

func TestClone(t *testing.T) {
	s := NewStack[int]()
	s.Push(1)
	s.Push(2)

	c := s.Clone()
	if c == nil {
		t.Fatal("Clone returned nil")
	}
	if c.Len() != 2 {
		t.Fatalf("Clone Len() = %d, want 2", c.Len())
	}
	if got, _ := c.Peek(); got != 2 {
		t.Errorf("Clone top = %d, want 2", got)
	}

	c.Push(3)
	if s.Len() != 2 {
		t.Errorf("pushing to the clone changed the original: Len() = %d, want 2", s.Len())
	}
	if _, ok := c.Pop(); !ok {
		t.Fatal("Pop on the clone returned ok = false")
	}
	c.Pop()
	if s.Len() != 2 {
		t.Errorf("popping the clone changed the original: Len() = %d, want 2", s.Len())
	}
}

func TestCloneOfEmpty(t *testing.T) {
	s := NewStack[int]()
	c := s.Clone()
	if c == nil || !c.IsEmpty() {
		t.Errorf("Clone of an empty stack = %v, want an empty stack", c)
	}
	c.Push(1)
	if s.Len() != 0 {
		t.Errorf("the clone shares state with the original: Len() = %d, want 0", s.Len())
	}
}

func TestMapStackChangesTheElementType(t *testing.T) {
	s := NewStack[int]()
	s.Push(1)
	s.Push(2)
	s.Push(3)

	out := MapStack(s, func(n int) string { return strings.Repeat("x", n) })
	if out == nil {
		t.Fatal("MapStack returned nil")
	}
	if out.Len() != 3 {
		t.Fatalf("MapStack Len() = %d, want 3", out.Len())
	}
	// Popping walks it top to bottom, so the order check reads in reverse.
	for _, want := range []string{"xxx", "xx", "x"} {
		got, ok := out.Pop()
		if !ok || got != want {
			t.Errorf("MapStack Pop = (%q, %t), want (%q, true) - keep the order", got, ok, want)
		}
	}
}

func TestMapStackLeavesTheSourceAlone(t *testing.T) {
	s := NewStack[int]()
	s.Push(7)
	out := MapStack(s, func(n int) int { return n * 2 })
	out.Push(99)

	if s.Len() != 1 {
		t.Errorf("MapStack changed the source: Len() = %d, want 1", s.Len())
	}
	if got, _ := s.Peek(); got != 7 {
		t.Errorf("source top = %d, want 7", got)
	}
}

func TestMapStackOfEmpty(t *testing.T) {
	s := NewStack[int]()
	out := MapStack(s, func(n int) bool { return n > 0 })
	if out == nil {
		t.Fatal("MapStack of an empty stack returned nil, want an empty stack")
	}
	if !out.IsEmpty() {
		t.Errorf("MapStack of an empty stack has Len() = %d, want 0", out.Len())
	}
}

func TestItems(t *testing.T) {
	got := Items(map[string]int{"a": 1, "b": 2})
	if len(got) != 2 {
		t.Fatalf("Items = %v, want 2 pairs", got)
	}
	sort.Slice(got, func(i, j int) bool { return got[i].Key < got[j].Key })
	if got[0].Key != "a" || got[0].Value != 1 {
		t.Errorf("Items[0] = %+v, want {Key:a Value:1}", got[0])
	}
	if got[1].Key != "b" || got[1].Value != 2 {
		t.Errorf("Items[1] = %+v, want {Key:b Value:2}", got[1])
	}
}

func TestItemsOfEmptyMapIsEmptyNotNil(t *testing.T) {
	got := Items(map[int]string{})
	if got == nil {
		t.Fatal("Items of an empty map returned nil, want an empty non-nil slice")
	}
	if len(got) != 0 {
		t.Errorf("Items of an empty map = %v, want length 0", got)
	}
	var m map[int]string
	if got := Items(m); got == nil || len(got) != 0 {
		t.Errorf("Items of a nil map = %v, want an empty non-nil slice", got)
	}
}

func TestPairsToMap(t *testing.T) {
	got := PairsToMap([]Pair[string, int]{{Key: "a", Value: 1}, {Key: "b", Value: 2}})
	if len(got) != 2 || got["a"] != 1 || got["b"] != 2 {
		t.Errorf("PairsToMap = %v, want map[a:1 b:2]", got)
	}
}

func TestPairsToMapLastKeyWins(t *testing.T) {
	got := PairsToMap([]Pair[string, int]{{Key: "a", Value: 1}, {Key: "a", Value: 9}})
	if len(got) != 1 || got["a"] != 9 {
		t.Errorf("PairsToMap with a duplicate key = %v, want map[a:9]", got)
	}
}

func TestPairsToMapOfEmptyIsEmptyNotNil(t *testing.T) {
	got := PairsToMap([]Pair[int, bool]{})
	if got == nil {
		t.Fatal("PairsToMap of an empty slice returned nil, want an empty non-nil map")
	}
	if len(got) != 0 {
		t.Errorf("PairsToMap of an empty slice = %v, want length 0", got)
	}
	if got := PairsToMap([]Pair[int, bool](nil)); got == nil {
		t.Error("PairsToMap of a nil slice returned nil, want an empty non-nil map")
	}
}

func TestRoundTripThroughPairs(t *testing.T) {
	want := map[string]int{"go": 1, "is": 2, "fun": 3}
	got := PairsToMap(Items(want))
	if len(got) != len(want) {
		t.Fatalf("round trip = %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("round trip[%q] = %d, want %d", k, got[k], v)
		}
	}
}
