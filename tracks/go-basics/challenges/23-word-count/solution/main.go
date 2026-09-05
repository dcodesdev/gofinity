package main

import (
	"fmt"
	"sort"
	"strings"
)

// Pair is one word and how often it appeared. Do not change this type - the
// tests construct it.
type Pair struct {
	Word  string
	Count int
}

// Punctuation is trimmed from both ends of every word before it is counted.
const Punctuation = `.,!?;:"'()`

// Count tallies the words in text. Words are separated by any whitespace,
// lowercased, and stripped of leading and trailing punctuation; a token that is
// nothing but punctuation is not a word and is skipped.
func Count(text string) map[string]int {
	counts := make(map[string]int)
	for _, field := range strings.Fields(text) {
		// Trim first, then fold: trimming cannot change the case and folding
		// cannot change the punctuation, so the order is free either way.
		word := strings.Trim(strings.ToLower(field), Punctuation)
		if word == "" {
			continue
		}
		counts[word]++
	}
	return counts
}

// Unique reports how many distinct words text contains.
func Unique(text string) int {
	// One key per distinct word is exactly what a map already gives us.
	return len(Count(text))
}

// SortedPairs returns every entry of counts as a Pair, ordered by descending
// count and, for equal counts, by ascending word - so the answer never depends
// on map range order.
func SortedPairs(counts map[string]int) []Pair {
	pairs := make([]Pair, 0, len(counts))
	for word, n := range counts {
		pairs = append(pairs, Pair{Word: word, Count: n})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Count != pairs[j].Count {
			return pairs[i].Count > pairs[j].Count
		}
		// Without this tie-break the order of equal counts would still be the
		// random range order, and the test would fail one run in a few.
		return pairs[i].Word < pairs[j].Word
	})
	return pairs
}

// TopN returns the first n entries of SortedPairs. A negative n gives an empty
// result and an n larger than the map gives all of it.
func TopN(counts map[string]int, n int) []Pair {
	pairs := SortedPairs(counts)
	if n < 0 {
		n = 0
	}
	if n > len(pairs) {
		n = len(pairs)
	}
	return pairs[:n]
}

// Report renders the top n entries as one "word: count" line each, in order,
// with a trailing newline after every line.
func Report(counts map[string]int, n int) string {
	var b strings.Builder
	for _, p := range TopN(counts, n) {
		fmt.Fprintf(&b, "%s: %d\n", p.Word, p.Count)
	}
	return b.String()
}

func main() {
	counts := Count("the quick brown fox jumps over the lazy dog. The dog barks!")
	fmt.Print(Report(counts, 3))
	fmt.Println(Unique("a a b"))
}
