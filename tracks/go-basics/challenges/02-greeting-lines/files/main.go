package main

import "fmt"

// GreetAll returns one numbered greeting line per name, joined by newlines.
//
// Line n (counting from 1) reads "<n>. Hello, <name>!", an empty name is
// greeted as "World", and no names at all give the empty string.
func GreetAll(names []string) string {
	// TODO: build one line per name, then join them with "\n".
	return ""
}

func main() {
	fmt.Println(GreetAll([]string{"Gofinity", "Ada", ""}))
}
