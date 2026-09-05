package main

import "fmt"

// AppName is the product name. It is a constant because it never changes while
// the program runs.
const AppName = "Gofinity"

// MaxRetries is how many times a run may be retried.
const MaxRetries = 3

// RetryStatus reports how many of the allowed retries have been used.
//
// For used = 2 and MaxRetries = 3 it returns:
//
//	Gofinity: 2 of 3 retries used, 1 remaining
//
// A negative `used` counts as 0, and anything above MaxRetries counts as
// MaxRetries: the sentence must never claim more retries than exist.
func RetryStatus(used int) string {
	spent := clamp(used)
	left := MaxRetries - spent
	return fmt.Sprintf("%s: %d of %d retries used, %d remaining", AppName, spent, MaxRetries, left)
}

// Remaining returns how many retries are left, never below 0.
func Remaining(used int) int {
	return MaxRetries - clamp(used)
}

// clamp pins a retry count into the range 0..MaxRetries.
func clamp(used int) int {
	if used < 0 {
		return 0
	}
	if used > MaxRetries {
		return MaxRetries
	}
	return used
}

func main() {
	fmt.Println(RetryStatus(2))
}
