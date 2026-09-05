package main

import (
	"fmt"
	"sort"
	"strings"
)

// Append adds value to the slice stored under key, creating the entry when the
// key is new. The caller sees the change; a nil map is left alone rather than
// panicking.
func Append(groups map[string][]string, key, value string) {
	if groups == nil {
		return
	}
	// A missing key reads as a nil slice, and append on a nil slice allocates,
	// so this one line handles both the new key and the existing one. The
	// assignment back into the map is what is easy to forget: append may
	// return a different header, and the map holds a copy of it.
	groups[key] = append(groups[key], value)
}

// GroupByFirstLetter groups words by their first letter, lowercased, keeping
// the input order inside each group. Empty strings are skipped.
func GroupByFirstLetter(words []string) map[string][]string {
	groups := make(map[string][]string)
	for _, word := range words {
		if word == "" {
			continue
		}
		key := strings.ToLower(word[:1])
		groups[key] = append(groups[key], word)
	}
	return groups
}

// Keys returns every key of groups in ascending order.
func Keys(groups map[string][]string) []string {
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Count returns the total number of members across every group.
func Count(groups map[string][]string) int {
	total := 0
	for _, members := range groups {
		total += len(members)
	}
	return total
}

// Clone returns a deep copy: changing a slice in the result must not change the
// original, and vice versa.
func Clone(groups map[string][]string) map[string][]string {
	out := make(map[string][]string, len(groups))
	for k, members := range groups {
		// Copying the map alone would copy the slice headers, which still
		// point at the original arrays. Each group needs its own.
		out[k] = append([]string(nil), members...)
	}
	return out
}

// Merge returns a new map holding every member of a followed by every member of
// b under the same key. Neither argument is modified, and the result shares no
// backing array with either.
func Merge(a, b map[string][]string) map[string][]string {
	// Clone first, so every slice in the result is already ours. Appending to
	// a borrowed slice with spare capacity would write into the caller's array.
	out := Clone(a)
	for k, members := range b {
		out[k] = append(out[k], members...)
	}
	return out
}

// Invert turns a group-to-members map into a member-to-groups map. Each
// member's group list comes back in ascending order.
func Invert(groups map[string][]string) map[string][]string {
	out := make(map[string][]string)
	// Range over the keys in sorted order, so each member's group list comes
	// out sorted without a second pass.
	for _, group := range Keys(groups) {
		for _, member := range groups[group] {
			out[member] = append(out[member], group)
		}
	}
	return out
}

func main() {
	groups := GroupByFirstLetter([]string{"go", "gopher", "rust", "ruby", "zig"})
	for _, k := range Keys(groups) {
		fmt.Println(k, groups[k])
	}
}
