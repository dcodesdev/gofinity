package main

import "fmt"

// Append adds value to the slice stored under key, creating the entry when the
// key is new. The caller sees the change; a nil map is left alone rather than
// panicking.
func Append(groups map[string][]string, key, value string) {
	// TODO
}

// GroupByFirstLetter groups words by their first letter, lowercased, keeping
// the input order inside each group. Empty strings are skipped.
func GroupByFirstLetter(words []string) map[string][]string {
	// TODO
	return nil
}

// Keys returns every key of groups in ascending order.
func Keys(groups map[string][]string) []string {
	// TODO
	return nil
}

// Count returns the total number of members across every group.
func Count(groups map[string][]string) int {
	// TODO
	return 0
}

// Clone returns a deep copy: changing a slice in the result must not change the
// original, and vice versa.
func Clone(groups map[string][]string) map[string][]string {
	// TODO
	return nil
}

// Merge returns a new map holding every member of a followed by every member of
// b under the same key. Neither argument is modified, and the result shares no
// backing array with either.
func Merge(a, b map[string][]string) map[string][]string {
	// TODO
	return nil
}

// Invert turns a group-to-members map into a member-to-groups map. Each
// member's group list comes back in ascending order.
func Invert(groups map[string][]string) map[string][]string {
	// TODO
	return nil
}

func main() {
	groups := GroupByFirstLetter([]string{"go", "gopher", "rust", "ruby", "zig"})
	for _, k := range Keys(groups) {
		fmt.Println(k, groups[k])
	}
}
