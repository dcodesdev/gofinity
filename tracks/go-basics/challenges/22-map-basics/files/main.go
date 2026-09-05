package main

import "fmt"

// Lookup returns the value stored under key and whether it was there at all.
// A missing key gives 0 and false, and so does a nil map.
func Lookup(m map[string]int, key string) (int, bool) {
	// TODO
	return 0, false
}

// Get returns the value stored under key, or fallback when the key is missing.
func Get(m map[string]int, key string, fallback int) int {
	// TODO
	return 0
}

// Add increases the value under key by n, starting from zero when the key is
// new. A map is a reference, so the caller sees the change.
func Add(m map[string]int, key string, n int) {
	// TODO
}

// Remove deletes key from m and reports whether it was present beforehand.
func Remove(m map[string]int, key string) bool {
	// TODO
	return false
}

// SortedKeys returns every key of m in ascending order. Range order over a map
// is random, so the sort is what makes the result worth returning.
func SortedKeys(m map[string]int) []string {
	// TODO
	return nil
}

// Total adds up every value in m.
func Total(m map[string]int) int {
	// TODO
	return 0
}

// FromPairs builds a map from parallel slices of keys and values. When the two
// slices have different lengths it returns nil, and a later key wins over an
// earlier duplicate.
func FromPairs(keys []string, values []int) map[string]int {
	// TODO
	return nil
}

func main() {
	scores := map[string]int{"go": 3, "rust": 2}
	Add(scores, "go", 1)
	fmt.Println(SortedKeys(scores), Total(scores))
	fmt.Println(Get(scores, "zig", -1))
}
