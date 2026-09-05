package main

import (
	"errors"
	"fmt"
)

// ErrNotPositive is the panic value MustPositive uses.
var ErrNotPositive = errors.New("not positive")

// SafeDivide returns a / b. Dividing by zero panics at runtime; catch that
// panic and return it as an error of the form "safe divide: <what happened>",
// with a result of 0.
func SafeDivide(a, b int) (result int, err error) {
	// recover only stops a panic when it is called from a deferred function of
	// the panicking call. Calling it inline returns nil and changes nothing.
	defer func() {
		if r := recover(); r != nil {
			// The results are already set by then, so reset both. A caller who
			// ignores err must not get a plausible-looking quotient.
			result = 0
			err = fmt.Errorf("safe divide: %v", r)
		}
	}()
	return a / b, nil
}

// MustPositive returns n, or panics with ErrNotPositive when n is zero or
// negative.
func MustPositive(n int) int {
	if n <= 0 {
		// A panic value is an any. An error is the most useful thing to put
		// there, because whoever recovers it can hand it straight back.
		panic(ErrNotPositive)
	}
	return n
}

// SafeIndex returns nums[i]. An index outside the slice panics at runtime;
// catch it and return 0 and an error of the form "safe index: <what happened>".
func SafeIndex(nums []int, i int) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			n = 0
			err = fmt.Errorf("safe index: %v", r)
		}
	}()
	return nums[i], nil
}

// Guard runs body and turns a panic into an error. A panic value that is
// already an error is returned as-is; anything else is formatted with %v into
// an error of the form "panic: <value>". Guard returns nil when body returns
// normally.
func Guard(body func()) (err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		if e, ok := r.(error); ok {
			err = e
			return
		}
		err = fmt.Errorf("panic: %v", r)
	}()
	body()
	return nil
}

func main() {
	fmt.Println(SafeDivide(7, 2))
	fmt.Println(SafeDivide(7, 0))
	fmt.Println(Guard(func() { MustPositive(-1) }))
}
