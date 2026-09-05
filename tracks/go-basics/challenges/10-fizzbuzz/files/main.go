package main

import "fmt"

// FizzBuzz classifies one number.
//
//	divisible by 3 and 5 -> "FizzBuzz"
//	divisible by 3       -> "Fizz"
//	divisible by 5       -> "Buzz"
//	anything else        -> the number's decimal digits, e.g. "7"
func FizzBuzz(n int) string {
	// TODO
	return ""
}

// FizzBuzzUpTo returns FizzBuzz(1) through FizzBuzz(n), in order.
// n below 1 gives an empty slice.
func FizzBuzzUpTo(n int) []string {
	// TODO
	return nil
}

// IsLeapYear reports whether y is a leap year in the Gregorian calendar: every
// year divisible by 4, except years divisible by 100, except years divisible
// by 400.
func IsLeapYear(y int) bool {
	// TODO
	return false
}

func main() {
	for _, s := range FizzBuzzUpTo(15) {
		fmt.Println(s)
	}
}
