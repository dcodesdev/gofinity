package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Budget reports how long ctx has left and whether it has a deadline at all.
//
// A context with no deadline reports (0, false). A context whose deadline has
// already passed reports a duration that is zero or negative, not an error -
// the caller decides whether a tiny budget is worth starting work for.
func Budget(ctx context.Context) (time.Duration, bool) {
	deadline, ok := ctx.Deadline()
	if !ok {
		return 0, false
	}
	// time.Until is now-relative and may be negative, which is the honest
	// answer for a deadline that has already passed.
	return time.Until(deadline), true
}

// Limit returns a child of ctx that finishes at most d from now.
//
// "At most" is the whole point: a child can always be stricter than its
// parent and can never be laxer. If ctx already ends sooner than d, the child
// ends when ctx does.
//
// The caller is responsible for calling the returned cancel function.
func Limit(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	// There is nothing to compare by hand: WithTimeout derives a child, and a
	// derived context can only ever be stricter than its parent. If ctx ends
	// sooner, the child ends with it.
	return context.WithTimeout(ctx, d)
}

// Do runs work and returns its result, unless ctx finishes first - in which
// case it returns ctx.Err() and does not wait for work to end.
//
// work is somebody else's function: it knows nothing about ctx and cannot be
// interrupted. Do returns early, so the goroutine running work must be able to
// finish and exit whether or not anybody is still listening.
func Do(ctx context.Context, work func() (int, error)) (int, error) {
	type result struct {
		value int
		err   error
	}
	// One slot of buffer, so the goroutine can deliver its result and exit
	// even when nobody is listening any more. On an unbuffered channel that
	// send would block for ever and the goroutine would leak.
	done := make(chan result, 1)
	go func() {
		value, err := work()
		done <- result{value, err}
	}()

	select {
	case r := <-done:
		return r.value, r.err
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

// Retry calls attempt with the attempt number - 1, then 2, and so on - until
// it returns nil, and returns how many attempts it made.
//
// It stops when ctx finishes. It also checks ctx *before* the first attempt,
// so an expired context costs zero attempts.
//
// When it gives up, the error must carry both halves of the story: the last
// error attempt returned, and the context's own error. errors.Is has to find
// each of them.
func Retry(ctx context.Context, attempt func(n int) error) (int, error) {
	var last error
	for n := 1; ; n++ {
		if err := ctx.Err(); err != nil {
			// errors.Join drops nils, so an expired context that never got to
			// attempt anything yields just the context error.
			return n - 1, errors.Join(last, err)
		}
		last = attempt(n)
		if last == nil {
			return n, nil
		}
	}
}

// Classify names what an error means to a caller: "ok" for nil, "timeout" for
// a deadline that passed, "cancelled" for a cancellation, and "failed" for
// anything else.
//
// It must see through wrapping, so compare with errors.Is rather than ==.
func Classify(err error) string {
	switch {
	case err == nil:
		return "ok"
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	case errors.Is(err, context.Canceled):
		return "cancelled"
	default:
		return "failed"
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	left, ok := Budget(ctx)
	fmt.Println(ok, left > 0)

	fmt.Println(Do(ctx, func() (int, error) { return 42, nil }))

	n, err := Retry(ctx, func(n int) error {
		if n < 3 {
			return fmt.Errorf("attempt %d: not yet", n)
		}
		return nil
	})
	fmt.Println(n, err, Classify(err))
}
