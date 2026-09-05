package main

import (
	"fmt"
	"strings"
)

// Normalize collapses every run of whitespace in s to a single space, trims the
// ends, and lowercases the result. "  Hello   WORLD \n" becomes "hello world".
func Normalize(s string) string {
	// Fields already splits on runs of whitespace and drops the empties, so
	// splitting and rejoining does the trimming and the collapsing at once.
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}

// CountWords reports how many whitespace-separated words s contains.
func CountWords(s string) int {
	return len(strings.Fields(s))
}

// TitleWords uppercases the first letter of every word in s and leaves the rest
// of each word untouched, keeping the words separated by a single space.
// "hello wide world" becomes "Hello Wide World".
func TitleWords(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		// []rune, not word[:1], so a multi-byte first letter survives.
		r := []rune(word)
		words[i] = strings.ToUpper(string(r[0])) + string(r[1:])
	}
	return strings.Join(words, " ")
}

// Initials returns the first letter of each word in name, uppercased and joined
// with nothing between them. "ada lovelace" gives "AL".
func Initials(name string) string {
	var b strings.Builder
	for _, word := range strings.Fields(name) {
		r := []rune(word)
		b.WriteString(strings.ToUpper(string(r[0])))
	}
	return b.String()
}

// Redact replaces every occurrence of secret in s with one '*' per byte of
// secret. An empty secret redacts nothing and s comes back unchanged.
func Redact(s, secret string) string {
	if secret == "" {
		// ReplaceAll with an empty old string inserts the replacement between
		// every rune, which is the opposite of leaving the string alone.
		return s
	}
	return strings.ReplaceAll(s, secret, strings.Repeat("*", len(secret)))
}

// SplitTrim splits s on sep, trims the whitespace around every field, and drops
// the fields that are then empty. It always returns a non-nil slice.
func SplitTrim(s, sep string) []string {
	fields := make([]string, 0)
	for _, part := range strings.Split(s, sep) {
		if part = strings.TrimSpace(part); part != "" {
			fields = append(fields, part)
		}
	}
	return fields
}

func main() {
	fmt.Println(Normalize("  Go   IS   fun  "))
	fmt.Println(TitleWords("hello wide world"))
	fmt.Println(Initials("ada lovelace"))
	fmt.Println(Redact("token=abc123 ok", "abc123"))
	fmt.Println(SplitTrim("a, b ,, c", ","))
}
