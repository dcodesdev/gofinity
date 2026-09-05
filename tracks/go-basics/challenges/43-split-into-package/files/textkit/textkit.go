// Package textkit turns free-form text into the shapes a page needs: a
// headline, a URL slug, a word count.
//
// The package comment sits directly above the package clause with no blank
// line between them. That is what `go doc gofinity/splitpackage/textkit`
// prints first.
package textkit

// Title returns s with the first letter of every space-separated word in upper
// case and the rest in lower case. Runs of spaces collapse to one, and leading
// and trailing spaces go away.
//
//	Title("  hello   go world ") == "Hello Go World"
//
// TODO: strings.Fields splits on any run of whitespace. Lower-case each field,
// upper-case its first rune, and join the result with a single space.
func Title(s string) string {
	return s
}

// Slug returns s as a URL-safe slug: lower case, with every run of characters
// that is not a letter or a digit replaced by a single "-", and no leading or
// trailing "-".
//
//	Slug("Hello, Go World!") == "hello-go-world"
//
// TODO: walk the runes, keeping the word characters and writing at most one
// "-" between runs, then trim any "-" off both ends.
func Slug(s string) string {
	return s
}

// WordCount returns how many space-separated words s holds. An empty or
// all-whitespace string has none.
//
// TODO: one call to strings.Fields is the whole function.
func WordCount(s string) int {
	return 0
}

// isWordChar reports whether r may appear in a slug. Its name is lower case, so
// it is unexported: nothing outside this package can call it, which is exactly
// what a helper wants.
func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}
