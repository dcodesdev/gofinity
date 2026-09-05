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
	s.items = append(s.items, v)
}

// PushAll pushes every value in order, so the last one ends up on top.
func (s *Stack) PushAll(vs ...int) {
	for _, v := range vs {
		s.Push(v)
	}
}

// Pop removes the top value and returns it. The second result is false when
// the stack was empty, and the first is then 0.
func (s *Stack) Pop() (int, bool) {
	if len(s.items) == 0 {
		return 0, false
	}
	last := len(s.items) - 1
	v := s.items[last]
	// Reslicing is the removal. The value stays in the backing array, which
	// matters only when it holds a pointer worth releasing.
	s.items = s.items[:last]
	return v, true
}

// Peek returns the top value without removing it, and false when the stack is
// empty.
func (s *Stack) Peek() (int, bool) {
	if len(s.items) == 0 {
		return 0, false
	}
	return s.items[len(s.items)-1], true
}

// Len reports how many values are on the stack. It only reads, so it takes a
// value receiver - and a value receiver copies the struct, not the backing
// array, so this stays cheap.
func (s Stack) Len() int {
	return len(s.items)
}

// Drain pops everything, top first, and returns the values in the order they
// came off. The stack is left empty.
func (s *Stack) Drain() []int {
	out := []int{}
	for {
		v, ok := s.Pop()
		if !ok {
			return out
		}
		out = append(out, v)
	}
}

// DrainAll drains every stack in the slice in order and returns all the values
// end to end. Drain has a pointer receiver, so the loop must reach the slice
// elements themselves.
func DrainAll(stacks []Stack) []int {
	out := []int{}
	for i := range stacks {
		// stacks[i] is addressable, so this is (&stacks[i]).Drain(). Ranging
		// over values would drain copies and leave the originals full.
		out = append(out, stacks[i].Drain()...)
	}
	return out
}

// Tally counts occurrences by name. Unlike Stack its zero value is not usable,
// because writing to a nil map panics - so NewTally exists and Add has to
// tolerate a Tally that has not been through it.
type Tally struct {
	counts map[string]int
}

// NewTally returns a ready Tally with its map made.
func NewTally() *Tally {
	return &Tally{counts: map[string]int{}}
}

// Add records one occurrence of name. It makes the map first if the Tally was
// built as a zero value rather than by NewTally.
func (t *Tally) Add(name string) {
	if t.counts == nil {
		t.counts = map[string]int{}
	}
	t.counts[name]++
}

// Count reports how many times name was added. A missing name is 0.
func (t *Tally) Count(name string) int {
	// Reading a nil map is fine; only writing panics.
	return t.counts[name]
}

// Total reports how many occurrences were recorded in total.
func (t *Tally) Total() int {
	total := 0
	for _, n := range t.counts {
		total += n
	}
	return total
}

// Merge folds every count from other into t, leaving other untouched.
func (t *Tally) Merge(other *Tally) {
	if other == nil {
		return
	}
	if t.counts == nil {
		t.counts = map[string]int{}
	}
	for name, n := range other.counts {
		t.counts[name] += n
	}
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
