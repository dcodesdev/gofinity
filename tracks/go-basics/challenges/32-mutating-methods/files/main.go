package main

import "fmt"

// Stack is a last-in, first-out stack of ints. Its zero value is an empty
// stack that is ready to use: appending to a nil slice allocates one, so no
// constructor is needed.
type Stack struct {
	items []int
}

// Push puts v on top of the stack.
func (s *Stack) Push(v int) {
	// TODO
}

// PushAll pushes every value in order, so the last one ends up on top.
func (s *Stack) PushAll(vs ...int) {
	// TODO
}

// Pop removes the top value and returns it. The second result is false when
// the stack was empty, and the first is then 0.
func (s *Stack) Pop() (int, bool) {
	// TODO
	return 0, false
}

// Peek returns the top value without removing it, and false when the stack is
// empty.
func (s *Stack) Peek() (int, bool) {
	// TODO
	return 0, false
}

// Len reports how many values are on the stack. It only reads, so it takes a
// value receiver - and a value receiver copies the struct, not the backing
// array, so this stays cheap.
func (s Stack) Len() int {
	// TODO
	return 0
}

// Drain pops everything, top first, and returns the values in the order they
// came off. The stack is left empty.
func (s *Stack) Drain() []int {
	// TODO
	return nil
}

// DrainAll drains every stack in the slice in order and returns all the values
// end to end. Drain has a pointer receiver, so the loop must reach the slice
// elements themselves.
func DrainAll(stacks []Stack) []int {
	// TODO
	return nil
}

// Tally counts occurrences by name. Unlike Stack its zero value is not usable,
// because writing to a nil map panics - so NewTally exists and Add has to
// tolerate a Tally that has not been through it.
type Tally struct {
	counts map[string]int
}

// NewTally returns a ready Tally with its map made.
func NewTally() *Tally {
	// TODO
	return nil
}

// Add records one occurrence of name. It makes the map first if the Tally was
// built as a zero value rather than by NewTally.
func (t *Tally) Add(name string) {
	// TODO
}

// Count reports how many times name was added. A missing name is 0.
func (t *Tally) Count(name string) int {
	// TODO
	return 0
}

// Total reports how many occurrences were recorded in total.
func (t *Tally) Total() int {
	// TODO
	return 0
}

// Merge folds every count from other into t, leaving other untouched.
func (t *Tally) Merge(other *Tally) {
	// TODO
}

func main() {
	var s Stack
	s.PushAll(1, 2, 3)
	fmt.Println(s.Drain(), s.Len())

	t := NewTally()
	t.Add("go")
	t.Add("go")
	fmt.Println(t.Count("go"), t.Total())
}
