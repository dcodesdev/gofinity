package main

import "fmt"

// Greet returns a greeting for name.
// When name is empty, it greets "World" instead.
func Greet(name string) string {
	if name == "" {
		name = "World"
	}
	return fmt.Sprintf("Hello, %s!", name)
}

func main() {
	fmt.Println(Greet("Gofinity"))
}
