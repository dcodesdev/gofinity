package main

import "fmt"

// Grade turns a percentage into a letter.
//
//	90 and above -> "A"
//	80 to 89     -> "B"
//	70 to 79     -> "C"
//	60 to 69     -> "D"
//	below 60     -> "F"
//
// A score outside 0..100 is not a score: return "?".
func Grade(score int) string {
	// TODO
	return ""
}

// DayKind classifies a lowercase three-letter day abbreviation.
// "sat" and "sun" are "weekend", the other five are "weekday", and anything
// else is "unknown".
func DayKind(day string) string {
	// TODO
	return ""
}

// Perks lists what a membership tier includes, most exclusive first.
//
//	"gold"   -> ["lounge", "priority", "points"]
//	"silver" -> ["priority", "points"]
//	"bronze" -> ["points"]
//
// Any other tier gets no perks at all.
func Perks(tier string) []string {
	// TODO
	return nil
}

// Season names the meteorological season of a month number, 1 through 12.
// December, January and February are "winter"; March to May "spring";
// June to August "summer"; September to November "autumn". A month outside
// 1..12 is "unknown".
func Season(month int) string {
	// TODO
	return ""
}

func main() {
	fmt.Println(Grade(87), DayKind("sat"), Season(12), Perks("silver"))
}
