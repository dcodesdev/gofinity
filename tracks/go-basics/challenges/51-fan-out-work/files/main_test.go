package main

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// waitFor runs fn in its own goroutine so a fan-out that never finishes fails
// with a message instead of hanging the whole run.
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

func TestAssignmentsStrides(t *testing.T) {
	cases := []struct {
		items, workers int
		want           [][]int
	}{
		{9, 3, [][]int{{0, 3, 6}, {1, 4, 7}, {2, 5, 8}}},
		{10, 3, [][]int{{0, 3, 6, 9}, {1, 4, 7}, {2, 5, 8}}},
		{5, 1, [][]int{{0, 1, 2, 3, 4}}},
		{4, 4, [][]int{{0}, {1}, {2}, {3}}},
		{4, 6, [][]int{{0}, {1}, {2}, {3}}},
		{3, 0, [][]int{{0, 1, 2}}},
		{3, -2, [][]int{{0, 1, 2}}},
	}
	for _, c := range cases {
		got := Assignments(c.items, c.workers)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("Assignments(%d, %d) = %v, want %v", c.items, c.workers, got, c.want)
		}
	}
}

func TestAssignmentsOfNothing(t *testing.T) {
	if got := Assignments(0, 4); len(got) != 0 {
		t.Errorf("Assignments(0, 4) = %v, want no workers", got)
	}
	if got := Assignments(-1, 4); len(got) != 0 {
		t.Errorf("Assignments(-1, 4) = %v, want no workers", got)
	}
}

func TestAssignmentsCoverEveryIndexExactlyOnce(t *testing.T) {
	for items := 1; items <= 30; items++ {
		for workers := 1; workers <= 10; workers++ {
			lists := Assignments(items, workers)
			var all []int
			for _, list := range lists {
				if len(list) == 0 {
					t.Fatalf("Assignments(%d, %d) = %v gave a worker nothing to do", items, workers, lists)
				}
				all = append(all, list...)
			}
			sort.Ints(all)
			if len(all) != items {
				t.Fatalf("Assignments(%d, %d) covers %d indices, want %d", items, workers, len(all), items)
			}
			for i, v := range all {
				if v != i {
					t.Fatalf("Assignments(%d, %d) = %v is not a partition of 0..%d", items, workers, lists, items-1)
				}
			}
		}
	}
}

func TestAssignmentsAreBalanced(t *testing.T) {
	// Striding is what keeps the lists within one of each other, whatever the
	// remainder is.
	for items := 1; items <= 30; items++ {
		for workers := 1; workers <= 10; workers++ {
			lists := Assignments(items, workers)
			smallest, largest := items, 0
			for _, list := range lists {
				smallest = min(smallest, len(list))
				largest = max(largest, len(list))
			}
			if largest-smallest > 1 {
				t.Fatalf("Assignments(%d, %d) = %v is uneven: %d..%d items", items, workers, lists, smallest, largest)
			}
		}
	}
}

func TestFanOutKeepsInputOrder(t *testing.T) {
	in := make([]int, 100)
	want := make([]int, 100)
	for i := range in {
		in[i] = i
		want[i] = i * i
	}
	var got []int
	waitFor(t, "FanOut", func() { got = FanOut(in, 7, func(v int) int { return v * v }) })
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FanOut = %v, want %v", got, want)
	}
}

func TestFanOutChangesType(t *testing.T) {
	var got []string
	waitFor(t, "FanOut", func() {
		got = FanOut([]int{1, 2, 3}, 3, func(v int) string { return fmt.Sprintf("n%d", v) })
	})
	want := []string{"n1", "n2", "n3"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FanOut = %v, want %v", got, want)
	}
}

func TestFanOutCallsFOncePerElement(t *testing.T) {
	in := make([]int, 199)
	var calls atomic.Int64
	var got []int
	waitFor(t, "FanOut", func() {
		got = FanOut(in, 8, func(v int) int {
			calls.Add(1)
			return v
		})
	})
	if c := calls.Load(); c != int64(len(in)) {
		t.Errorf("f was called %d times, want %d", c, len(in))
	}
	if len(got) != len(in) {
		t.Errorf("FanOut returned %d results, want %d", len(got), len(in))
	}
}

