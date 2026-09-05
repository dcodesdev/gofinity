package main

import (
	"fmt"
	"sync"
)

// Chunks splits the index range [0, n) into contiguous, half-open [start, end)
// ranges, as evenly as possible, in order.
//
//	Chunks(10, 4) -> [[0 3] [3 6] [6 8] [8 10]]
//	Chunks(3, 10) -> [[0 1] [1 2] [2 3]]
//	Chunks(5, 0)  -> [[0 5]]
//	Chunks(0, 4)  -> no chunks
//
// The rules:
//
//   - every index in [0, n) appears in exactly one chunk,
//   - no two chunks differ in size by more than one, so the first n%workers
//     chunks get one extra element,
//   - no chunk is empty, so more workers than items means fewer chunks,
//   - workers < 1 is treated as 1.
//
// It starts no goroutines. Range arithmetic is where off-by-ones live, and a
// pure function can be tested without concurrency in the way.
func Chunks(n, workers int) [][2]int {
	if n <= 0 {
		return nil
	}
	if workers < 1 {
		workers = 1
	}
	// Never an empty chunk: with more workers than items, each item is its own
	// chunk and the spare workers get nothing to do.
	if workers > n {
		workers = n
	}

	size, extra := n/workers, n%workers
	chunks := make([][2]int, 0, workers)
	start := 0
	for i := range workers {
		end := start + size
		// The first `extra` chunks absorb the remainder, one element each, so
		// no two chunk sizes differ by more than one.
		if i < extra {
			end++
		}
		chunks = append(chunks, [2]int{start, end})
		start = end
	}
	return chunks
}

// ParallelSum returns the sum of nums, computed by `workers` goroutines.
//
// Split with Chunks, give each goroutine one cell of a partials slice to write,
// Wait, then add the partials up here. Do not have the goroutines add into a
// shared total: `total += v` is a read, an add and a write, and two goroutines
// interleaving those lose updates.
//
// The answer must equal the sequential sum exactly, every time.
func ParallelSum(nums []int, workers int) int {
	chunks := Chunks(len(nums), workers)
	// One cell per goroutine. Nothing is shared, so nothing needs guarding.
	partials := make([]int, len(chunks))

	var wg sync.WaitGroup
	for i, c := range chunks {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sum := 0
			// Reslicing does not copy. Every goroutine reads the same backing
			// array, which is fine: only writes need care.
			for _, v := range nums[c[0]:c[1]] {
				sum += v
			}
			partials[i] = sum
		}()
	}
	wg.Wait()

	// Combine, back on one goroutine, after every write has happened.
	total := 0
	for _, p := range partials {
		total += p
	}
	return total
}

// ParallelCount returns how many elements of nums satisfy pred, computed by
// `workers` goroutines.
//
// The same six lines as ParallelSum with a different inner loop, which is the
// point: split, work, combine is one shape, not one function.
//
// pred is called exactly once per element.
func ParallelCount(nums []int, pred func(int) bool, workers int) int {
	chunks := Chunks(len(nums), workers)
	partials := make([]int, len(chunks))

	var wg sync.WaitGroup
	for i, c := range chunks {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count := 0
			for _, v := range nums[c[0]:c[1]] {
				if pred(v) {
					count++
				}
			}
			partials[i] = count
		}()
	}
	wg.Wait()

	total := 0
	for _, p := range partials {
		total += p
	}
	return total
}

func main() {
	nums := make([]int, 100)
	for i := range nums {
		nums[i] = i
	}
	fmt.Println(Chunks(10, 4))
	fmt.Println(ParallelSum(nums, 4))
	fmt.Println(ParallelCount(nums, func(v int) bool { return v%2 == 0 }, 4))
}
