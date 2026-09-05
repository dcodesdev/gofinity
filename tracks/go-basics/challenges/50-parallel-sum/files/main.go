package main

import "fmt"

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
	// TODO
	return nil
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
	// TODO
	return 0
}

// ParallelCount returns how many elements of nums satisfy pred, computed by
// `workers` goroutines.
//
// The same six lines as ParallelSum with a different inner loop, which is the
// point: split, work, combine is one shape, not one function.
//
// pred is called exactly once per element.
func ParallelCount(nums []int, pred func(int) bool, workers int) int {
	// TODO
	return 0
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
