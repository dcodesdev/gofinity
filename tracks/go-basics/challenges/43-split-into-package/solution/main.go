package main

import (
	"fmt"

	"gofinity/splitpackage/textkit"
)

// Headline formats one line of copy: the title-cased text, then its slug in
// parentheses.
func Headline(s string) string {
	return fmt.Sprintf("%s (%s)", textkit.Title(s), textkit.Slug(s))
}

// Summary reports how many words s holds.
func Summary(s string) string {
	switch n := textkit.WordCount(s); n {
	case 0:
		return "no words"
	case 1:
		return "1 word"
	default:
		return fmt.Sprintf("%d words", n)
	}
}

func main() {
	fmt.Println(Headline("hello go world"))
	fmt.Println(Summary("hello go world"))
}
