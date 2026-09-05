package main

import "fmt"

// Weekday is a day of the week, Sunday first.
type Weekday int

// The seven days, numbered 0 through 6.
//
// TODO: replace these placeholders with a single iota block, so that Sunday is
// 0, Monday is 1, and Saturday is 6.
const (
	Sunday    Weekday = 0
	Monday    Weekday = 0
	Tuesday   Weekday = 0
	Wednesday Weekday = 0
	Thursday  Weekday = 0
	Friday    Weekday = 0
	Saturday  Weekday = 0
)

// String returns the day's English name: "Sunday" through "Saturday".
// A value outside 0..6 formats as "Unknown(<n>)", for example "Unknown(9)".
func (d Weekday) String() string {
	// TODO
	return ""
}

// IsWeekend reports whether the day is Saturday or Sunday.
func (d Weekday) IsWeekend() bool {
	// TODO
	return false
}

// Next returns the following day, wrapping Saturday back to Sunday.
// A value outside 0..6 has no following day and comes back unchanged.
func (d Weekday) Next() Weekday {
	// TODO
	return d
}

func main() {
	fmt.Println(Sunday, Sunday.Next(), Sunday.IsWeekend())
}
