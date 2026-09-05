package main

import "fmt"

// LIFO defers one call per label, each appending its label to order, and
// returns the order those calls actually ran in. Deferring "a", "b", "c"
// gives ["c", "b", "a"].
func LIFO(labels []string) (order []string) {
	// TODO
	return nil
}

// Steps logs "enter", defers two cleanup steps, logs "work", and returns the
// whole log as it stands once the function is finished:
// ["enter", "work", "cleanup-2", "cleanup-1"].
func Steps() (log []string) {
	// TODO
	return nil
}

// CapturedValue starts n at 1, defers a call that takes n as an *argument* and
// records it, then sets n to 99. It returns what the deferred call recorded.
func CapturedValue() (recorded int) {
	// TODO
	return 0
}

// CapturedVariable is the same story, except the deferred function closes over
// n instead of taking it as an argument. It returns what that function read.
func CapturedVariable() (recorded int) {
	// TODO
	return 0
}

// DoubleResult returns n, but a deferred function doubles the named result
// before the caller ever sees it, so DoubleResult(21) is 42.
func DoubleResult(n int) (result int) {
	// TODO
	return 0
}

func main() {
	fmt.Println(LIFO([]string{"a", "b", "c"}))
	fmt.Println(Steps())
	fmt.Println(CapturedValue(), CapturedVariable(), DoubleResult(21))
}
