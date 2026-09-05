package main

import "fmt"

// ByteLen returns the number of bytes s occupies.
func ByteLen(s string) int {
	return len(s)
}

// RuneLen returns the number of runes (Unicode code points) in s.
func RuneLen(s string) int {
	// range over a string decodes UTF-8 and yields one rune per step, so
	// counting the steps counts the runes. utf8.RuneCountInString does the
	// same job without the loop.
	count := 0
	for range s {
		count++
	}
	return count
}

// RuneAt returns the rune at position i counted in runes, not bytes, and
// whether that position exists. RuneAt("héllo", 1) is 'é', true.
func RuneAt(s string, i int) (rune, bool) {
	if i < 0 {
		return 0, false
	}
	at := 0
	for _, r := range s {
		if at == i {
			return r, true
		}
		at++
	}
	return 0, false
}

// ReverseRunes returns s with its runes in reverse order. A multi-byte rune
// must come back out whole.
func ReverseRunes(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func main() {
	fmt.Println(ByteLen("héllo"), RuneLen("héllo"), ReverseRunes("héllo"))
}
