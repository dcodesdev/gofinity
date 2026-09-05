// Package textkit turns free-form text into the shapes a page needs: a
// headline, a URL slug, a word count.
//
// The package comment sits directly above the package clause with no blank
// line between them. That is what `go doc gofinity/splitpackage/textkit`
// prints first.
package textkit

import (
	"strings"
	"unicode"
)

// Title returns s with the first letter of every space-separated word in upper
// case and the rest in lower case.
func Title(s string) string {
	fields := strings.Fields(s)
	for i, field := range fields {
		runes := []rune(strings.ToLower(field))
		runes[0] = unicode.ToUpper(runes[0])
		fields[i] = string(runes)
	}
	return strings.Join(fields, " ")
}

// Slug returns s as a URL-safe slug.
func Slug(s string) string {
	var b strings.Builder
	pendingSeparator := false
	for _, r := range strings.ToLower(s) {
		if isWordChar(r) {
			if pendingSeparator && b.Len() > 0 {
				b.WriteByte('-')
			}
			pendingSeparator = false
			b.WriteRune(r)
			continue
		}
		pendingSeparator = true
	}
	return b.String()
}

// WordCount returns how many space-separated words s holds.
func WordCount(s string) int {
	return len(strings.Fields(s))
}

// isWordChar reports whether r may appear in a slug. Its name is lower case, so
// it is unexported: nothing outside this package can call it, which is exactly
// what a helper wants.
func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}
