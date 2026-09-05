package main

import (
	"maps"
	"slices"
	"testing"
)

func TestLookup(t *testing.T) {
	m := map[string]int{"a": 1, "zero": 0}

	if got, ok := Lookup(m, "a"); got != 1 || !ok {
		t.Errorf(`Lookup(m, "a") = %d, %v, want 1, true`, got, ok)
	}
	// A stored zero is not a missing key. Only the second result can tell them
	// apart, which is why the comma-ok form exists.
	if got, ok := Lookup(m, "zero"); got != 0 || !ok {
		t.Errorf(`Lookup(m, "zero") = %d, %v, want 0, true`, got, ok)
	}
	if got, ok := Lookup(m, "missing"); got != 0 || ok {
		t.Errorf(`Lookup(m, "missing") = %d, %v, want 0, false`, got, ok)
	}
	// Reading a nil map is legal and behaves like an empty one.
	if got, ok := Lookup(nil, "a"); got != 0 || ok {
		t.Errorf(`Lookup(nil, "a") = %d, %v, want 0, false`, got, ok)
	}
}

func TestGet(t *testing.T) {
	m := map[string]int{"a": 1, "zero": 0}

	if got := Get(m, "a", -1); got != 1 {
		t.Errorf(`Get(m, "a", -1) = %d, want 1`, got)
	}
	if got := Get(m, "zero", -1); got != 0 {
		t.Errorf(`Get(m, "zero", -1) = %d, want 0`, got)
	}
	if got := Get(m, "missing", -1); got != -1 {
		t.Errorf(`Get(m, "missing", -1) = %d, want -1`, got)
	}
	if got := Get(nil, "a", 42); got != 42 {
		t.Errorf(`Get(nil, "a", 42) = %d, want 42`, got)
	}
}

func TestAdd(t *testing.T) {
	m := map[string]int{"go": 3}

	Add(m, "go", 1)
	Add(m, "rust", 2)
	Add(m, "rust", -2)

	want := map[string]int{"go": 4, "rust": 0}
	if !maps.Equal(m, want) {
		t.Errorf("after Add calls m = %v, want %v", m, want)
	}
}

func TestAddSeenByCaller(t *testing.T) {
	// A map value is a reference to the same table, so a function that writes
	// into it needs no pointer and no return value.
	m := make(map[string]int)
	Add(m, "x", 5)
	if m["x"] != 5 {
		t.Errorf(`m["x"] = %d after Add, want 5 - the caller must see the write`, m["x"])
	}
}

func TestRemove(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}

	if !Remove(m, "a") {
		t.Error(`Remove(m, "a") = false, want true`)
	}
	if _, ok := m["a"]; ok {
		t.Error(`"a" is still in m after Remove`)
	}
	if Remove(m, "a") {
		t.Error(`Remove(m, "a") a second time = true, want false`)
	}
	if Remove(m, "nope") {
		t.Error(`Remove(m, "nope") = true, want false`)
	}
	if len(m) != 1 {
		t.Errorf("len(m) = %d after removing one of two keys, want 1", len(m))
	}
}

func TestSortedKeys(t *testing.T) {
	m := map[string]int{"pear": 1, "apple": 2, "fig": 3, "banana": 4}
	want := []string{"apple", "banana", "fig", "pear"}

	// Ten runs, because a single one can match a random range order by luck.
	for range 10 {
		if got := SortedKeys(m); !slices.Equal(got, want) {
			t.Fatalf("SortedKeys(m) = %v, want %v", got, want)
		}
	}
	if got := SortedKeys(map[string]int{}); len(got) != 0 {
		t.Errorf("SortedKeys(empty) = %v, want an empty result", got)
	}
	if got := SortedKeys(nil); len(got) != 0 {
		t.Errorf("SortedKeys(nil) = %v, want an empty result", got)
	}
}

func TestTotal(t *testing.T) {
	if got := Total(map[string]int{"a": 1, "b": 2, "c": 3}); got != 6 {
		t.Errorf("Total = %d, want 6", got)
	}
	if got := Total(map[string]int{"a": 5, "b": -5}); got != 0 {
		t.Errorf("Total = %d, want 0", got)
	}
	if got := Total(nil); got != 0 {
		t.Errorf("Total(nil) = %d, want 0", got)
	}
}

func TestFromPairs(t *testing.T) {
	got := FromPairs([]string{"a", "b"}, []int{1, 2})
	if want := (map[string]int{"a": 1, "b": 2}); !maps.Equal(got, want) {
		t.Errorf("FromPairs = %v, want %v", got, want)
	}

	// A later duplicate overwrites the earlier one: assignment, not insertion.
	got = FromPairs([]string{"a", "a"}, []int{1, 9})
	if want := (map[string]int{"a": 9}); !maps.Equal(got, want) {
		t.Errorf("FromPairs with a duplicate key = %v, want %v", got, want)
	}

	if got := FromPairs([]string{"a", "b"}, []int{1}); got != nil {
		t.Errorf("FromPairs with mismatched lengths = %v, want nil", got)
	}

	empty := FromPairs(nil, nil)
	if empty == nil {
		t.Fatal("FromPairs(nil, nil) = nil, want an empty map - the lengths match")
	}
	if len(empty) != 0 {
		t.Errorf("FromPairs(nil, nil) = %v, want an empty map", empty)
	}
	// The result must be usable, which a nil map would not be.
	empty["x"] = 1
}
