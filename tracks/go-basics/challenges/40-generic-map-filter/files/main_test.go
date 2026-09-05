package main

import (
	"sort"
	"strings"
	"testing"
)

func TestMapChangesTheElementType(t *testing.T) {
	got := Map([]string{"go", "is", "fun"}, func(s string) int { return len(s) })
	want := []int{2, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("Map lengths = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Map(...)[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestMapKeepsOrderAndSameType(t *testing.T) {
	got := Map([]int{1, 2, 3}, func(n int) int { return n * n })
	want := []int{1, 4, 9}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Map(squares)[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestMapOfEmptyIsEmptyNotNil(t *testing.T) {
	got := Map([]int{}, func(n int) string { return "x" })
	if got == nil {
		t.Fatal("Map of an empty slice returned nil, want an empty non-nil slice")
	}
	if len(got) != 0 {
		t.Errorf("Map of an empty slice = %v, want length 0", got)
	}
}

func TestFilterKeepsMatchesInOrder(t *testing.T) {
	got := Filter([]int{1, 2, 3, 4, 5, 6}, func(n int) bool { return n%2 == 0 })
	want := []int{2, 4, 6}
	if len(got) != len(want) {
		t.Fatalf("Filter = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Filter(...)[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestFilterOnStrings(t *testing.T) {
	got := Filter([]string{"go", "rust", "gopher"}, func(s string) bool {
		return strings.HasPrefix(s, "go")
	})
	if len(got) != 2 || got[0] != "go" || got[1] != "gopher" {
		t.Errorf("Filter(strings) = %v, want [go gopher]", got)
	}
}

func TestFilterKeepingNothingIsEmptyNotNil(t *testing.T) {
	got := Filter([]int{1, 3, 5}, func(n int) bool { return n%2 == 0 })
	if got == nil {
		t.Fatal("Filter keeping nothing returned nil, want an empty non-nil slice")
	}
	if len(got) != 0 {
		t.Errorf("Filter keeping nothing = %v, want length 0", got)
	}
}

func TestFilterDoesNotAliasTheInput(t *testing.T) {
	in := []int{2, 4, 6}
	got := Filter(in, func(n int) bool { return true })
	if len(got) != 3 {
		t.Fatalf("Filter kept %d elements, want 3", len(got))
	}
	got[0] = 99
	if in[0] != 2 {
		t.Errorf("writing to the result changed the input: in[0] = %d, want 2", in[0])
	}
}

func TestReduceSums(t *testing.T) {
	got := Reduce([]int{1, 2, 3, 4}, 0, func(acc, n int) int { return acc + n })
	if got != 10 {
		t.Errorf("Reduce(sum) = %d, want 10", got)
	}
}

func TestReduceStartsAtInit(t *testing.T) {
	got := Reduce([]int{}, 7, func(acc, n int) int { return acc + n })
	if got != 7 {
		t.Errorf("Reduce over an empty slice = %d, want the init value 7", got)
	}
}

func TestReduceAcrossTypes(t *testing.T) {
	got := Reduce([]string{"go", "is", "fun"}, 0, func(acc int, s string) int {
		return acc + len(s)
	})
	if got != 7 {
		t.Errorf("Reduce(total length) = %d, want 7", got)
	}
}

func TestReduceIsLeftToRight(t *testing.T) {
	got := Reduce([]string{"a", "b", "c"}, "", func(acc, s string) string { return acc + s })
	if got != "abc" {
		t.Errorf("Reduce(concat) = %q, want %q - fold left to right", got, "abc")
	}
}

func TestKeys(t *testing.T) {
	got := Keys(map[string]int{"a": 1, "b": 2, "c": 3})
	sort.Strings(got)
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("Keys = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Keys sorted[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestKeysOfEmptyMapIsEmptyNotNil(t *testing.T) {
	got := Keys(map[int]string{})
	if got == nil {
		t.Fatal("Keys of an empty map returned nil, want an empty non-nil slice")
	}
	if len(got) != 0 {
		t.Errorf("Keys of an empty map = %v, want length 0", got)
	}
}

func TestKeysOfNilMap(t *testing.T) {
	var m map[string]bool
	got := Keys(m)
	if got == nil || len(got) != 0 {
		t.Errorf("Keys of a nil map = %v (nil: %t), want an empty non-nil slice", got, got == nil)
	}
}

func TestContains(t *testing.T) {
	s := []string{"go", "is", "fun"}
	if !Contains(s, "is") {
		t.Error(`Contains(s, "is") = false, want true`)
	}
	if Contains(s, "rust") {
		t.Error(`Contains(s, "rust") = true, want false`)
	}
	if Contains([]int{}, 1) {
		t.Error("Contains over an empty slice = true, want false")
	}
}

func TestIndexOf(t *testing.T) {
	s := []int{10, 20, 30, 20}
	if got := IndexOf(s, 20); got != 1 {
		t.Errorf("IndexOf(s, 20) = %d, want 1 - the first match", got)
	}
	if got := IndexOf(s, 10); got != 0 {
		t.Errorf("IndexOf(s, 10) = %d, want 0", got)
	}
	if got := IndexOf(s, 99); got != -1 {
		t.Errorf("IndexOf(s, 99) = %d, want -1", got)
	}
	if got := IndexOf([]string{}, "x"); got != -1 {
		t.Errorf("IndexOf over an empty slice = %d, want -1", got)
	}
}

func TestFirstFinds(t *testing.T) {
	got, ok := First([]int{1, 3, 4, 6}, func(n int) bool { return n%2 == 0 })
	if !ok {
		t.Fatal("First returned ok = false, want true")
	}
	if got != 4 {
		t.Errorf("First(even) = %d, want 4", got)
	}
}

func TestFirstMissReturnsTheZeroValue(t *testing.T) {
	got, ok := First([]string{"go", "is"}, func(s string) bool { return s == "rust" })
	if ok {
		t.Fatal("First returned ok = true for a miss, want false")
	}
	if got != "" {
		t.Errorf("First miss = %q, want the zero value %q", got, "")
	}

	n, ok := First([]int{}, func(n int) bool { return true })
	if ok || n != 0 {
		t.Errorf("First over an empty slice = (%d, %t), want (0, false)", n, ok)
	}
}

func TestGenericsComposeOverAStructType(t *testing.T) {
	type user struct {
		Name string
		Age  int
	}
	users := []user{{"ada", 36}, {"alan", 41}, {"grace", 45}}

	names := Map(Filter(users, func(u user) bool { return u.Age > 40 }), func(u user) string {
		return u.Name
	})
	if len(names) != 2 || names[0] != "alan" || names[1] != "grace" {
		t.Errorf("Map(Filter(users)) = %v, want [alan grace]", names)
	}

	total := Reduce(users, 0, func(acc int, u user) int { return acc + u.Age })
	if total != 122 {
		t.Errorf("Reduce(total age) = %d, want 122", total)
	}
}