func TestFanOutOfNothingIsEmptyNotNil(t *testing.T) {
	var got []int
	waitFor(t, "FanOut", func() { got = FanOut([]int{}, 4, func(v int) int { return v }) })
	if got == nil {
		t.Fatal("FanOut of an empty slice = nil, want an empty slice")
	}
	if len(got) != 0 {
		t.Errorf("FanOut of an empty slice = %v, want no results", got)
	}
}

func TestFanOutHandlesOddWorkerCounts(t *testing.T) {
	in := []int{1, 2, 3}
	want := []int{2, 4, 6}
	for _, workers := range []int{0, -4, 1, 2, 3, 9} {
		var got []int
		waitFor(t, "FanOut", func() { got = FanOut(in, workers, func(v int) int { return v * 2 }) })
		if !reflect.DeepEqual(got, want) {
			t.Errorf("FanOut(in, %d, double) = %v, want %v", workers, got, want)
		}
	}
}

func TestFanOutIsActuallyConcurrent(t *testing.T) {
	const workers = 4
	// One item per worker, and each call blocks until all four have started.
	// A sequential FanOut would wait on the first one forever.
	var started sync.WaitGroup
	started.Add(workers)
	release := make(chan struct{})
	go func() {
		started.Wait()
		close(release)
	}()

	in := []int{0, 1, 2, 3}
	var got []int
	waitFor(t, "FanOut", func() {
		got = FanOut(in, workers, func(v int) int {
			started.Done()
			<-release
			return v
		})
	})
	if !reflect.DeepEqual(got, in) {
		t.Errorf("FanOut = %v, want %v", got, in)
	}
}

func TestFanOutIsStable(t *testing.T) {
	in := make([]int, 200)
	want := make([]int, 200)
	for i := range in {
		in[i] = i
		want[i] = i + 1
	}
	for range 30 {
		got := FanOut(in, 16, func(v int) int { return v + 1 })
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("FanOut is not stable: %v", got)
		}
	}
}

func TestFanOutErrSucceeds(t *testing.T) {
	var got []int
	var err error
	waitFor(t, "FanOutErr", func() {
		got, err = FanOutErr([]string{"1", "2", "3", "4"}, 3, strconv.Atoi)
	})
	if err != nil {
		t.Fatalf("FanOutErr returned an error: %v", err)
	}
	want := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FanOutErr = %v, want %v", got, want)
	}
}

func TestFanOutErrReturnsTheLowestFailingIndex(t *testing.T) {
	// Three items fail. Whichever goroutine got there first, the answer must be
	// the one at the lowest index, on every run.
	in := []string{"ok", "ok", "bad-2", "ok", "bad-4", "ok", "bad-6", "ok"}
	for range 30 {
		got, err := FanOutErr(in, 4, func(s string) (string, error) {
			if len(s) > 2 {
				return "", errors.New(s)
			}
			return s, nil
		})
		if err == nil {
			t.Fatal("FanOutErr returned no error, want the one from index 2")
		}
		if err.Error() != "bad-2" {
			t.Fatalf("FanOutErr returned %q, want %q - the lowest failing index wins", err, "bad-2")
		}
		if got != nil {
			t.Fatalf("FanOutErr returned results %v alongside an error, want nil", got)
		}
	}
}

func TestFanOutErrStillCallsFOnEveryElement(t *testing.T) {
	// No early exit: cancelling the rest of the work needs a channel and then a
	// context, and neither is here yet.
	in := make([]int, 50)
	var calls atomic.Int64
	_, err := FanOutErr(in, 5, func(v int) (int, error) {
		calls.Add(1)
		return 0, errors.New("always")
	})
	if err == nil {
		t.Fatal("FanOutErr returned no error, want one")
	}
	if c := calls.Load(); c != int64(len(in)) {
		t.Errorf("f was called %d times, want %d", c, len(in))
	}
}

func TestFanOutErrOfNothing(t *testing.T) {
	got, err := FanOutErr([]int{}, 4, func(v int) (int, error) { return v, nil })
	if err != nil {
		t.Fatalf("FanOutErr of an empty slice returned an error: %v", err)
	}
	if got == nil {
		t.Fatal("FanOutErr of an empty slice = nil results, want an empty slice")
	}
	if len(got) != 0 {
		t.Errorf("FanOutErr of an empty slice = %v, want no results", got)
	}
}
