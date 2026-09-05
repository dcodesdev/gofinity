package main

import (
	"strings"
	"testing"
)

func TestRepeat(t *testing.T) {
	tests := []struct {
		s    string
		n    int
		want string
	}{
		{"ab", 3, "ababab"},
		{"x", 1, "x"},
		{"ab", 0, ""},
		{"ab", -2, ""},
		{"", 5, ""},
	}
	for _, tt := range tests {
		if got := Repeat(tt.s, tt.n); got != tt.want {
			t.Errorf("Repeat(%q, %d) = %q, want %q", tt.s, tt.n, got, tt.want)
		}
	}
}

// The point of the exercise: a Builder that is grown once allocates once, while
// s += piece in a loop allocates on every iteration and copies everything it
// has already written.
func TestRepeatAllocatesOnce(t *testing.T) {
	const limit = 4
	got := testing.AllocsPerRun(10, func() {
		Repeat("abcd", 64)
	})
	if got > limit {
		t.Errorf("Repeat allocated %.0f times, want at most %d - use a strings.Builder and Grow it before the loop", got, limit)
	}
}

func TestJoinLines(t *testing.T) {
	tests := []struct {
		in   []string
		want string
	}{
		{[]string{"a", "b"}, "a\nb\n"},
		{[]string{"only"}, "only\n"},
		{[]string{""}, "\n"},
		{[]string{}, ""},
		{nil, ""},
	}
	for _, tt := range tests {
		if got := JoinLines(tt.in); got != tt.want {
			t.Errorf("JoinLines(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestAppendTo(t *testing.T) {
	var b strings.Builder
	AppendTo(&b, "a", "b")
	if got := b.String(); got != "a, b" {
		t.Fatalf("after one call = %q, want %q", got, "a, b")
	}

	// The second call has to see what the first one wrote and keep going.
	AppendTo(&b, "c")
	if got := b.String(); got != "a, b, c" {
		t.Fatalf("after two calls = %q, want %q", got, "a, b, c")
	}

	AppendTo(&b)
	if got := b.String(); got != "a, b, c" {
		t.Errorf("after an empty call = %q, want it unchanged at %q", got, "a, b, c")
	}
}

func TestAppendToEmptyBuilder(t *testing.T) {
	var b strings.Builder
	AppendTo(&b)
	if got := b.String(); got != "" {
		t.Errorf("AppendTo with no items wrote %q, want nothing", got)
	}

	var c strings.Builder
	c.WriteString("head")
	AppendTo(&c, "tail")
	if got := c.String(); got != "head, tail" {
		t.Errorf("AppendTo onto a pre-filled builder = %q, want %q", got, "head, tail")
	}
}

func TestQuoteList(t *testing.T) {
	tests := []struct {
		in   []string
		want string
	}{
		{[]string{"a", "b"}, `["a", "b"]`},
		{[]string{"one"}, `["one"]`},
		{[]string{}, "[]"},
		{nil, "[]"},
		{[]string{`he said "hi"`}, `["he said \"hi\""]`},
		{[]string{"line\n"}, `["line\n"]`},
	}
	for _, tt := range tests {
		if got := QuoteList(tt.in); got != tt.want {
			t.Errorf("QuoteList(%q) = %s, want %s", tt.in, got, tt.want)
		}
	}
}

func TestHexDump(t *testing.T) {
	tests := []struct {
		in   []byte
		want string
	}{
		{[]byte("Go"), "47 6f"},
		{[]byte{0, 255, 15}, "00 ff 0f"},
		{[]byte{7}, "07"},
		{[]byte{}, ""},
		{nil, ""},
	}
	for _, tt := range tests {
		if got := HexDump(tt.in); got != tt.want {
			t.Errorf("HexDump(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestTable(t *testing.T) {
	pad := func(n int) string { return strings.Repeat(" ", n) }

	got := Table([]string{"go", "rust"}, []int{3, 2})
	want := "go" + pad(8) + pad(4) + "3\n" + "rust" + pad(6) + pad(4) + "2\n"
	if got != want {
		t.Errorf("Table = %q, want %q", got, want)
	}

	// Exactly ten characters gets no padding, and eleven is not truncated.
	got = Table([]string{"javascript", "typescript!"}, []int{5, 123456})
	want = "javascript" + pad(4) + "5\n" + "typescript!" + "123456\n"
	if got != want {
		t.Errorf("Table(wide) = %q, want %q", got, want)
	}

	if got := Table([]string{"a", "b"}, []int{1}); got != "a"+pad(9)+pad(4)+"1\n" {
		t.Errorf("Table with a short counts slice = %q", got)
	}
	if got := Table(nil, nil); got != "" {
		t.Errorf("Table(nil, nil) = %q, want an empty string", got)
	}
}
