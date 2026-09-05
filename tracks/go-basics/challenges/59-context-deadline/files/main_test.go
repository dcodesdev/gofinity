package main

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"testing"
	"time"
)

// guard is how long any one step may take. Nothing here waits for a deadline
// to arrive - the expired contexts are already expired when they are made - so
// a step still running after this long is stuck rather than slow.
const guard = 500 * time.Millisecond

// expired returns a context whose deadline passed a second ago, so no test has
// to spend real time waiting for one.
func expired(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	t.Cleanup(cancel)
	return ctx
}

// doWithin calls Do in its own goroutine, so an implementation that waits for
// work no matter what fails the test instead of hanging the run.
func doWithin(t *testing.T, ctx context.Context, work func() (int, error)) (int, error) {
	t.Helper()
	type result struct {
		value int
		err   error
	}
	done := make(chan result, 1)
	go func() {
		v, err := Do(ctx, work)
		done <- result{v, err}
	}()
	select {
	case r := <-done:
		return r.value, r.err
	case <-time.After(guard):
		t.Fatal("Do did not return - it must select on ctx.Done() as well as on work finishing")
		return 0, nil
	}
}

func TestBudgetReportsNoDeadline(t *testing.T) {
	left, ok := Budget(context.Background())
	if ok {
		t.Errorf("Budget(Background()) reported a deadline of %v, want none", left)
	}
	if left != 0 {
		t.Errorf("Budget(Background()) = %v, want 0 when there is no deadline", left)
	}
}

func TestBudgetReportsTimeRemaining(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	left, ok := Budget(ctx)
	if !ok {
		t.Fatal("Budget on a context with a one-minute timeout reported no deadline")
	}
	if left <= 0 || left > time.Minute {
		t.Errorf("Budget = %v, want something above 0 and no more than a minute", left)
	}
}

func TestBudgetIsNegativeOnceTheDeadlineHasPassed(t *testing.T) {
	left, ok := Budget(expired(t))
	if !ok {
		t.Fatal("Budget on an expired context reported no deadline, want one that has passed")
	}
	if left > 0 {
		t.Errorf("Budget = %v for a deadline a second in the past, want zero or less", left)
	}
}

func TestLimitAppliesItsOwnTimeout(t *testing.T) {
	ctx, cancel := Limit(context.Background(), time.Minute)
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("Limit(Background(), time.Minute) returned a context with no deadline")
	}
	if left := time.Until(deadline); left <= 0 || left > time.Minute {
		t.Errorf("Limit left %v on the clock, want something up to a minute", left)
	}
}

func TestLimitCannotOutliveItsParent(t *testing.T) {
	parent, cancelParent := context.WithTimeout(context.Background(), time.Second)
	defer cancelParent()
	parentDeadline, _ := parent.Deadline()

	ctx, cancel := Limit(parent, time.Hour)
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("Limit returned a context with no deadline, want its parent's")
	}
	if !deadline.Equal(parentDeadline) {
		t.Errorf("Limit(parent ending in 1s, 1h) has deadline %v, want the parent's %v - a child may be stricter, never laxer",
			deadline, parentDeadline)
	}
}

func TestLimitOfAnExpiredParentIsAlreadyDone(t *testing.T) {
	ctx, cancel := Limit(expired(t), time.Hour)
	defer cancel()

	select {
	case <-ctx.Done():
		if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Errorf("ctx.Err() = %v, want context.DeadlineExceeded", ctx.Err())
		}
	case <-time.After(guard):
		t.Error("Limit of a context whose deadline has passed is still live, want it done straight away")
	}
}

func TestLimitCancelStopsTheChildOnly(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	defer cancelParent()

	ctx, cancel := Limit(parent, time.Hour)
	cancel()

	select {
	case <-ctx.Done():
	case <-time.After(guard):
		t.Fatal("the cancel function Limit returned did not stop the context it returned")
	}
	if parent.Err() != nil {
		t.Errorf("cancelling the child also finished the parent: parent.Err() = %v, want nil", parent.Err())
	}
}

func TestDoReturnsTheWorkResult(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	got, err := doWithin(t, ctx, func() (int, error) { return 42, nil })
	if got != 42 || err != nil {
		t.Errorf("Do = (%d, %v), want (42, <nil>)", got, err)
	}
}

