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
	// TODO
	return ""
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
	// TODO
	return ""
}

// SumNumbers adds up every value that is a number, ignoring everything else.
// Count int, int64 and float64; a numeric string is not a number here.
func SumNumbers(vals []any) float64 {
	// TODO
	return 0
}

// AsInt returns the int inside v, and false when v holds anything else -
// including an int64 or a float64, which are different types even when they
// hold the same number. This is a single assertion, not a switch.
func AsInt(v any) (int, bool) {
	// TODO
	return 0, false
}

// JoinStrings joins the text in vals, in order, separated by sep. A string
// contributes itself and a fmt.Stringer contributes its String(); everything
// else is skipped.
func JoinStrings(vals []any, sep string) string {
	// TODO
	return ""
}

// Point is here so Describe has something to fall through to.
type Point struct{ X, Y int }

func main() {
	for _, v := range []any{nil, 42, "hi", []int{1, 2}, Celsius(21.5), Point{}} {
		fmt.Println(Describe(v))
	}
	_ = strings.Join
}
