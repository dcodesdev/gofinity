package main

import "fmt"

// Counter counts things. N is exported so the tests can read it directly.
type Counter struct {
	N int
}

// Value reports the current count. It only reads, so a value receiver is the
// right choice.
func (c Counter) Value() int {
	// TODO
	return 0
}

// Add increases the count by delta. This one has to be a pointer receiver: a
// value receiver is handed a copy, and the caller would never see the change.
func (c *Counter) Add(delta int) {
	// TODO
}

// Plus leaves c alone and returns a new Counter delta higher. This is the
// value-receiver way to "change" something: return the changed copy.
func (c Counter) Plus(delta int) Counter {
	// TODO
	return Counter{}
}

// Reset puts the count back to zero.
func (c *Counter) Reset() {
	// TODO
}

// SumValues adds up the counts of every counter in the slice by calling Value
// on each one. Value has a value receiver, so ranging over copies is fine.
func SumValues(cs []Counter) int {
	// TODO
	return 0
}

// AddEach adds delta to every counter in the slice, in place. Add has a pointer
// receiver, so the loop has to reach the slice element itself rather than the
// copy that range hands out.
func AddEach(cs []Counter, delta int) {
	// TODO
}

// Temperature is a named type over float64. Methods can be declared on any
// named type defined in this package, not only on structs.
type Temperature float64

// Warmer returns a Temperature d degrees higher. Temperature is not a struct,
// but the value-receiver rule is the same: return the new value.
func (t Temperature) Warmer(d float64) Temperature {
	// TODO
	return 0
}

// String renders a Temperature as "21.5C", one digit after the point. Calling
// the method String is what makes fmt use it.
func (t Temperature) String() string {
	// TODO
	return ""
}

func main() {
	var c Counter
	c.Add(3)
	c.Add(4)
	fmt.Println(c.Value())
	fmt.Println(Temperature(21.5).Warmer(0.5))
}
