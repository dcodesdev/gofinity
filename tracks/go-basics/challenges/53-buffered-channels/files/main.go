package main

import "fmt"

// Buffered returns a closed channel already holding every value, in order.
//
// This is the one producer that needs no goroutine: a channel with capacity
// len(values) has room for all of them, so none of the sends can block, and a
// closed channel still hands out everything left in its buffer.
//
// Buffered(nil) is a closed channel with capacity 0.
func Buffered(values []int) <-chan int {
	// TODO
	return nil
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
	// TODO
	return 0
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
	// TODO
	return nil
}

func main() {
	ch := make(chan int, 2)
	fmt.Println(FillUpTo(ch, []int{1, 2, 3}), len(ch), cap(ch))
	fmt.Println(MapLimited([]int{1, 2, 3, 4}, 2, func(v int) int { return v * v }))
}
