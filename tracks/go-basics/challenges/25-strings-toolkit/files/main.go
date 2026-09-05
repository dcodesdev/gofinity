package main

import "fmt"

// Normalize collapses every run of whitespace in s to a single space, trims the
// ends, and lowercases the result. "  Hello   WORLD \n" becomes "hello world".
func Normalize(s string) string {
	// TODO
	return ""
}

// CountWords reports how many whitespace-separated words s contains.
func CountWords(s string) int {
	// TODO
	return 0
}

// TitleWords uppercases the first letter of every word in s and leaves the rest
// of each word untouched, keeping the words separated by a single space.
// "hello wide world" becomes "Hello Wide World".
func TitleWords(s string) string {
	// TODO
	return ""
}

// Initials returns the first letter of each word in name, uppercased and joined
// with nothing between them. "ada lovelace" gives "AL".
func Initials(name string) string {
	// TODO
	return ""
}

// Redact replaces every occurrence of secret in s with one '*' per byte of
// secret. An empty secret redacts nothing and s comes back unchanged.
func Redact(s, secret string) string {
	// TODO
	return ""
}

// SplitTrim splits s on sep, trims the whitespace around every field, and drops
// the fields that are then empty. It always returns a non-nil slice.
func SplitTrim(s, sep string) []string {
	// TODO
	return nil
}

func main() {
	fmt.Println(Normalize("  Go   IS   fun  "))
	fmt.Println(TitleWords("hello wide world"))
	fmt.Println(Initials("ada lovelace"))
	fmt.Println(Redact("token=abc123 ok", "abc123"))
	fmt.Println(SplitTrim("a, b ,, c", ","))
}
