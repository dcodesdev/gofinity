package main

import (
	"slices"
	"testing"
)

func TestAppendAll(t *testing.T) {
	tests := []struct {
		name   string
		dst    []int
		values []int
		want   []int
	}{
		{"onto nil", nil, []int{1, 2, 3}, []int{1, 2, 3}},
		{"onto a slice", []int{1}, []int{2, 3}, []int{1, 2, 3}},
		{"nothing to append", []int{1, 2}, nil, []int{1, 2}},
		{"nothing onto nothing", nil, nil, nil},
	}
	for _, tt := range tests {
		if got := AppendAll(tt.dst, tt.values...); !slices.Equal(got, tt.want) {
			t.Errorf("AppendAll(%s) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAppendAllReturnsTheGrownSlice(t *testing.T) {
	// The backing array here is full, so append has to allocate a new one. A
	// solution that appends and throws the result away returns the old slice.
	dst := []int{1, 2, 3}
	got := AppendAll(dst, 4)
	if want := []int{1, 2, 3, 4}; !slices.Equal(got, want) {
		t.Fatalf("AppendAll([1 2 3], 4) = %v, want %v", got, want)
	}
	if len(dst) != 3 {
		t.Errorf("the argument became %v; append does not resize the caller's slice", dst)
	}
}

func TestCloneInts(t *testing.T) {
	s := []int{1, 2, 3}
	got := CloneInts(s)
	if !slices.Equal(got, s) {
		t.Fatalf("CloneInts(%v) = %v, want the same elements", s, got)
	}
	got[0] = 99
	if s[0] != 1 {
		t.Errorf("writing to the clone changed the original to %v; the two must not share an array", s)
	}
}

func TestCloneIntsEdges(t *testing.T) {
	if got := CloneInts(nil); got != nil {
		t.Errorf("CloneInts(nil) = %v, want nil", got)
	}
	got := CloneInts([]int{})
	if got == nil || len(got) != 0 {
		t.Errorf("CloneInts(empty) = %v, want an empty non-nil slice", got)
	}
}

func TestCloneIntsGrowsIndependently(t *testing.T) {
	s := make([]int, 2, 8)
	s[0], s[1] = 1, 2
	clone := CloneInts(s)
	clone = append(clone, 3)
	if len(s) != 2 || s[0] != 1 || s[1] != 2 {
		t.Errorf("appending to the clone disturbed the original: %v", s)
	}
	if want := []int{1, 2, 3}; !slices.Equal(clone, want) {
		t.Errorf("clone = %v, want %v", clone, want)
	}
}

func TestCopyInto(t *testing.T) {
	tests := []struct {
		name  string
		dst   []int
		src   []int
		wantN int
		want  []int
	}{
		{"same length", make([]int, 3), []int{1, 2, 3}, 3, []int{1, 2, 3}},
		{"dst is shorter", make([]int, 2), []int{1, 2, 3}, 2, []int{1, 2}},
		{"src is shorter", []int{9, 9, 9}, []int{1}, 1, []int{1, 9, 9}},
		{"dst is empty", []int{}, []int{1, 2}, 0, []int{}},
		{"src is nil", []int{9, 9}, nil, 0, []int{9, 9}},
	}
	for _, tt := range tests {
		n := CopyInto(tt.dst, tt.src)
		if n != tt.wantN {
			t.Errorf("CopyInto(%s) returned %d, want %d", tt.name, n, tt.wantN)
		}
		if !slices.Equal(tt.dst, tt.want) {
			t.Errorf("CopyInto(%s) left dst = %v, want %v", tt.name, tt.dst, tt.want)
		}
	}
}

func TestCopyIntoDoesNotGrowDst(t *testing.T) {
	dst := make([]int, 0, 10)
	if n := CopyInto(dst, []int{1, 2, 3}); n != 0 {
		t.Errorf("CopyInto(len 0, cap 10) copied %d, want 0: copy stops at len, not cap", n)
	}
}

func TestInsert(t *testing.T) {
	tests := []struct {
		name string
		s    []int
		i    int
		v    int
		want []int
	}{
		{"in the middle", []int{1, 2, 4}, 2, 3, []int{1, 2, 3, 4}},
		{"at the front", []int{2, 3}, 0, 1, []int{1, 2, 3}},
		{"at the end", []int{1, 2}, 2, 3, []int{1, 2, 3}},
		{"past the end", []int{1, 2}, 9, 3, []int{1, 2, 3}},
		{"before the front", []int{2, 3}, -4, 1, []int{1, 2, 3}},
		{"into nil", nil, 0, 1, []int{1}},
	}
	for _, tt := range tests {
		if got := Insert(tt.s, tt.i, tt.v); !slices.Equal(got, tt.want) {
			t.Errorf("Insert(%s) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestInsertLeavesTheOriginalAlone(t *testing.T) {
	s := make([]int, 3, 16)
	copy(s, []int{1, 2, 4})
	Insert(s, 2, 3)
	if want := []int{1, 2, 4}; !slices.Equal(s, want) {
		t.Errorf("Insert changed its argument to %v, want %v", s, want)
	}
}

func TestRemoveAt(t *testing.T) {
	tests := []struct {
		name string
		s    []int
		i    int
		want []int
	}{
		{"from the middle", []int{1, 2, 3}, 1, []int{1, 3}},
		{"the first", []int{1, 2, 3}, 0, []int{2, 3}},
		{"the last", []int{1, 2, 3}, 2, []int{1, 2}},
		{"the only one", []int{1}, 0, []int{}},
		{"past the end", []int{1, 2}, 5, []int{1, 2}},
		{"negative", []int{1, 2}, -1, []int{1, 2}},
	}
	for _, tt := range tests {
		if got := RemoveAt(tt.s, tt.i); !slices.Equal(got, tt.want) {
			t.Errorf("RemoveAt(%s) = %v, want %v", tt.name, got, tt.want)
		}
	}
	if got := RemoveAt(nil, 0); len(got) != 0 {
		t.Errorf("RemoveAt(nil, 0) = %v, want an empty result", got)
	}
}

func TestRemoveAtLeavesTheOriginalAlone(t *testing.T) {
	s := []int{1, 2, 3}
	RemoveAt(s, 1)
	if want := []int{1, 2, 3}; !slices.Equal(s, want) {
		t.Errorf("RemoveAt changed its argument to %v, want %v", s, want)
	}
}

func TestConcat(t *testing.T) {
	tests := []struct {
		name string
		a    []int
		b    []int
		want []int
	}{
		{"both", []int{1, 2}, []int{3, 4}, []int{1, 2, 3, 4}},
		{"empty first", nil, []int{3, 4}, []int{3, 4}},
		{"empty second", []int{1, 2}, nil, []int{1, 2}},
		{"both empty", nil, nil, []int{}},
	}
	for _, tt := range tests {
		if got := Concat(tt.a, tt.b); !slices.Equal(got, tt.want) {
			t.Errorf("Concat(%s) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestConcatDoesNotShareWithA(t *testing.T) {
	// a has room to spare, so append(a, b...) would write b straight into a's
	// backing array and hand back a view of it. That is the bug this catches.
	a := make([]int, 2, 16)
	a[0], a[1] = 1, 2
	b := []int{3, 4}

	got := Concat(a, b)
	if want := []int{1, 2, 3, 4}; !slices.Equal(got, want) {
		t.Fatalf("Concat = %v, want %v", got, want)
	}

	got[0], got[2] = 90, 92
	if want := []int{1, 2}; !slices.Equal(a, want) {
		t.Errorf("writing to the result changed a to %v, want %v", a, want)
	}
	if want := []int{3, 4}; !slices.Equal(b, want) {
		t.Errorf("writing to the result changed b to %v, want %v", b, want)
	}
}
