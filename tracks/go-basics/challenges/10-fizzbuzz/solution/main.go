package main

import (
	"fmt"
	"strconv"
)

// FizzBuzz classifies one number.
//
//	divisible by 3 and 5 -> "FizzBuzz"
//	divisible by 3       -> "Fizz"
//	divisible by 5       -> "Buzz"
//	anything else        -> the number's decimal digits, e.g. "7"
func FizzBuzz(n int) string {
	// The combined case has to come first: 15 is divisible by 3, so an
	// `if n%3 == 0` above it would answer "Fizz" and never look further.
	if n%15 == 0 {
		return "FizzBuzz"
	}
	if n%3 == 0 {
		return "Fizz"
	}
	if n%5 == 0 {
		return "Buzz"
	}
	return strconv.Itoa(n)
}

// FizzBuzzUpTo returns FizzBuzz(1) through FizzBuzz(n), in order.
// n below 1 gives an empty slice.
func FizzBuzzUpTo(n int) []string {
	out := []string{}
	for i := 1; i <= n; i++ {
		out = append(out, FizzBuzz(i))
	}
	return out
}

// IsLeapYear reports whether y is a leap year in the Gregorian calendar: every
// year divisible by 4, except years divisible by 100, except years divisible
// by 400.
func IsLeapYear(y int) bool {
	if y%400 == 0 {
		return true
	}
	if y%100 == 0 {
		return false
	}
	return y%4 == 0
}

func main() {
	for _, s := range FizzBuzzUpTo(15) {
		fmt.Println(s)
	}
}
