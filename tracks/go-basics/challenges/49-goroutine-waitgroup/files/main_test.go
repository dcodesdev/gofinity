package main

import (
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// waitFor runs fn in its own goroutine and reports whether it finished. It is
// how every test here refuses to hang: a RunAll that forgot to Wait finishes
// instantly, and a RunAll that forgot to Add hangs, and both should be a
// failure with a message rather than a stuck test binary.
func waitFor(t *testing.T, what string, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("%s did not return within a second", what)
	}
}

func TestRunAllRunsEveryTask(t *testing.T) {
	const n = 50
	var count atomic.Int64
	tasks := make([]func(), n)
	for i := range tasks {
		tasks[i] = func() { count.Add(1) }
	}
	waitFor(t, "RunAll", func() { RunAll(tasks) })
	if got := count.Load(); got != n {
		t.Errorf("after RunAll, %d of %d tasks had run - RunAll must wait for all of them", got, n)
	}
}

func TestRunAllOfNothingReturns(t *testing.T) {
	waitFor(t, "RunAll(nil)", func() { RunAll(nil) })
	waitFor(t, "RunAll of an empty slice", func() { RunAll([]func(){}) })
}

func TestRunAllRunsTasksConcurrently(t *testing.T) {
	const n = 8
	// Every task reports that it started, then blocks until all n have. If
	// RunAll ran them one after another the first task would wait forever, so
	// only a concurrent RunAll can finish this.
	var started sync.WaitGroup
	started.Add(n)
	release := make(chan struct{})
	go func() {
		started.Wait()
		close(release)
	}()

	tasks := make([]func(), n)
	for i := range tasks {
		tasks[i] = func() {
			started.Done()
			<-release
		}
	}
	waitFor(t, "RunAll", func() { RunAll(tasks) })
}

func TestSquares(t *testing.T) {
	want := []int{0, 1, 4, 9, 16, 25}
	var got []int
	waitFor(t, "Squares", func() { got = Squares(6) })
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Squares(6) = %v, want %v", got, want)
	}
}

func TestSquaresOfZeroIsEmptyNotNil(t *testing.T) {
	var got []int
	waitFor(t, "Squares", func() { got = Squares(0) })
	if got == nil {
		t.Fatal("Squares(0) = nil, want an empty slice")
	}
	if len(got) != 0 {
		t.Errorf("Squares(0) = %v, want an empty slice", got)
	}
}

func TestSquaresIsStable(t *testing.T) {
	// Every goroutine owns one index, so the answer cannot depend on the order
	// they happen to run in. Repeating catches a result assembled by appending.
	want := make([]int, 200)
	for i := range want {
		want[i] = i * i
	}
	for range 20 {
		var got []int
		waitFor(t, "Squares", func() { got = Squares(200) })
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Squares(200) came back wrong or out of order: %v", got)
		}
	}
}

func TestGatherKeepsInputOrder(t *testing.T) {
	fns := []func() int{
		func() int { return 10 },
		func() int { return 20 },
		func() int { return 30 },
		func() int { return 40 },
	}
	var got []int
	waitFor(t, "Gather", func() { got = Gather(fns) })
	want := []int{10, 20, 30, 40}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Gather = %v, want %v", got, want)
	}
}

func TestGatherKeepsInputOrderWhenTheLastOneFinishesFirst(t *testing.T) {
	const n = 6
	// Each function blocks until every other one has started, so they finish in
	// whatever order the scheduler picks. The output order must still be the
	// input order.
	var started sync.WaitGroup
	started.Add(n)
	release := make(chan struct{})
	go func() {
		started.Wait()
		close(release)
	}()

	fns := make([]func() int, n)
	want := make([]int, n)
	for i := range fns {
		want[i] = i * 100
		fns[i] = func() int {
			started.Done()
			<-release
			return i * 100
		}
	}
	var got []int
	waitFor(t, "Gather", func() { got = Gather(fns) })
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Gather = %v, want %v - results go at their own index, not in finishing order", got, want)
	}
}

func TestGatherCallsEachFunctionExactlyOnce(t *testing.T) {
	const n = 30
	var calls atomic.Int64
	fns := make([]func() int, n)
	for i := range fns {
		fns[i] = func() int {
			calls.Add(1)
			return i
		}
	}
	var got []int
	waitFor(t, "Gather", func() { got = Gather(fns) })
	if c := calls.Load(); c != n {
		t.Errorf("Gather made %d calls, want %d", c, n)
	}
	if len(got) != n {
		t.Errorf("Gather returned %d results, want %d", len(got), n)
	}
}

func TestGatherOfNothing(t *testing.T) {
	var got []int
	waitFor(t, "Gather", func() { got = Gather(nil) })
	if len(got) != 0 {
		t.Errorf("Gather(nil) = %v, want no results", got)
	}
}
