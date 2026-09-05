package main

import (
	"fmt"
	"strings"
)

// Repeat returns s written out n times. A zero or negative n gives "".
// Build the result with a strings.Builder and reserve the space up front.
func Repeat(s string, n int) string {
	// TODO
	return ""
}

// JoinLines writes every line followed by "\n". A nil or empty slice gives "".
func JoinLines(lines []string) string {
	// TODO
	return ""
}

// AppendTo writes items into b, separated by ", ". It joins onto whatever is
// already in b, so a second call continues the same list rather than starting a
// new one, and an empty items adds nothing at all.
func AppendTo(b *strings.Builder, items ...string) {
	// TODO
}

// QuoteList renders items as a Go-style list of quoted strings:
// ["a", "b"]. An empty or nil slice gives [].
func QuoteList(items []string) string {
	// TODO
	return ""
}

// HexDump renders every byte of data as two lowercase hex digits, separated by
// a single space. Empty data gives "".
func HexDump(data []byte) string {
	// TODO
	return ""
}

// Table renders one "name count" row per line: the name left-aligned in a
// 10-column field, then the count right-aligned in a 5-column field, then a
// newline. It stops at the shorter of the two slices.
func Table(names []string, counts []int) string {
	// TODO
	return ""
}

func main() {
	fmt.Print(Table([]string{"go", "rust"}, []int{3, 2}))
	fmt.Println(QuoteList([]string{"a", "b"}))
	fmt.Println(HexDump([]byte("Go")))
}
