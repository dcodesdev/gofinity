package main

import (
	"fmt"
	"strings"
)

// Repeat returns s written out n times. A zero or negative n gives "".
// Build the result with a strings.Builder and reserve the space up front.
func Repeat(s string, n int) string {
	if n <= 0 {
		return ""
	}
	var b strings.Builder
	// One allocation instead of one per iteration: the final size is known.
	b.Grow(len(s) * n)
	for range n {
		b.WriteString(s)
	}
	return b.String()
}

// JoinLines writes every line followed by "\n". A nil or empty slice gives "".
func JoinLines(lines []string) string {
	var b strings.Builder
	for _, line := range lines {
		b.WriteString(line)
		// WriteByte for a single ASCII byte; WriteRune would work too and does
		// more.
		b.WriteByte('\n')
	}
	return b.String()
}

// AppendTo writes items into b, separated by ", ". It joins onto whatever is
// already in b, so a second call continues the same list rather than starting a
// new one, and an empty items adds nothing at all.
func AppendTo(b *strings.Builder, items ...string) {
	for _, item := range items {
		// Len is what already-written looks like from the inside, and it is
		// why this works across calls.
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		b.WriteString(item)
	}
}

// QuoteList renders items as a Go-style list of quoted strings:
// ["a", "b"]. An empty or nil slice gives [].
func QuoteList(items []string) string {
	var b strings.Builder
	b.WriteByte('[')
	for i, item := range items {
		if i > 0 {
			b.WriteString(", ")
		}
		// %q quotes and escapes, so a newline or a quote inside the item comes
		// out as source-legal Go rather than breaking the line.
		fmt.Fprintf(&b, "%q", item)
	}
	b.WriteByte(']')
	return b.String()
}

// HexDump renders every byte of data as two lowercase hex digits, separated by
// a single space. Empty data gives "".
func HexDump(data []byte) string {
	var b strings.Builder
	b.Grow(3 * len(data))
	for i, c := range data {
		if i > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%02x", c)
	}
	return b.String()
}

// Table renders one "name count" row per line: the name left-aligned in a
// 10-column field, then the count right-aligned in a 5-column field, then a
// newline. It stops at the shorter of the two slices.
func Table(names []string, counts []int) string {
	rows := min(len(names), len(counts))
	var b strings.Builder
	for i := range rows {
		fmt.Fprintf(&b, "%-10s%5d\n", names[i], counts[i])
	}
	return b.String()
}

func main() {
	fmt.Print(Table([]string{"go", "rust"}, []int{3, 2}))
	fmt.Println(QuoteList([]string{"a", "b"}))
	fmt.Println(HexDump([]byte("Go")))
}
