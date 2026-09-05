package main

import (
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// waitFor runs fn in its own goroutine and reports whether it finished, so a
// missing wg.Add is a failure with a message rather than a stuck test binary.
func waitFor[T any](t *testing.T, what string, fn func() T) T {
	t.Helper()
	var out T
	done := make(chan struct{})
	go func() {
		defer close(done)
		out = fn()
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("%s did not return within a second", what)
	}
	return out
}

func TestChunks(t *testing.T) {
	cases := []struct {
		n, workers int
		want       [][2]int
	}{
		{10, 4, [][2]int{{0, 3}, {3, 6}, {6, 8}, {8, 10}}},
		{8, 4, [][2]int{{0, 2}, {2, 4}, {4, 6}, {6, 8}}},
		{9, 2, [][2]int{{0, 5}, {5, 9}}},
		{7, 1, [][2]int{{0, 7}}},
		{5, 5, [][2]int{{0, 1}, {1, 2}, {2, 3}, {3, 4}, {4, 5}}},
		{3, 10, [][2]int{{0, 1}, {1, 2}, {2, 3}}},
		{5, 0, [][2]int{{0, 5}}},
		{5, -3, [][2]int{{0, 5}}},
	}
	for _, c := range cases {
		got := Chunks(c.n, c.workers)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("Chunks(%d, %d) = %v, want %v", c.n, c.workers, got, c.want)
		}
	}
}

func TestChunksOfNothingIsNoChunks(t *testing.T) {
	if got := Chunks(0, 4); len(got) != 0 {
		t.Errorf("Chunks(0, 4) = %v, want no chunks", got)
	}
	if got := Chunks(-2, 4); len(got) != 0 {
		t.Errorf("Chunks(-2, 4) = %v, want no chunks", got)
	}
}

func TestChunksCoverEveryIndexExactlyOnce(t *testing.T) {
	for n := 1; n <= 40; n++ {
		for w := 1; w <= 12; w++ {
			chunks := Chunks(n, w)
			seen := make([]int, n)
			prevEnd := 0
			for _, c := range chunks {
				if c[0] != prevEnd {
					t.Fatalf("Chunks(%d, %d) = %v has a gap or an overlap at %v", n, w, chunks, c)
				}
				if c[1] <= c[0] {
					t.Fatalf("Chunks(%d, %d) = %v contains an empty chunk", n, w, chunks)
				}
				for i := c[0]; i < c[1]; i++ {
					seen[i]++
				}
				prevEnd = c[1]
			}
			if prevEnd != n {
				t.Fatalf("Chunks(%d, %d) = %v stops at %d, want %d", n, w, chunks, prevEnd, n)
			}
			for i, count := range seen {
				if count != 1 {
					t.Fatalf("Chunks(%d, %d) covers index %d %d times", n, w, i, count)
				}
			}
		}
	}
}

func TestChunkSizesDifferByAtMostOne(t *testing.T) {
	for n := 1; n <= 40; n++ {
		for w := 1; w <= 12; w++ {
			chunks := Chunks(n, w)
			smallest, largest := n, 0
			for _, c := range chunks {
				size := c[1] - c[0]
				smallest = min(smallest, size)
				largest = max(largest, size)
			}
			if largest-smallest > 1 {
				t.Fatalf("Chunks(%d, %d) = %v is uneven: sizes %d..%d", n, w, chunks, smallest, largest)
			}
		}
	}
}

func TestChunksNeverMakesMoreChunksThanItems(t *testing.T) {
	if got := len(Chunks(3, 10)); got != 3 {
		t.Errorf("Chunks(3, 10) made %d chunks, want 3 - an empty chunk is a worker with nothing to do", got)
	}
}

func sequentialSum(nums []int) int {
	total := 0
	for _, v := range nums {
		total += v
	}
	return total
}

