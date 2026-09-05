package main

import "fmt"

// AppName is the product name. It is a constant because it never changes while
// the program runs.
//
// TODO: give it the value "Gofinity".
const AppName = ""

// MaxRetries is how many times a run may be retried.
//
// TODO: give it the value 3.
const MaxRetries = 0

// RetryStatus reports how many of the allowed retries have been used.
//
// For used = 2 and MaxRetries = 3 it returns:
//
//	Gofinity: 2 of 3 retries used, 1 remaining
//
// A negative `used` counts as 0, and anything above MaxRetries counts as
// MaxRetries: the sentence must never claim more retries than exist.
func RetryStatus(used int) string {
	// TODO: clamp `used`, then format the sentence with fmt.Sprintf.
	return ""
}

// Remaining returns how many retries are left, never below 0.
func Remaining(used int) int {
	// TODO
	return 0
}

func main() {
	fmt.Println(RetryStatus(2))
}
