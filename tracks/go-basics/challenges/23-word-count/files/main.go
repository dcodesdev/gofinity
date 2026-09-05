package main

import "fmt"

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
	// TODO
	return nil
}

// Unique reports how many distinct words text contains.
func Unique(text string) int {
	// TODO
	return 0
}

// SortedPairs returns every entry of counts as a Pair, ordered by descending
// count and, for equal counts, by ascending word - so the answer never depends
// on map range order.
func SortedPairs(counts map[string]int) []Pair {
	// TODO
	return nil
}

// TopN returns the first n entries of SortedPairs. A negative n gives an empty
// result and an n larger than the map gives all of it.
func TopN(counts map[string]int, n int) []Pair {
	// TODO
	return nil
}

// Report renders the top n entries as one "word: count" line each, in order,
// with a trailing newline after every line.
func Report(counts map[string]int, n int) string {
	// TODO
	return ""
}

func main() {
	counts := Count("the quick brown fox jumps over the lazy dog. The dog barks!")
	fmt.Print(Report(counts, 3))
	fmt.Println(Unique("a a b"))
}
