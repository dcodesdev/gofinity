package main

import (
	"context"
	"fmt"
	"time"
)

// Budget reports how long ctx has left and whether it has a deadline at all.
//
// A context with no deadline reports (0, false). A context whose deadline has
// already passed reports a duration that is zero or negative, not an error -
// the caller decides whether a tiny budget is worth starting work for.
func Budget(ctx context.Context) (time.Duration, bool) {
	// TODO
	return 0, false
}

// Limit returns a child of ctx that finishes at most d from now.
//
// "At most" is the whole point: a child can always be stricter than its
// parent and can never be laxer. If ctx already ends sooner than d, the child
// ends when ctx does.
//
// The caller is responsible for calling the returned cancel function.
func Limit(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	// TODO
	return ctx, func() {}
}

// Do runs work and returns its result, unless ctx finishes first - in which
// case it returns ctx.Err() and does not wait for work to end.
//
// work is somebody else's function: it knows nothing about ctx and cannot be
// interrupted. Do returns early, so the goroutine running work must be able to
// finish and exit whether or not anybody is still listening.
func Do(ctx context.Context, work func() (int, error)) (int, error) {
	// TODO
	return 0, nil
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
	// TODO
	return 0, nil
}

// Classify names what an error means to a caller: "ok" for nil, "timeout" for
// a deadline that passed, "cancelled" for a cancellation, and "failed" for
// anything else.
//
// It must see through wrapping, so compare with errors.Is rather than ==.
func Classify(err error) string {
	// TODO
	return ""
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
