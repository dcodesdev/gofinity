package main

import (
	"fmt"
	"sort"
)

// Lookup returns the value stored under key and whether it was there at all.
// A missing key gives 0 and false, and so does a nil map.
func Lookup(m map[string]int, key string) (int, bool) {
	// The comma-ok form is the whole answer: the second result is the only way
	// to tell a stored zero from a key that was never there.
	v, ok := m[key]
	return v, ok
}

// Get returns the value stored under key, or fallback when the key is missing.
func Get(m map[string]int, key string, fallback int) int {
	if v, ok := m[key]; ok {
		return v
	}
	return fallback
}

// Add increases the value under key by n, starting from zero when the key is
// new. A map is a reference, so the caller sees the change.
func Add(m map[string]int, key string, n int) {
	// A missing key reads as the zero value, so this needs no "if present"
	// branch. It would panic on a nil map, which is why every caller here
	// makes one first.
	m[key] += n
}

// Remove deletes key from m and reports whether it was present beforehand.
func Remove(m map[string]int, key string) bool {
	if _, ok := m[key]; !ok {
		return false
	}
	delete(m, key)
	return true
}

// SortedKeys returns every key of m in ascending order. Range order over a map
// is random, so the sort is what makes the result worth returning.
func SortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Total adds up every value in m.
func Total(m map[string]int) int {
	total := 0
	for _, v := range m {
		total += v
	}
	return total
}

// FromPairs builds a map from parallel slices of keys and values. When the two
// slices have different lengths it returns nil, and a later key wins over an
// earlier duplicate.
func FromPairs(keys []string, values []int) map[string]int {
	if len(keys) != len(values) {
		return nil
	}
	// make, not a var declaration: a nil map cannot be written to.
	m := make(map[string]int, len(keys))
	for i, k := range keys {
		m[k] = values[i]
	}
	return m
}

func main() {
	scores := map[string]int{"go": 3, "rust": 2}
	Add(scores, "go", 1)
	fmt.Println(SortedKeys(scores), Total(scores))
	fmt.Println(Get(scores, "zig", -1))
}