func TestDoReturnsTheWorkError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	fail := errors.New("boom")
	got, err := doWithin(t, ctx, func() (int, error) { return 7, fail })
	if !errors.Is(err, fail) {
		t.Errorf("Do returned %v, want work's own error - pass it through untouched", err)
	}
	if got != 7 {
		t.Errorf("Do returned value %d, want work's own 7", got)
	}
}

func TestDoGivesUpOnAnExpiredContext(t *testing.T) {
	release := make(chan struct{})
	defer close(release)

	// work never finishes on its own. The only way out is the context.
	got, err := doWithin(t, expired(t), func() (int, error) {
		<-release
		return 1, nil
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Do on an expired context = %v, want context.DeadlineExceeded", err)
	}
	if got != 0 {
		t.Errorf("Do returned value %d alongside its error, want 0", got)
	}
}

func TestDoLeavesNoGoroutineBehind(t *testing.T) {
	release := make(chan struct{})
	before := runtime.NumGoroutine()

	if _, err := doWithin(t, expired(t), func() (int, error) { <-release; return 1, nil }); err == nil {
		t.Fatal("Do on an expired context returned no error")
	}

	// Do has already returned. When work finally finishes, its goroutine must
	// be able to hand over the result and exit - which it can only do if the
	// channel it sends on has room for a value nobody will ever read.
	close(release)
	for until := time.Now().Add(guard); time.Now().Before(until); {
		if runtime.NumGoroutine() <= before {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Error("the goroutine running work never exited - it is parked on a send nobody will receive, so give that channel a buffer of one")
}

func TestRetryStopsAtTheFirstSuccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var seen []int
	n, err := Retry(ctx, func(n int) error {
		seen = append(seen, n)
		if n < 3 {
			return fmt.Errorf("attempt %d failed", n)
		}
		return nil
	})
	if err != nil {
		t.Errorf("Retry = %v, want nil once an attempt succeeds", err)
	}
	if n != 3 {
		t.Errorf("Retry made %d attempts, want 3", n)
	}
	if len(seen) != 3 || seen[0] != 1 || seen[1] != 2 || seen[2] != 3 {
		t.Errorf("attempt was called with %v, want 1, 2, 3 - the count starts at one", seen)
	}
}

func TestRetrySucceedsFirstTime(t *testing.T) {
	n, err := Retry(context.Background(), func(int) error { return nil })
	if n != 1 || err != nil {
		t.Errorf("Retry with an attempt that works = (%d, %v), want (1, <nil>)", n, err)
	}
}

func TestRetryMakesNoAttemptOnAnExpiredContext(t *testing.T) {
	calls := 0
	n, err := Retry(expired(t), func(int) error {
		calls++
		return errors.New("should never run")
	})
	if calls != 0 {
		t.Errorf("attempt ran %d times on an already-expired context, want 0 - check ctx before attempting", calls)
	}
	if n != 0 {
		t.Errorf("Retry reported %d attempts, want 0", n)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Retry = %v, want an error carrying context.DeadlineExceeded", err)
	}
}

func TestRetryStopsWhenTheContextIsCancelledAndKeepsBothErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	last := errors.New("the database is asleep")
	n, err := Retry(ctx, func(n int) error {
		if n == 2 {
			cancel()
		}
		return fmt.Errorf("attempt %d: %w", n, last)
	})
	if n != 2 {
		t.Errorf("Retry made %d attempts, want 2 - the context was cancelled during the second", n)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Retry = %v, want an error carrying context.Canceled", err)
	}
	if !errors.Is(err, last) {
		t.Errorf("Retry = %v, want it to carry the last attempt's error too - errors.Join keeps both", err)
	}
}

func TestClassifyNamesTheOutcome(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"nil", nil, "ok"},
		{"deadline", context.DeadlineExceeded, "timeout"},
		{"cancel", context.Canceled, "cancelled"},
		{"wrapped deadline", fmt.Errorf("fetching user: %w", context.DeadlineExceeded), "timeout"},
		{"wrapped cancel", fmt.Errorf("fetching user: %w", context.Canceled), "cancelled"},
		{"joined", errors.Join(errors.New("no route to host"), context.Canceled), "cancelled"},
		{"other", errors.New("no route to host"), "failed"},
	}
	for _, c := range cases {
		if got := Classify(c.err); got != c.want {
			t.Errorf("Classify(%v) = %q, want %q", c.err, got, c.want)
		}
	}
}
