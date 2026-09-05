package main

import "fmt"

// ByteLen returns the number of bytes s occupies.
func ByteLen(s string) int {
	// TODO
	return 0
}

// RuneLen returns the number of runes (Unicode code points) in s.
func RuneLen(s string) int {
	// TODO
	return 0
}

// RuneAt returns the rune at position i counted in runes, not bytes, and
// whether that position exists. RuneAt("héllo", 1) is 'é', true.
func RuneAt(s string, i int) (rune, bool) {
	// TODO
	return 0, false
}

// ReverseRunes returns s with its runes in reverse order. A multi-byte rune
// must come back out whole.
func ReverseRunes(s string) string {
	// TODO
	return ""
}

func main() {
	fmt.Println(ByteLen("héllo"), RuneLen("héllo"), ReverseRunes("héllo"))
}
