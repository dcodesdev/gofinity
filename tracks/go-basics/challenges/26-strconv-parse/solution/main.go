package main

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseIntOr parses s as a base-10 integer, ignoring the whitespace around it.
// Anything that does not parse gives fallback.
func ParseIntOr(s string, fallback int) int {
	// Atoi rejects surrounding spaces, so the trim is not optional.
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return fallback
	}
	return n
}

// ParseFloatOr parses s as a 64-bit float, ignoring the whitespace around it.
// Anything that does not parse gives fallback.
func ParseFloatOr(s string, fallback float64) float64 {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return fallback
	}
	return f
}

// ParseBoolOr parses s as a bool - "1", "t", "T", "true", "True", "TRUE" and
// their false counterparts - ignoring the whitespace around it. Anything else
// gives fallback.
func ParseBoolOr(s string, fallback bool) bool {
	b, err := strconv.ParseBool(strings.TrimSpace(s))
	if err != nil {
		return fallback
	}
	return b
}

// SumFields splits csv on commas, parses each field as an integer and adds them
// up. Whitespace around a field is ignored and an empty field is skipped, so
// "" sums to 0. The first field that does not parse stops the sum: the total is
// 0 and the error is the one strconv returned, unchanged.
func SumFields(csv string) (int, error) {
	total := 0
	for _, field := range strings.Split(csv, ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		n, err := strconv.Atoi(field)
		if err != nil {
			// Returned as-is: it is already a *strconv.NumError carrying the
			// input and either ErrSyntax or ErrRange.
			return 0, err
		}
		total += n
	}
	return total, nil
}

// ParseKeyValues reads "a=1,b=2" into map[string]int. Whitespace around a key
// or a value is ignored, an empty entry is skipped, and a repeated key keeps
// the last value. An entry with no "=", an empty key, or a value that does not
// parse is an error.
func ParseKeyValues(s string) (map[string]int, error) {
	out := make(map[string]int)
	for _, entry := range strings.Split(s, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		// Cut splits on the first "=" and reports whether it found one, so a
		// value containing "=" would stay intact.
		key, value, found := strings.Cut(entry, "=")
		key = strings.TrimSpace(key)
		if !found || key == "" {
			return nil, fmt.Errorf("entry %q is not a key=value pair", entry)
		}
		n, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return nil, err
		}
		out[key] = n
	}
	return out, nil
}

// JoinInts renders nums as a comma-separated string with no spaces.
func JoinInts(nums []int) string {
	parts := make([]string, 0, len(nums))
	for _, n := range nums {
		parts = append(parts, strconv.Itoa(n))
	}
	return strings.Join(parts, ",")
}

func main() {
	fmt.Println(ParseIntOr(" 42 ", -1), ParseIntOr("forty", -1))
	fmt.Println(SumFields("1, 2, 3"))
	fmt.Println(ParseKeyValues("a=1, b=2"))
	fmt.Println(JoinInts([]int{1, 2, 3}))
}
