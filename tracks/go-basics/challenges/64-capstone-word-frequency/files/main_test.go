package main

import (
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

// refTokenize is the reference the tokenizer is graded against.
func refTokenize(text string) []string {
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

func refTally(words []string, stop map[string]bool, minLen int) map[string]int {
	counts := make(map[string]int)
	for _, w := range words {
		if stop[w] || utf8.RuneCountInString(w) < minLen {
			continue
		}
		counts[w]++
	}
	return counts
}

func refRank(counts map[string]int, top int) []Count {
	ranked := make([]Count, 0, len(counts))
	for w, n := range counts {
		ranked = append(ranked, Count{Word: w, N: n})
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

func sameWords(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func TestTokenizeCases(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"words and punctuation", "Go, go -- stop!", []string{"go", "go", "stop"}},
		{"lowercased", "GO Go gO", []string{"go", "go", "go"}},
		{"apostrophe inside a word", "don't stop", []string{"don't", "stop"}},
		{"quotes are not part of the word", "'quoted'", []string{"quoted"}},
		{"trailing apostrophe drops", "dogs' bowls", []string{"dogs", "bowls"}},
		{"digits count", "go1 24 v2", []string{"go1", "24", "v2"}},
		{"newlines separate", "one\ntwo\r\nthree", []string{"one", "two", "three"}},
		{"hyphen separates", "well-known", []string{"well", "known"}},
		{"empty text", "", nil},
		{"punctuation only", " ... !? ", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Tokenize(c.in)
			if !sameWords(got, c.want) {
				t.Errorf("Tokenize(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestTokenizeUnicode(t *testing.T) {
	got := Tokenize("Café NAÏVE Ångström")
	want := []string{"café", "naïve", "ångström"}
	if !sameWords(got, want) {
		t.Errorf("Tokenize(...) = %q, want %q - decide letters with the unicode package and lowercase runes, not bytes", got, want)
	}
}

func TestTokenizeAgainstReference(t *testing.T) {
	corpus := []string{
		"The quick brown fox jumps over the lazy dog.",
		"Rob said: \"it's Go, not GOLANG\" -- twice!",
		"tabs\tand   spaces\n\nand 42 numbers, 3.14 too",
		"---",
		"a'b'c d''e 'f",
		"Ünïcödé, straße, 東京 tokyo",
	}
	for _, text := range corpus {
		got := Tokenize(text)
		want := refTokenize(text)
		if !sameWords(got, want) {
			t.Errorf("Tokenize(%q)\n got %q\nwant %q", text, got, want)
		}
	}
}

func TestTallyFilters(t *testing.T) {
	words := []string{"go", "go", "the", "gopher", "the", "run"}
	stop := map[string]bool{"the": true}
	got := Tally(words, stop, 3)
	want := map[string]int{"gopher": 1, "run": 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Tally(...) = %v, want %v", got, want)
	}
}

func TestTallyLengthIsInRunes(t *testing.T) {
	got := Tally([]string{"café", "naïve", "go"}, nil, 5)
	want := map[string]int{"naïve": 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Tally([café naïve go], nil, 5) = %v, want %v - minLen counts runes; café is 4 characters and 5 bytes", got, want)
	}
}

func TestTallyNilStopAndEmptyResult(t *testing.T) {
	if got := Tally([]string{"a", "b", "a"}, nil, 0); !reflect.DeepEqual(got, map[string]int{"a": 2, "b": 1}) {
		t.Errorf("a nil stop map must skip nothing, got %v", got)
	}
	empty := Tally(nil, nil, 0)
	if empty == nil {
		t.Fatalf("Tally returned a nil map; it must return an empty one so callers can read from it")
	}
	if len(empty) != 0 {
		t.Errorf("Tally(nil, nil, 0) = %v, want an empty map", empty)
	}
}

func TestRankOrdersByCountThenWord(t *testing.T) {
	counts := map[string]int{"go": 5, "run": 5, "stop": 9, "wait": 1}
	got := Rank(counts, 0)
	want := []Count{{"stop", 9}, {"go", 5}, {"run", 5}, {"wait", 1}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Rank(...) = %v, want %v - most frequent first, alphabetical among ties", got, want)
	}
}

func TestRankTop(t *testing.T) {
	counts := map[string]int{"a": 4, "b": 3, "c": 2, "d": 1}
	if got := Rank(counts, 2); !reflect.DeepEqual(got, []Count{{"a", 4}, {"b", 3}}) {
		t.Errorf("Rank(..., 2) = %v, want the first two", got)
	}
	if got := Rank(counts, 99); len(got) != 4 {
		t.Errorf("Rank(..., 99) returned %d entries, want all 4 - top larger than the ranking is not an error", len(got))
	}
	if got := Rank(counts, -1); len(got) != 4 {
		t.Errorf("Rank(..., -1) returned %d entries, want all 4 - zero or negative means all", len(got))
	}
}

func TestRankEmpty(t *testing.T) {
	got := Rank(map[string]int{}, 3)
	if got == nil {
		t.Fatalf("Rank returned nil; it must return an empty slice")
	}
	if len(got) != 0 {
		t.Errorf("Rank(empty) = %v, want no entries", got)
	}
}

func TestRankAgainstReference(t *testing.T) {
	counts := refTally(refTokenize(sampleText), map[string]bool{"the": true}, 3)
	for _, top := range []int{0, 1, 3, 7, 100} {
		got, want := Rank(counts, top), refRank(counts, top)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Rank(counts, %d) = %v, want %v", top, got, want)
		}
	}
}

func TestReportFormat(t *testing.T) {
	got := Report([]Count{{"go", 12}, {"internationalization", 7}, {"run", 1}}, 50)
	want := " 1. go             12  24.0%\n 2. internationalization    7  14.0%\n 3. run             1   2.0%\n"
	if got != want {
		t.Errorf("Report(...) =\n%q\nwant\n%q", got, want)
	}
}

func TestReportEmptyAndZeroTotal(t *testing.T) {
	if got := Report(nil, 10); got != "" {
		t.Errorf("Report(nil, 10) = %q, want the empty string", got)
	}
	if got := Report([]Count{}, 10); got != "" {
		t.Errorf("Report(empty, 10) = %q, want the empty string", got)
	}
	got := Report([]Count{{"go", 3}}, 0)
	want := " 1. go              3   0.0%\n"
	if got != want {
		t.Errorf("Report(..., 0) = %q, want %q - a zero total is a zero share, not a division", got, want)
	}
}

func TestReportEveryLineEndsInNewline(t *testing.T) {
	got := Report([]Count{{"a", 1}, {"b", 1}}, 2)
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("Report(...) = %q, want the last line terminated too", got)
	}
	if n := strings.Count(got, "\n"); n != 2 {
		t.Errorf("Report(...) has %d newlines, want one per entry", n)
	}
}

const sampleText = `Go is expressive, concise, clean and efficient. Its concurrency
mechanisms make it easy to write programs that get the most out of
multicore machines, while its novel type system enables flexible and
modular program construction. Go compiles quickly to machine code, yet
it has the convenience of garbage collection and the power of run-time
reflection. It is a fast, statically typed, compiled language that feels
like a dynamically typed, interpreted language.`

func TestAnalyze(t *testing.T) {
	stop := map[string]bool{"the": true, "and": true, "that": true, "its": true, "it": true, "of": true, "to": true}
	got, err := Analyze(sampleText, stop, 4, 5)
	if err != nil {
		t.Fatalf("Analyze(...) returned %v, want a report", err)
	}
	want := " 1. language        2   4.7%\n 2. typed           2   4.7%\n 3. clean           1   2.3%\n 4. code            1   2.3%\n 5. collection      1   2.3%\n"
	if got != want {
		t.Errorf("Analyze(...) =\n%s\nwant\n%s", got, want)
	}
}

func TestAnalyzeSharesUseTheWholeCount(t *testing.T) {
	// Eight counted words, four of them "goes": goes is 50%, however small top is.
	text := "goes goes goes goes alpha bravo charlie delta"
	got, err := Analyze(text, nil, 3, 1)
	if err != nil {
		t.Fatalf("Analyze(...) returned %v", err)
	}
	if !strings.Contains(got, " 50.0%") {
		t.Errorf("Analyze(...) = %q, want goes at 50.0%% - the share is of every counted word, not of the entries that survived top", got)
	}
}

func TestAnalyzeNoWords(t *testing.T) {
	cases := []struct {
		name   string
		text   string
		stop   map[string]bool
		minLen int
	}{
		{"nothing at all", "  ... !! ", nil, 0},
		{"every word is a stop word", "the and the", map[string]bool{"the": true, "and": true}, 0},
		{"every word is too short", "a bc de", nil, 4},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			report, err := Analyze(c.text, c.stop, c.minLen, 5)
			if err == nil {
				t.Fatalf("Analyze(%q, ...) returned %q and no error, want ErrNoWords", c.text, report)
			}
			if !errors.Is(err, ErrNoWords) {
				t.Fatalf("Analyze(%q, ...) returned %v, want an error matching ErrNoWords", c.text, err)
			}
			if report != "" {
				t.Errorf("Analyze returned both a report %q and an error", report)
			}
			if err.Error() == ErrNoWords.Error() {
				t.Errorf("Analyze returned ErrNoWords unchanged; wrap it with context, as in fmt.Errorf(\"analyze: %%w\", ErrNoWords)")
			}
		})
	}
}
