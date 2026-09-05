package main

import "fmt"

// Counter counts things. N is exported so the tests can read it directly.
type Counter struct {
	N int
}

// Value reports the current count. It only reads, so a value receiver is the
// right choice.
func (c Counter) Value() int {
	return c.N
}

// Add increases the count by delta. This one has to be a pointer receiver: a
// value receiver is handed a copy, and the caller would never see the change.
func (c *Counter) Add(delta int) {
	c.N += delta
}

// Plus leaves c alone and returns a new Counter delta higher. This is the
// value-receiver way to "change" something: return the changed copy.
func (c Counter) Plus(delta int) Counter {
	// c is already a copy, so writing to it cannot reach the caller's Counter.
	c.N += delta
	return c
}

// Reset puts the count back to zero.
func (c *Counter) Reset() {
	c.N = 0
}

// SumValues adds up the counts of every counter in the slice by calling Value
// on each one. Value has a value receiver, so ranging over copies is fine.
func SumValues(cs []Counter) int {
	total := 0
	for _, c := range cs {
		total += c.Value()
	}
	return total
}

// AddEach adds delta to every counter in the slice, in place. Add has a pointer
// receiver, so the loop has to reach the slice element itself rather than the
// copy that range hands out.
func AddEach(cs []Counter, delta int) {
	for i := range cs {
		// &cs[i] is the element; cs[i].Add(delta) means the same thing,
		// because a slice element is addressable.
		cs[i].Add(delta)
	}
}

// Temperature is a named type over float64. Methods can be declared on any
// named type defined in this package, not only on structs.
type Temperature float64

// Warmer returns a Temperature d degrees higher. Temperature is not a struct,
// but the value-receiver rule is the same: return the new value.
func (t Temperature) Warmer(d float64) Temperature {
	return t + Temperature(d)
}

// String renders a Temperature as "21.5C", one digit after the point. Calling
// the method String is what makes fmt use it.
func (t Temperature) String() string {
	// float64(t), not t: %.1f on t would call String again and recurse.
	return fmt.Sprintf("%.1fC", float64(t))
}

func main() {
	var c Counter
	c.Add(3)
	c.Add(4)
	fmt.Println(c.Value())
	fmt.Println(Temperature(21.5).Warmer(0.5))
}
