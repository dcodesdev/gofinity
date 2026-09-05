package main

import (
	"fmt"
	"strings"
)

// Celsius is a named float64 with a String method, so it satisfies
// fmt.Stringer. The type switch below has to reach it through that interface,
// not by naming the type.
type Celsius float64

// String renders a Celsius as "21.5C".
func (c Celsius) String() string {
	// float64(c), not c: %.1f on c would call String again and recurse.
	return fmt.Sprintf("%.1fC", float64(c))
}

// Describe reports what is inside an any, in one line per kind:
//
//	nil          -> "nil"
//	int          -> "int 42"
//	float64      -> "float64 3.50"          (two digits after the point)
//	string       -> `string "hi"`           (quoted)
//	bool         -> "bool true"
//	[]int        -> "[]int of 3"            (the length)
//	fmt.Stringer -> "stringer 21.5C"        (whatever String returns)
//	anything else -> "other main.Point"     (the type name, from %T)
//
// A type switch is `switch v := v.(type)`, and inside each case v already has
// that type. Order matters: a case naming an interface matches anything that
// satisfies it, so it must come after the concrete cases you want to catch
// first.
func Describe(v any) string {
	switch v := v.(type) {
	case nil:
		// The nil case matches an interface with no type in it at all. It
		// cannot be written as a value comparison inside another case.
		return "nil"
	case int:
		return fmt.Sprintf("int %d", v)
	case float64:
		return fmt.Sprintf("float64 %.2f", v)
	case string:
		return fmt.Sprintf("string %q", v)
	case bool:
		return fmt.Sprintf("bool %t", v)
	case []int:
		return fmt.Sprintf("[]int of %d", len(v))
	case fmt.Stringer:
		// Last of the matching cases on purpose: a Celsius reaches this one,
		// but so would a string type with a String method, and the concrete
		// cases above get first refusal.
		return "stringer " + v.String()
	default:
		// v is the original any here, which is what %T wants.
		return fmt.Sprintf("other %T", v)
	}
}

// SumNumbers adds up every value that is a number, ignoring everything else.
// Count int, int64 and float64; a numeric string is not a number here.
func SumNumbers(vals []any) float64 {
	total := 0.0
	for _, v := range vals {
		// One case can list several types. When it does, v keeps the
		// interface type, so each branch still needs its own conversion -
		// which is why these are written separately.
		switch v := v.(type) {
		case int:
			total += float64(v)
		case int64:
			total += float64(v)
		case float64:
			total += v
		}
	}
	return total
}

// AsInt returns the int inside v, and false when v holds anything else -
// including an int64 or a float64, which are different types even when they
// hold the same number. This is a single assertion, not a switch.
func AsInt(v any) (int, bool) {
	n, ok := v.(int)
	return n, ok
}

// JoinStrings joins the text in vals, in order, separated by sep. A string
// contributes itself and a fmt.Stringer contributes its String(); everything
// else is skipped.
func JoinStrings(vals []any, sep string) string {
	parts := make([]string, 0, len(vals))
	for _, v := range vals {
		switch v := v.(type) {
		case string:
			parts = append(parts, v)
		case fmt.Stringer:
			parts = append(parts, v.String())
		}
	}
	return strings.Join(parts, sep)
}

// Point is here so Describe has something to fall through to.
type Point struct{ X, Y int }

func main() {
	for _, v := range []any{nil, 42, "hi", []int{1, 2}, Celsius(21.5), Point{}} {
		fmt.Println(Describe(v))
	}
	_ = strings.Join
}