func TestParallelSumMatchesTheSequentialSum(t *testing.T) {
	nums := make([]int, 1000)
	for i := range nums {
		nums[i] = i*7 - 3
	}
	want := sequentialSum(nums)
	for _, workers := range []int{1, 2, 3, 4, 8, 16, 64, 0, -1} {
		got := waitFor(t, "ParallelSum", func() int { return ParallelSum(nums, workers) })
		if got != want {
			t.Errorf("ParallelSum(nums, %d) = %d, want %d", workers, got, want)
		}
	}
}

func TestParallelSumIsStable(t *testing.T) {
	// Repeating is the cheap stand-in for the race detector: a lost update or a
	// dropped chunk shows up as a different answer on some run, not on all.
	nums := make([]int, 500)
	for i := range nums {
		nums[i] = i + 1
	}
	want := sequentialSum(nums)
	for range 50 {
		if got := ParallelSum(nums, 8); got != want {
			t.Fatalf("ParallelSum = %d, want %d - the partials are being shared, not owned", got, want)
		}
	}
}

func TestParallelSumEdgeCases(t *testing.T) {
	if got := waitFor(t, "ParallelSum", func() int { return ParallelSum(nil, 4) }); got != 0 {
		t.Errorf("ParallelSum(nil, 4) = %d, want 0", got)
	}
	if got := waitFor(t, "ParallelSum", func() int { return ParallelSum([]int{42}, 8) }); got != 42 {
		t.Errorf("ParallelSum([]int{42}, 8) = %d, want 42", got)
	}
	if got := waitFor(t, "ParallelSum", func() int { return ParallelSum([]int{-5, 5, -7}, 2) }); got != -7 {
		t.Errorf("ParallelSum([]int{-5, 5, -7}, 2) = %d, want -7", got)
	}
}

func TestParallelCountActuallyUsesGoroutines(t *testing.T) {
	// Four chunks, and each one blocks until all four have started. A sequential
	// ParallelSum would wait forever on the first, so only a concurrent one can
	// finish this within the second waitFor allows.
	const workers = 4
	var started sync.WaitGroup
	started.Add(workers)
	release := make(chan struct{})
	go func() {
		started.Wait()
		close(release)
	}()

	nums := make([]int, workers)
	for i := range nums {
		nums[i] = i
	}
	// pred is the hook: ParallelCount calls it once per element, so with one
	// element per worker it runs once on each goroutine.
	got := waitFor(t, "ParallelCount", func() int {
		return ParallelCount(nums, func(int) bool {
			started.Done()
			<-release
			return true
		}, workers)
	})
	if got != workers {
		t.Errorf("ParallelCount = %d, want %d", got, workers)
	}
}

func TestParallelCount(t *testing.T) {
	nums := make([]int, 300)
	for i := range nums {
		nums[i] = i
	}
	even := func(v int) bool { return v%2 == 0 }
	for _, workers := range []int{1, 3, 7, 32} {
		got := waitFor(t, "ParallelCount", func() int { return ParallelCount(nums, even, workers) })
		if got != 150 {
			t.Errorf("ParallelCount(0..299, even, %d) = %d, want 150", workers, got)
		}
	}
}

func TestParallelCountCallsPredOncePerElement(t *testing.T) {
	nums := make([]int, 257)
	var calls atomic.Int64
	got := waitFor(t, "ParallelCount", func() int {
		return ParallelCount(nums, func(int) bool {
			calls.Add(1)
			return false
		}, 6)
	})
	if got != 0 {
		t.Errorf("ParallelCount = %d, want 0", got)
	}
	if c := calls.Load(); c != int64(len(nums)) {
		t.Errorf("pred was called %d times, want %d - every element belongs to exactly one chunk", c, len(nums))
	}
}

func TestParallelCountOfNothing(t *testing.T) {
	got := waitFor(t, "ParallelCount", func() int {
		return ParallelCount(nil, func(int) bool { return true }, 4)
	})
	if got != 0 {
		t.Errorf("ParallelCount(nil, ...) = %d, want 0", got)
	}
}
