package main

import (
	"maps"
	"slices"
	"testing"
)

func TestCount(t *testing.T) {
	tests := []struct {
		name string
		text string
		want map[string]int
	}{
		{"empty", "", map[string]int{}},
		{"whitespace only", "   \n\t ", map[string]int{}},
		{"simple", "a b a", map[string]int{"a": 2, "b": 1}},
		{"case folded", "Go go GO", map[string]int{"go": 3}},
		{
			"punctuation trimmed",
			`the dog. "The dog," barks!`,
			map[string]int{"the": 2, "dog": 2, "barks": 1},
		},
		{"inner punctuation kept", "don't stop", map[string]int{"don't": 1, "stop": 1}},
		{"punctuation only token", "hi -- there", map[string]int{"hi": 1, "--": 1, "there": 1}},
		{"bare punctuation dropped", "hi ... there", map[string]int{"hi": 1, "there": 1}},
		{"newlines and tabs split", "one\ntwo\tthree", map[string]int{"one": 1, "two": 1, "three": 1}},
	}
	for _, tt := range tests {
		got := Count(tt.text)
		if got == nil {
			t.Errorf("Count(%s) = nil, want a map", tt.name)
			continue
		}
		if !maps.Equal(got, tt.want) {
			t.Errorf("Count(%s) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestUnique(t *testing.T) {
	if got := Unique("a b a c b a"); got != 3 {
		t.Errorf("Unique = %d, want 3", got)
	}
	if got := Unique("Go go GO!"); got != 1 {
		t.Errorf("Unique = %d, want 1", got)
	}
	if got := Unique(""); got != 0 {
		t.Errorf("Unique(empty) = %d, want 0", got)
	}
}

func TestSortedPairs(t *testing.T) {
	counts := map[string]int{"b": 2, "a": 2, "c": 5, "d": 1}
	want := []Pair{{"c", 5}, {"a", 2}, {"b", 2}, {"d", 1}}

	// Repeated, because map range order changes between runs and a tie broken
	// by luck rather than by the comparison would pass once.
	for range 10 {
		if got := SortedPairs(counts); !slices.Equal(got, want) {
			t.Fatalf("SortedPairs = %v, want %v", got, want)
		}
	}
	if got := SortedPairs(map[string]int{}); len(got) != 0 {
		t.Errorf("SortedPairs(empty) = %v, want an empty result", got)
	}
	if got := SortedPairs(nil); len(got) != 0 {
		t.Errorf("SortedPairs(nil) = %v, want an empty result", got)
	}
}

func TestSortedPairsDoesNotDisturbTheMap(t *testing.T) {
	counts := map[string]int{"a": 1, "b": 2}
	SortedPairs(counts)
	if want := (map[string]int{"a": 1, "b": 2}); !maps.Equal(counts, want) {
		t.Errorf("SortedPairs changed its argument to %v, want %v", counts, want)
	}
}

func TestTopN(t *testing.T) {
	counts := map[string]int{"b": 2, "a": 2, "c": 5, "d": 1}
	tests := []struct {
		n    int
		want []Pair
	}{
		{2, []Pair{{"c", 5}, {"a", 2}}},
		{0, nil},
		{-3, nil},
		{4, []Pair{{"c", 5}, {"a", 2}, {"b", 2}, {"d", 1}}},
		{99, []Pair{{"c", 5}, {"a", 2}, {"b", 2}, {"d", 1}}},
	}
	for _, tt := range tests {
		got := TopN(counts, tt.n)
		if len(got) != len(tt.want) || !slices.Equal(got, tt.want) {
			t.Errorf("TopN(counts, %d) = %v, want %v", tt.n, got, tt.want)
		}
	}
}

func TestReport(t *testing.T) {
	counts := Count("the dog. The dog barks the loudest")
	want := "the: 3\ndog: 2\nbarks: 1\n"
	if got := Report(counts, 3); got != want {
		t.Errorf("Report(counts, 3) = %q, want %q", got, want)
	}
	if got := Report(counts, 0); got != "" {
		t.Errorf("Report(counts, 0) = %q, want an empty string", got)
	}
	if got := Report(nil, 3); got != "" {
		t.Errorf("Report(nil, 3) = %q, want an empty string", got)
	}
}
