package main

import "fmt"

// LIFO defers one call per label, each appending its label to order, and
// returns the order those calls actually ran in. Deferring "a", "b", "c"
// gives ["c", "b", "a"].
func LIFO(labels []string) (order []string) {
	for _, label := range labels {
		// Every defer pushes onto a stack that unwinds when LIFO returns, so
		// the last one deferred is the first one to run.
		defer func() {
			order = append(order, label)
		}()
	}
	// The bare return sets order to what it is now, an empty slice, and then
	// the deferred calls fill it in. They can, because order is a named result
	// and therefore an ordinary variable this closure captures.
	return
}

// Steps logs "enter", defers two cleanup steps, logs "work", and returns the
// whole log as it stands once the function is finished:
// ["enter", "work", "cleanup-2", "cleanup-1"].
func Steps() (log []string) {
	log = append(log, "enter")
	defer func() { log = append(log, "cleanup-1") }()
	defer func() { log = append(log, "cleanup-2") }()
	log = append(log, "work")
	return
}

// CapturedValue starts n at 1, defers a call that takes n as an *argument* and
// records it, then sets n to 99. It returns what the deferred call recorded.
func CapturedValue() (recorded int) {
	n := 1
	// The argument is evaluated now, at the defer statement, and the copy is
	// stashed with the call. Changing n afterwards cannot reach it.
	defer func(seen int) { recorded = seen }(n)
	n = 99
	_ = n
	return
}

// CapturedVariable is the same story, except the deferred function closes over
// n instead of taking it as an argument. It returns what that function read.
func CapturedVariable() (recorded int) {
	n := 1
	// Nothing is evaluated here except the function value itself. The body runs
	// at return time and reads whatever n holds by then.
	defer func() { recorded = n }()
	n = 99
	return
}

// DoubleResult returns n, but a deferred function doubles the named result
// before the caller ever sees it, so DoubleResult(21) is 42.
func DoubleResult(n int) (result int) {
	defer func() { result *= 2 }()
	result = n
	return result
}

func main() {
	fmt.Println(LIFO([]string{"a", "b", "c"}))
	fmt.Println(Steps())
	fmt.Println(CapturedValue(), CapturedVariable(), DoubleResult(21))
}
