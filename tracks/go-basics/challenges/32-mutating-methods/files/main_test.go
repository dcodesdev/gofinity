package main

import (
	"slices"
	"testing"
)

func TestZeroStackIsUsable(t *testing.T) {
	// No constructor: append to a nil slice allocates one.
	var s Stack
	if s.Len() != 0 {
		t.Fatalf("zero Stack Len = %d, want 0", s.Len())
	}
	if _, ok := s.Pop(); ok {
		t.Error("Pop on an empty stack reported ok, want false")
	}
	if v, ok := s.Peek(); ok || v != 0 {
		t.Errorf("Peek on an empty stack = (%d, %v), want (0, false)", v, ok)
	}

	s.Push(7)
	if s.Len() != 1 {
		t.Errorf("after one Push, Len = %d, want 1", s.Len())
	}
}

func TestPushPopPeek(t *testing.T) {
	var s Stack
	s.PushAll(1, 2, 3)
	if s.Len() != 3 {
		t.Fatalf("Len after PushAll(1, 2, 3) = %d, want 3", s.Len())
	}

	if v, ok := s.Peek(); !ok || v != 3 {
		t.Errorf("Peek = (%d, %v), want (3, true) - the last value pushed is on top", v, ok)
	}
	if s.Len() != 3 {
		t.Errorf("Peek removed a value: Len = %d, want 3", s.Len())
	}

	for _, want := range []int{3, 2, 1} {
		v, ok := s.Pop()
		if !ok || v != want {
			t.Fatalf("Pop = (%d, %v), want (%d, true)", v, ok, want)
		}
	}
	if s.Len() != 0 {
		t.Errorf("Len after popping everything = %d, want 0", s.Len())
	}
	if _, ok := s.Pop(); ok {
		t.Error("Pop past the bottom reported ok, want false")
	}
}

func TestPushAllEmpty(t *testing.T) {
	var s Stack
	s.PushAll()
	if s.Len() != 0 {
		t.Errorf("PushAll() with no arguments pushed something: Len = %d", s.Len())
	}
}

func TestLenTakesACopyAndDoesNotMutate(t *testing.T) {
	var s Stack
	s.PushAll(1, 2)
	_ = s.Len()
	if got, ok := s.Peek(); !ok || got != 2 {
		t.Errorf("Len disturbed the stack: Peek = (%d, %v), want (2, true)", got, ok)
	}
}

func TestDrain(t *testing.T) {
	var s Stack
	s.PushAll(1, 2, 3)
	got := s.Drain()
	if want := []int{3, 2, 1}; !slices.Equal(got, want) {
		t.Errorf("Drain = %v, want %v - top first", got, want)
	}
	if s.Len() != 0 {
		t.Errorf("Drain left %d values behind, want an empty stack", s.Len())
	}

	var empty Stack
	if got := empty.Drain(); len(got) != 0 {
		t.Errorf("Drain of an empty stack = %v, want nothing", got)
	}
}

func TestDrainAll(t *testing.T) {
	stacks := make([]Stack, 3)
	stacks[0].PushAll(1, 2)
	stacks[1].PushAll(3)
	// stacks[2] is left empty on purpose.

	got := DrainAll(stacks)
	if want := []int{2, 1, 3}; !slices.Equal(got, want) {
		t.Errorf("DrainAll = %v, want %v", got, want)
	}
	for i := range stacks {
		if n := stacks[i].Len(); n != 0 {
			t.Errorf("stacks[%d] still holds %d values - Drain was called on a range copy", i, n)
		}
	}

	if got := DrainAll(nil); len(got) != 0 {
		t.Errorf("DrainAll(nil) = %v, want nothing", got)
	}
}

func TestNewTally(t *testing.T) {
	tally := NewTally()
	if tally == nil {
		t.Fatal("NewTally returned nil")
	}
	if tally.Total() != 0 {
		t.Errorf("a fresh Tally has Total %d, want 0", tally.Total())
	}
	if got := tally.Count("missing"); got != 0 {
		t.Errorf("Count of an unseen name = %d, want 0", got)
	}
}

func TestTallyAddAndCount(t *testing.T) {
	tally := NewTally()
	tally.Add("go")
	tally.Add("rust")
	tally.Add("go")

	if got := tally.Count("go"); got != 2 {
		t.Errorf("Count(go) = %d, want 2", got)
	}
	if got := tally.Count("rust"); got != 1 {
		t.Errorf("Count(rust) = %d, want 1", got)
	}
	if got := tally.Total(); got != 3 {
		t.Errorf("Total = %d, want 3", got)
	}
}

func TestZeroTallyDoesNotPanic(t *testing.T) {
	// The map field is nil here. Reading a nil map is fine, writing to one
	// panics, so Add has to make the map before it writes.
	var tally Tally
	if got := tally.Count("go"); got != 0 {
		t.Errorf("Count on a zero Tally = %d, want 0", got)
	}
	tally.Add("go")
	if got := tally.Count("go"); got != 1 {
		t.Errorf("Count after Add on a zero Tally = %d, want 1", got)
	}
}

func TestMerge(t *testing.T) {
	a := NewTally()
	a.Add("go")
	a.Add("go")

	b := NewTally()
	b.Add("go")
	b.Add("zig")

	a.Merge(b)
	if got := a.Count("go"); got != 3 {
		t.Errorf("after Merge, Count(go) = %d, want 3", got)
	}
	if got := a.Count("zig"); got != 1 {
		t.Errorf("after Merge, Count(zig) = %d, want 1", got)
	}
	if got := a.Total(); got != 4 {
		t.Errorf("after Merge, Total = %d, want 4", got)
	}

	if got := b.Total(); got != 2 {
		t.Errorf("Merge changed the source Tally: Total = %d, want 2", got)
	}
	if got := b.Count("go"); got != 1 {
		t.Errorf("Merge changed the source Tally: Count(go) = %d, want 1", got)
	}

	// Merging into a zero Tally has to make the map first.
	var into Tally
	into.Merge(b)
	if got := into.Total(); got != 2 {
		t.Errorf("merging into a zero Tally gave Total %d, want 2", got)
	}

	into.Merge(nil)
	if got := into.Total(); got != 2 {
		t.Errorf("Merge(nil) changed Total to %d, want 2", got)
	}
}
