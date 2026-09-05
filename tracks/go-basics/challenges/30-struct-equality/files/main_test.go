package main

import (
	"slices"
	"testing"
)

func TestSamePoint(t *testing.T) {
	if !SamePoint(Point{1, 2}, Point{1, 2}) {
		t.Error("SamePoint({1,2}, {1,2}) = false, want true")
	}
	if SamePoint(Point{1, 2}, Point{2, 1}) {
		t.Error("SamePoint({1,2}, {2,1}) = true, want false")
	}
	if !SamePoint(Point{}, Point{0, 0}) {
		t.Error("the zero Point should equal an explicit {0, 0}")
	}
}

func TestCountUnique(t *testing.T) {
	tests := []struct {
		in   []Point
		want int
	}{
		{[]Point{{1, 2}, {1, 2}, {3, 4}}, 2},
		{[]Point{{1, 2}, {2, 1}}, 2},
		{[]Point{{0, 0}, {}}, 1},
		{[]Point{{1, 1}}, 1},
		{nil, 0},
	}
	for _, tt := range tests {
		if got := CountUnique(tt.in); got != tt.want {
			t.Errorf("CountUnique(%v) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestEqualConfigs(t *testing.T) {
	base := Config{Name: "api", Where: Point{1, 2}, Hosts: []string{"a", "b"}}

	same := Config{Name: "api", Where: Point{1, 2}, Hosts: []string{"a", "b"}}
	if !EqualConfigs(base, same) {
		t.Error("two identically-built Configs compared unequal")
	}

	if EqualConfigs(base, Config{Name: "web", Where: Point{1, 2}, Hosts: []string{"a", "b"}}) {
		t.Error("a different Name compared equal")
	}
	if EqualConfigs(base, Config{Name: "api", Where: Point{9, 9}, Hosts: []string{"a", "b"}}) {
		t.Error("a different Where compared equal")
	}
	if EqualConfigs(base, Config{Name: "api", Where: Point{1, 2}, Hosts: []string{"b", "a"}}) {
		t.Error("Hosts in a different order compared equal - order counts")
	}
	if EqualConfigs(base, Config{Name: "api", Where: Point{1, 2}, Hosts: []string{"a"}}) {
		t.Error("a shorter Hosts compared equal")
	}

	// The point of writing it by hand: nil and empty both mean "no hosts".
	nilHosts := Config{Name: "api"}
	emptyHosts := Config{Name: "api", Hosts: []string{}}
	if !EqualConfigs(nilHosts, emptyHosts) {
		t.Error("EqualConfigs treated a nil Hosts as different from an empty one")
	}
}

func TestDeepEqualConfigs(t *testing.T) {
	base := Config{Name: "api", Where: Point{1, 2}, Hosts: []string{"a", "b"}}
	same := Config{Name: "api", Where: Point{1, 2}, Hosts: []string{"a", "b"}}
	if !DeepEqualConfigs(base, same) {
		t.Error("DeepEqual on two identically-built Configs = false, want true")
	}
	if DeepEqualConfigs(base, Config{Name: "api", Where: Point{1, 2}, Hosts: []string{"a"}}) {
		t.Error("DeepEqual on different Hosts = true, want false")
	}

	// Where the two comparisons part company.
	nilHosts := Config{Name: "api"}
	emptyHosts := Config{Name: "api", Hosts: []string{}}
	if DeepEqualConfigs(nilHosts, emptyHosts) {
		t.Error("DeepEqual said a nil slice equals an empty one - it does not, and that is the difference from EqualConfigs")
	}
}

func TestSameNode(t *testing.T) {
	tail := &Node{Label: "tail"}

	// The same pointer on both sides: equal.
	if !SameNode(Node{Label: "head", Next: tail}, Node{Label: "head", Next: tail}) {
		t.Error("two nodes sharing one Next pointer compared unequal")
	}

	// Different pointers holding equal values: NOT equal.
	otherTail := &Node{Label: "tail"}
	if SameNode(Node{Label: "head", Next: tail}, Node{Label: "head", Next: otherTail}) {
		t.Error("two nodes with distinct but identical Next pointers compared equal - == on a pointer compares addresses")
	}

	if !SameNode(Node{Label: "solo"}, Node{Label: "solo"}) {
		t.Error("two nodes with a nil Next and the same label compared unequal")
	}
	if SameNode(Node{Label: "a"}, Node{Label: "b"}) {
		t.Error("different labels compared equal")
	}
}

func TestCanCompare(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want bool
	}{
		{"a comparable struct", Point{1, 2}, true},
		{"a pointer", &Node{}, true},
		{"a string", "go", true},
		{"a struct with a slice field", Config{}, false},
		{"a slice", []string{}, false},
		{"a map", map[string]int{}, false},
		{"a func", func() {}, false},
		{"nil", nil, false},
		{"an array of a comparable type", [2]Point{}, true},
		{"an array of a non-comparable type", [2]Config{}, false},
	}
	for _, tt := range tests {
		if got := CanCompare(tt.in); got != tt.want {
			t.Errorf("CanCompare(%s) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestDedup(t *testing.T) {
	got := Dedup([]Point{{1, 2}, {3, 4}, {1, 2}, {3, 4}, {5, 6}})
	want := []Point{{1, 2}, {3, 4}, {5, 6}}
	if !slices.Equal(got, want) {
		t.Errorf("Dedup = %v, want %v - keep the first occurrence of each", got, want)
	}

	if got := Dedup([]Point{{1, 1}}); !slices.Equal(got, []Point{{1, 1}}) {
		t.Errorf("Dedup of one point = %v", got)
	}
	if got := Dedup(nil); got == nil || len(got) != 0 {
		t.Errorf("Dedup(nil) = %v, want an empty non-nil slice", got)
	}
}
