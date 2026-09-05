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
	// A switch with no expression switches on `true`, so each case is a
	// condition. It is the readable form of a long if/else if chain.
	switch {
	case score < 0 || score > 100:
		return "?"
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}

// DayKind classifies a lowercase three-letter day abbreviation.
// "sat" and "sun" are "weekend", the other five are "weekday", and anything
// else is "unknown".
func DayKind(day string) string {
	switch day {
	case "sat", "sun":
		return "weekend"
	case "mon", "tue", "wed", "thu", "fri":
		return "weekday"
	default:
		return "unknown"
	}
}

// Perks lists what a membership tier includes, most exclusive first.
//
//	"gold"   -> ["lounge", "priority", "points"]
//	"silver" -> ["priority", "points"]
//	"bronze" -> ["points"]
//
// Any other tier gets no perks at all.
func Perks(tier string) []string {
	var perks []string

	// The one place `fallthrough` earns its keep: each tier includes every
	// tier below it, so the cases stack instead of repeating each other.
	switch tier {
	case "gold":
		perks = append(perks, "lounge")
		fallthrough
	case "silver":
		perks = append(perks, "priority")
		fallthrough
	case "bronze":
		perks = append(perks, "points")
	}

	return perks
}

// Season names the meteorological season of a month number, 1 through 12.
// December, January and February are "winter"; March to May "spring";
// June to August "summer"; September to November "autumn". A month outside
// 1..12 is "unknown".
func Season(month int) string {
	switch month {
	case 12, 1, 2:
		return "winter"
	case 3, 4, 5:
		return "spring"
	case 6, 7, 8:
		return "summer"
	case 9, 10, 11:
		return "autumn"
	default:
		return "unknown"
	}
}

func main() {
	fmt.Println(Grade(87), DayKind("sat"), Season(12), Perks("silver"))
}
