package main

import (
	"fmt"
	"sync"
)

// Buffered returns a closed channel already holding every value, in order.
//
// This is the one producer that needs no goroutine: a channel with capacity
// len(values) has room for all of them, so none of the sends can block, and a
// closed channel still hands out everything left in its buffer.
//
// Buffered(nil) is a closed channel with capacity 0.
func Buffered(values []int) <-chan int {
	ch := make(chan int, len(values))
	for _, v := range values {
		ch <- v
	}
	// Closing says "no more will arrive". It does not discard the buffer, so
	// the caller still receives every value before the channel reports closed.
	close(ch)
	return ch
}

// FillUpTo sends values into ch until ch is full or values runs out, and
// returns how many it sent.
//
// len(ch) is the number of values waiting in the buffer and cap(ch) is the
// size it was made with, so len(ch) < cap(ch) means "there is room". That is
// only trustworthy while you are the sole sender, which here you are.
//
// An unbuffered ch has cap 0 and is therefore always full: FillUpTo sends
// nothing and returns 0.
func FillUpTo(ch chan int, values []int) int {
	sent := 0
	for _, v := range values {
		if len(ch) >= cap(ch) {
			break
		}
		ch <- v
		sent++
	}
	return sent
}

// MapLimited applies f to every element of in and returns the results in the
// input order, running at most limit calls at the same time.
//
// The shape to learn - a buffered channel used as a semaphore:
//
//	sem := make(chan struct{}, limit)
//	sem <- struct{}{}        // take a permit, blocking while all are out
//	defer func() { <-sem }() // give it back, whatever happens
//
// Every goroutine starts immediately; only limit of them are ever past the
// send. A limit below 1 is treated as 1, and an empty input returns an empty,
// non-nil slice.
func MapLimited(in []int, limit int, f func(int) int) []int {
	if limit < 1 {
		limit = 1
	}
	// make, not append: every index exists before any goroutine starts, so no
	// two goroutines ever touch the same memory.
	out := make([]int, len(in))
	// struct{} carries no information and occupies no memory. The capacity is
	// the whole point: it is the number of permits.
	sem := make(chan struct{}, limit)
	var wg sync.WaitGroup
	for i, v := range in {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			// The permit goes back even if f panics, which is the same reason
			// Done is deferred.
			defer func() { <-sem }()
			out[i] = f(v)
		}()
	}
	wg.Wait()
	return out
}

func main() {
	ch := make(chan int, 2)
	fmt.Println(FillUpTo(ch, []int{1, 2, 3}), len(ch), cap(ch))
	fmt.Println(MapLimited([]int{1, 2, 3, 4}, 2, func(v int) int { return v * v }))
}
