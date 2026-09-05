package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ErrNoWords is what Analyze reports when the text contains nothing to count:
// no words at all, or none that survive the stop list and the length filter.
//
// It is a sentinel: callers compare with errors.Is, so Analyze must wrap it
// rather than return a freshly formatted error that only reads the same.
var ErrNoWords = errors.New("no countable words")

// Count is one word and how many times it was seen.
type Count struct {
	Word string
	N    int
}

// Tokenize splits text into lowercase words.
//
// A word is a maximal run of letters and digits. An apostrophe (') is kept
// only when it sits directly between two of those, so "don't" stays one word
// while 'quoted' loses both quotes. Everything else - spaces, punctuation,
// newlines, em dashes, hyphens - separates.
//
// Letters and digits are decided with the unicode package, not with a range
// check on a byte, so "Café" and "naïve" are one word each and lowercase
// correctly.
//
//	Tokenize("Go, go -- don't stop!")  ->  ["go", "go", "don't", "stop"]
//
// Text with no words gives an empty slice.
func Tokenize(text string) []string {
	runes := []rune(text)
	words := []string{}
	var cur []rune
	flush := func() {
		if len(cur) > 0 {
			words = append(words, string(cur))
			cur = nil
		}
	}
	for i, r := range runes {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			cur = append(cur, unicode.ToLower(r))
		case r == '\'' && len(cur) > 0 && i+1 < len(runes) && (unicode.IsLetter(runes[i+1]) || unicode.IsDigit(runes[i+1])):
			cur = append(cur, r)
		default:
			flush()
		}
	}
	flush()
	return words
}

// Tally counts words, skipping the ones the caller does not want.
//
// A word is skipped when stop[word] is true, or when it is shorter than
// minLen. Length is counted in *runes*, not bytes: "café" is four characters
// long even though it takes five bytes, and a byte count would quietly filter
// out different words depending on the language.
//
// A nil stop map is fine and skips nothing. The returned map is never nil.
func Tally(words []string, stop map[string]bool, minLen int) map[string]int {
	counts := make(map[string]int)
	for _, w := range words {
		if stop[w] {
			continue
		}
		if utf8.RuneCountInString(w) < minLen {
			continue
		}
		counts[w]++
	}
	return counts
}

// Rank turns a tally into a ranking: most frequent first, and alphabetical
// among words that tie, because a report that reorders itself between runs
// cannot be diffed or trusted.
//
// When top is greater than zero the result is cut to that many entries. Zero
// or negative means "all of them". The returned slice is never nil.
func Rank(counts map[string]int, top int) []Count {
	ranked := make([]Count, 0, len(counts))
	for word, n := range counts {
		ranked = append(ranked, Count{Word: word, N: n})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].N != ranked[j].N {
			return ranked[i].N > ranked[j].N
		}
		return ranked[i].Word < ranked[j].Word
	})
	if top > 0 && top < len(ranked) {
		ranked = ranked[:top]
	}
	return ranked
}

// Report renders a ranking as the program's output format. Each entry is one
// line, numbered from 1, and every line ends in a newline - including the
// last, so the output can be appended to:
//
//	" 1. go              12  24.0%\n"
//
// The exact line is:
//
//	fmt.Sprintf("%2d. %-12s %4d %5.1f%%\n", rank, c.Word, c.N, share)
//
// where share is the entry's percentage of total: 100 * N / total, as a
// float. When total is zero or negative every share is 0.
//
// An empty ranking renders as the empty string, not as a header or a blank
// line.
func Report(counts []Count, total int) string {
	var b strings.Builder
	for i, c := range counts {
		share := 0.0
		if total > 0 {
			share = 100 * float64(c.N) / float64(total)
		}
		fmt.Fprintf(&b, "%2d. %-12s %4d %5.1f%%\n", i+1, c.Word, c.N, share)
	}
	return b.String()
}

// Analyze is the whole program in one call: tokenize, tally, rank, report.
//
// total for the shares is the number of words that were actually counted -
// what is left after the stop list and minLen, before top cuts the ranking -
// so the shares describe the text rather than the part of it that fitted on
// the page.
//
// When nothing survives to count, Analyze returns an error wrapping ErrNoWords
// with context of its own, in the usual form:
//
//	fmt.Errorf("analyze: %w", ErrNoWords)
func Analyze(text string, stop map[string]bool, minLen, top int) (string, error) {
	counts := Tally(Tokenize(text), stop, minLen)
	if len(counts) == 0 {
		return "", fmt.Errorf("analyze: %w", ErrNoWords)
	}
	total := 0
	for _, n := range counts {
		total += n
	}
	return Report(Rank(counts, top), total), nil
}

func main() {
	const text = `Go is expressive, concise, clean, and efficient.
Its concurrency mechanisms make it easy to write programs that get the most
out of multicore and networked machines, while its novel type system enables
flexible and modular program construction.`

	stop := map[string]bool{"and": true, "it": true, "its": true, "the": true, "that": true, "to": true, "of": true}

	report, err := Analyze(text, stop, 3, 5)
	if err != nil {
		fmt.Println("no report:", err)
		return
	}
	fmt.Print(report)
}
