package main

import (
	"maps"
	"slices"
	"testing"
)

func equalGroups(a, b map[string][]string) bool {
	return maps.EqualFunc(a, b, slices.Equal)
}

func TestAppend(t *testing.T) {
	groups := map[string][]string{"g": {"go"}}

	Append(groups, "g", "gopher")
	Append(groups, "r", "rust")

	want := map[string][]string{"g": {"go", "gopher"}, "r": {"rust"}}
	if !equalGroups(groups, want) {
		t.Errorf("after Append calls groups = %v, want %v", groups, want)
	}
}

func TestAppendCreatesTheEntry(t *testing.T) {
	// Appending to a key that is not there works: the missing value reads as a
	// nil slice, and append on a nil slice allocates one.
	groups := make(map[string][]string)
	Append(groups, "new", "first")
	if want := (map[string][]string{"new": {"first"}}); !equalGroups(groups, want) {
		t.Errorf("groups = %v, want %v", groups, want)
	}
}

func TestAppendOnNilMapDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Append on a nil map panicked with %v, want it to do nothing", r)
		}
	}()
	Append(nil, "k", "v")
}

func TestGroupByFirstLetter(t *testing.T) {
	got := GroupByFirstLetter([]string{"go", "Gopher", "rust", "ruby", "zig", "", "Ruby"})
	want := map[string][]string{
		"g": {"go", "Gopher"},
		"r": {"rust", "ruby", "Ruby"},
		"z": {"zig"},
	}
	if !equalGroups(got, want) {
		t.Errorf("GroupByFirstLetter = %v, want %v", got, want)
	}

	empty := GroupByFirstLetter(nil)
	if empty == nil {
		t.Fatal("GroupByFirstLetter(nil) = nil, want an empty map")
	}
	if len(empty) != 0 {
		t.Errorf("GroupByFirstLetter(nil) = %v, want an empty map", empty)
	}
}

func TestKeys(t *testing.T) {
	groups := map[string][]string{"z": {"1"}, "a": {"2"}, "m": {"3"}}
	want := []string{"a", "m", "z"}
	for range 10 {
		if got := Keys(groups); !slices.Equal(got, want) {
			t.Fatalf("Keys = %v, want %v", got, want)
		}
	}
	if got := Keys(nil); len(got) != 0 {
		t.Errorf("Keys(nil) = %v, want an empty result", got)
	}
}

func TestCount(t *testing.T) {
	groups := map[string][]string{"a": {"1", "2"}, "b": {"3"}, "c": nil}
	if got := Count(groups); got != 3 {
		t.Errorf("Count = %d, want 3", got)
	}
	if got := Count(nil); got != 0 {
		t.Errorf("Count(nil) = %d, want 0", got)
	}
}

func TestCloneIsDeep(t *testing.T) {
	original := map[string][]string{"a": {"1", "2"}, "b": {"3"}}
	clone := Clone(original)

	if !equalGroups(clone, original) {
		t.Fatalf("Clone = %v, want %v", clone, original)
	}

	// Copying the map alone copies the slice headers, which still point at the
	// original arrays. Writing through one would then be visible in the other.
	clone["a"][0] = "changed"
	if original["a"][0] != "1" {
		t.Errorf(`original["a"] = %v after writing to the clone, want ["1" "2"]`, original["a"])
	}

	clone["c"] = []string{"4"}
	if _, ok := original["c"]; ok {
		t.Error("adding a key to the clone added it to the original too")
	}
}

func TestMerge(t *testing.T) {
	a := map[string][]string{"x": {"1"}, "y": {"2"}}
	b := map[string][]string{"x": {"3"}, "z": {"4"}}

	got := Merge(a, b)
	want := map[string][]string{"x": {"1", "3"}, "y": {"2"}, "z": {"4"}}
	if !equalGroups(got, want) {
		t.Fatalf("Merge = %v, want %v", got, want)
	}

	if !equalGroups(a, map[string][]string{"x": {"1"}, "y": {"2"}}) {
		t.Errorf("Merge modified a: %v", a)
	}
	if !equalGroups(b, map[string][]string{"x": {"3"}, "z": {"4"}}) {
		t.Errorf("Merge modified b: %v", b)
	}

	// The result must own its arrays, not borrow them from a or b.
	got["y"][0] = "changed"
	if a["y"][0] != "2" {
		t.Errorf(`a["y"] = %v after writing to the result, want ["2"]`, a["y"])
	}
	got["z"][0] = "changed"
	if b["z"][0] != "4" {
		t.Errorf(`b["z"] = %v after writing to the result, want ["4"]`, b["z"])
	}
}

func TestMergeGrowsWithoutOverwriting(t *testing.T) {
	// a["k"] has spare capacity, so append(a["k"], ...) would write into a's
	// own array and hand back a view of it.
	backing := make([]string, 1, 4)
	backing[0] = "first"
	a := map[string][]string{"k": backing}
	b := map[string][]string{"k": {"second"}}

	got := Merge(a, b)
	if want := []string{"first", "second"}; !slices.Equal(got["k"], want) {
		t.Fatalf(`Merge()["k"] = %v, want %v`, got["k"], want)
	}
	got["k"][0] = "changed"
	if a["k"][0] != "first" {
		t.Errorf(`a["k"][0] = %q after writing to the result, want "first"`, a["k"][0])
	}
	if full := backing[:cap(backing)]; full[1] != "" {
		t.Errorf("Merge appended into a's backing array: %v", full)
	}
}

func TestInvert(t *testing.T) {
	groups := map[string][]string{
		"admins": {"ada", "grace"},
		"devs":   {"ada", "linus"},
		"ops":    {"grace"},
	}
	want := map[string][]string{
		"ada":   {"admins", "devs"},
		"grace": {"admins", "ops"},
		"linus": {"devs"},
	}
	for range 10 {
		if got := Invert(groups); !equalGroups(got, want) {
			t.Fatalf("Invert = %v, want %v", got, want)
		}
	}

	empty := Invert(nil)
	if empty == nil || len(empty) != 0 {
		t.Errorf("Invert(nil) = %v, want an empty map", empty)
	}
}
