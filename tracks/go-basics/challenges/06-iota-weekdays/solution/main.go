package main

import "fmt"

// Weekday is a day of the week, Sunday first.
type Weekday int

// The seven days, numbered 0 through 6.
const (
	Sunday Weekday = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

// dayNames is indexed by Weekday, so the order here has to match the constants.
var dayNames = [...]string{
	"Sunday",
	"Monday",
	"Tuesday",
	"Wednesday",
	"Thursday",
	"Friday",
	"Saturday",
}

// valid reports whether d names one of the seven days.
func (d Weekday) valid() bool {
	return d >= Sunday && d <= Saturday
}

// String returns the day's English name: "Sunday" through "Saturday".
// A value outside 0..6 formats as "Unknown(<n>)", for example "Unknown(9)".
func (d Weekday) String() string {
	if !d.valid() {
		return fmt.Sprintf("Unknown(%d)", int(d))
	}
	return dayNames[d]
}

// IsWeekend reports whether the day is Saturday or Sunday.
func (d Weekday) IsWeekend() bool {
	return d == Saturday || d == Sunday
}

// Next returns the following day, wrapping Saturday back to Sunday.
// A value outside 0..6 has no following day and comes back unchanged.
func (d Weekday) Next() Weekday {
	if !d.valid() {
		return d
	}
	if d == Saturday {
		return Sunday
	}
	return d + 1
}

func main() {
	fmt.Println(Sunday, Sunday.Next(), Sunday.IsWeekend())
}
