package main

import (
	"fmt"

	"gofinity/splitpackage/textkit"
)

// Headline formats one line of copy: the title-cased text, then its slug in
// parentheses.
//
//	Headline("hello go world") == "Hello Go World (hello-go-world)"
//
// TODO: call textkit.Title and textkit.Slug and build the string. fmt.Sprintf
// is the readable way to do it.
func Headline(s string) string {
	return textkit.Title(s)
}

// Summary reports how many words s holds, in words rather than digits when
// there is nothing to count.
//
//	Summary("go is fun") == "3 words"
//	Summary("go")        == "1 word"
//	Summary("   ")       == "no words"
//
// TODO: ask textkit.WordCount, then switch on the answer.
func Summary(s string) string {
	return ""
}

func main() {
	fmt.Println(Headline("hello go world"))
	fmt.Println(Summary("hello go world"))
}
