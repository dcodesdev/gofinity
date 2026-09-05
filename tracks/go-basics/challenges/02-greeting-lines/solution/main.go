package main

import (
	"fmt"
	"strings"
)

// GreetAll returns one numbered greeting line per name, joined by newlines.
//
// Line n (counting from 1) reads "<n>. Hello, <name>!", an empty name is
// greeted as "World", and no names at all give the empty string.
func GreetAll(names []string) string {
	lines := make([]string, 0, len(names))
	for i, name := range names {
		if name == "" {
			name = "World"
		}
		lines = append(lines, fmt.Sprintf("%d. Hello, %s!", i+1, name))
	}
	return strings.Join(lines, "\n")
}

func main() {
	fmt.Println(GreetAll([]string{"Gofinity", "Ada", ""}))
}
