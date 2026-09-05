package main

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestParseLineValid(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want Entry
	}{
		{"a plain record", "INFO|auth|12|login ok", Entry{"INFO", "auth", 12, "login ok"}},
		{"an error record", "ERROR|billing|1450|card declined", Entry{"ERROR", "billing", 1450, "card declined"}},
		{"zero milliseconds", "WARN|cache|0|cold start", Entry{"WARN", "cache", 0, "cold start"}},
		{"pipes in the message", "WARN|search|31|slow: SELECT | FROM docs | LIMIT 1", Entry{"WARN", "search", 31, "slow: SELECT | FROM docs | LIMIT 1"}},
		{"fields are trimmed", "  INFO | auth |  12 | login ok  ", Entry{"INFO", "auth", 12, "login ok"}},
		{"an empty message is allowed", "INFO|auth|3|", Entry{"INFO", "auth", 3, ""}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ParseLine(7, c.raw)
			if err != nil {
				t.Fatalf("ParseLine(7, %q) returned %v, want an entry", c.raw, err)
			}
			if got != c.want {
				t.Errorf("ParseLine(7, %q) = %+v, want %+v", c.raw, got, c.want)
			}
		})
	}
}

func TestParseLineFailures(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want error
	}{
		{"no separators at all", "just some text", ErrMalformed},
		{"three fields", "INFO|auth|12", ErrMalformed},
		{"empty service", "INFO||12|login ok", ErrMalformed},
		{"lowercase level", "info|auth|12|login ok", ErrBadLevel},
		{"unknown level", "DEBUG|auth|12|login ok", ErrBadLevel},
		{"duration is not a number", "INFO|auth|fast|login ok", ErrBadDuration},
		{"duration is empty", "INFO|auth||login ok", ErrBadDuration},
		{"negative duration", "INFO|auth|-3|login ok", ErrBadDuration},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			entry, err := ParseLine(4, c.raw)
			if err == nil {
				t.Fatalf("ParseLine(4, %q) = %+v, want an error", c.raw, entry)
			}
			if !errors.Is(err, c.want) {
				t.Fatalf("ParseLine(4, %q) returned %v, want an error matching %v", c.raw, err, c.want)
			}
			if entry != (Entry{}) {
				t.Errorf("ParseLine(4, %q) returned both %+v and an error", c.raw, entry)
			}
			var pe *ParseError
			if !errors.As(err, &pe) {
				t.Fatalf("ParseLine(4, %q) returned %T, want a *ParseError", c.raw, err)
			}
			if pe.Line != 4 {
				t.Errorf("ParseError.Line = %d, want the line it was given, 4", pe.Line)
			}
			if pe.Raw != c.raw {
				t.Errorf("ParseError.Raw = %q, want the line as it arrived, %q", pe.Raw, c.raw)
			}
		})
	}
}

func TestParseErrorMessage(t *testing.T) {
	err := &ParseError{Line: 4, Raw: "DEBUG|auth|10|hi", Err: ErrBadLevel}
	want := `line 4: unknown level: "DEBUG|auth|10|hi"`
	if got := err.Error(); got != want {
		t.Errorf("(*ParseError).Error() = %q, want %q", got, want)
	}
	if !errors.Is(err, ErrBadLevel) {
		t.Errorf("errors.Is(err, ErrBadLevel) = false; Unwrap must return the sentinel")
	}
	if errors.Is(err, ErrMalformed) {
		t.Errorf("errors.Is(err, ErrMalformed) = true for an ErrBadLevel failure")
	}
}

func TestStatsAggregates(t *testing.T) {
	s := NewStats()
	for _, e := range []Entry{
		{"INFO", "auth", 10, "login ok"},
		{"ERROR", "auth", 140, "token expired"},
		{"INFO", "search", 300, "query ok"},
		{"WARN", "search", 20, "slow"},
		{"INFO", "auth", 30, "logout ok"},
	} {
		s.Add(e)
	}

	got := s.Snapshot()
	if got.Total != 5 {
		t.Errorf("Total = %d, want 5", got.Total)
	}
	wantLevels := map[string]int{"INFO": 3, "ERROR": 1, "WARN": 1}
	if !reflect.DeepEqual(got.Levels, wantLevels) {
		t.Errorf("Levels = %v, want %v - only the levels that appeared", got.Levels, wantLevels)
	}
	wantServices := []ServiceStat{
		{Service: "search", Count: 2, TotalMillis: 320, SlowestMillis: 300, SlowestMessage: "query ok"},
		{Service: "auth", Count: 3, TotalMillis: 180, SlowestMillis: 140, SlowestMessage: "token expired"},
	}
	if !reflect.DeepEqual(got.Services, wantServices) {
		t.Errorf("Services = %+v, want %+v - busiest total first", got.Services, wantServices)
	}
}

func TestStatsOrderingAndTies(t *testing.T) {
	s := NewStats()
	for _, e := range []Entry{
		{"INFO", "beta", 50, "b"},
		{"INFO", "alpha", 25, "a1"},
		{"INFO", "alpha", 25, "a2"},
		{"INFO", "zulu", 40, "z"},
		{"INFO", "zulu", 10, "z2"},
	} {
		s.Add(e)
	}
	got := s.Snapshot()

	var names []string
	for _, svc := range got.Services {
		names = append(names, svc.Service)
	}
	want := []string{"alpha", "beta", "zulu"}
	if !reflect.DeepEqual(names, want) {
		t.Errorf("service order = %v, want %v - 50ms each, so the tie breaks on the name", names, want)
	}
	for _, svc := range got.Services {
		if svc.Service == "alpha" && svc.SlowestMessage != "a1" {
			t.Errorf("alpha SlowestMessage = %q, want %q - equal durations break on the smaller message", svc.SlowestMessage, "a1")
		}
	}
}

func TestStatsSnapshotDoesNotAlias(t *testing.T) {
	s := NewStats()
	s.Add(Entry{"INFO", "auth", 10, "one"})
	first := s.Snapshot()
	s.Add(Entry{"ERROR", "auth", 90, "two"})

	if first.Total != 1 {
		t.Errorf("an earlier Snapshot has Total = %d; a snapshot is a copy, not a window", first.Total)
	}
	if n := first.Levels["ERROR"]; n != 0 {
		t.Errorf("an earlier Snapshot gained an ERROR count of %d; copy the map", n)
	}
	if len(first.Services) == 1 && first.Services[0].Count != 1 {
		t.Errorf("an earlier Snapshot's ServiceStat moved to Count = %d; copy the values out", first.Services[0].Count)
	}
}

func TestStatsAddIsConcurrencySafe(t *testing.T) {
	const goroutines, each = 8, 500

	s := NewStats()
	var wg sync.WaitGroup
	for g := range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range each {
				s.Add(Entry{"INFO", fmt.Sprintf("svc%d", g%4), i % 7, "msg"})
			}
		}()
	}
	wg.Wait()

	got := s.Snapshot()
	if got.Total != goroutines*each {
		t.Errorf("Total = %d after %d concurrent Adds, want %d - guard the counters with a mutex",
			got.Total, goroutines*each, goroutines*each)
	}
	if n := got.Levels["INFO"]; n != goroutines*each {
		t.Errorf("Levels[INFO] = %d, want %d", n, goroutines*each)
	}
	sum := 0
	for _, svc := range got.Services {
		sum += svc.Count
	}
	if sum != goroutines*each {
		t.Errorf("service counts add up to %d, want %d", sum, goroutines*each)
	}
}

// collector is the simplest possible Sink: it keeps what it was given, under a
// lock, because ParseAll calls Add from several goroutines.
type collector struct {
	mu      sync.Mutex
	entries []Entry
}

func (c *collector) Add(e Entry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = append(c.entries, e)
}

func (c *collector) sorted() []Entry {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := append([]Entry(nil), c.entries...)
	sort.Slice(out, func(i, j int) bool { return out[i].Message < out[j].Message })
	return out
}

var mixedLines = []string{
	"INFO|auth|12|a login",
	"nonsense",
	"ERROR|auth|145|b token",
	"",
	"DEBUG|auth|10|c level",
	"   ",
	"WARN|search|310|d slow",
	"INFO|billing|fast|e duration",
	"INFO|search|28|f query",
}

func TestParseAllFeedsEveryGoodLine(t *testing.T) {
	c := &collector{}
	errs := ParseAll(mixedLines, 4, c)

	if len(errs) != 3 {
		t.Fatalf("ParseAll returned %d errors, want 3: %v", len(errs), errs)
	}
	want := []Entry{
		{"INFO", "auth", 12, "a login"},
		{"ERROR", "auth", 145, "b token"},
		{"WARN", "search", 310, "d slow"},
		{"INFO", "search", 28, "f query"},
	}
	if got := c.sorted(); !reflect.DeepEqual(got, want) {
		t.Errorf("entries = %+v, want %+v", got, want)
	}
}

func TestParseAllErrorsAreInLineOrder(t *testing.T) {
	// Blank lines are skipped but still count, so the failures are lines 2, 5
	// and 8. Twenty runs, because the wrong answer here is the one that only
	// shows up when the goroutines finish in a different order.
	for run := range 20 {
		errs := ParseAll(mixedLines, 4, &collector{})
		var lines []int
		for _, err := range errs {
			var pe *ParseError
			if !errors.As(err, &pe) {
				t.Fatalf("run %d: error %v is not a *ParseError", run, err)
			}
			lines = append(lines, pe.Line)
		}
		if !reflect.DeepEqual(lines, []int{2, 5, 8}) {
			t.Fatalf("run %d: failing lines = %v, want [2 5 8] every time - sort by line before returning", run, lines)
		}
	}
}

func TestParseAllNoErrorsIsEmptyNotNil(t *testing.T) {
	errs := ParseAll([]string{"INFO|auth|1|ok", ""}, 2, &collector{})
	if errs == nil {
		t.Fatalf("ParseAll returned a nil slice; return an empty one when every line parsed")
	}
	if len(errs) != 0 {
		t.Errorf("ParseAll returned %v, want no errors", errs)
	}
}

func TestParseAllHandlesTinyWorkerCounts(t *testing.T) {
	for _, workers := range []int{0, -1, 1} {
		c := &collector{}
		errs := ParseAll(mixedLines, workers, c)
		if len(errs) != 3 || len(c.sorted()) != 4 {
			t.Errorf("ParseAll(..., %d, ...) gave %d entries and %d errors, want 4 and 3 - fewer than one worker means one",
				workers, len(c.sorted()), len(errs))
		}
	}
}

// barrier is a Sink that will not let an Add return until `target` of them are
// in flight at once. A parser that works through the lines on one goroutine
// never gets there, waits out the guard, and fails this test.
type barrier struct {
	mu       sync.Mutex
	target   int
	arrived  int
	entries  int
	closed   bool
	timedOut bool
	reached  chan struct{}
}

func newBarrier(target int) *barrier {
	return &barrier{target: target, reached: make(chan struct{})}
}

func (b *barrier) release(timedOut bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.timedOut = b.timedOut || timedOut; !b.closed {
		b.closed = true
		close(b.reached)
	}
}

func (b *barrier) Add(e Entry) {
	b.mu.Lock()
	b.arrived++
	b.entries++
	arrived := b.arrived
	b.mu.Unlock()

	if arrived >= b.target {
		b.release(false)
		return
	}
	select {
	case <-b.reached:
	case <-time.After(500 * time.Millisecond):
		b.release(true)
	}
}

func TestParseAllIsConcurrent(t *testing.T) {
	lines := make([]string, 8)
	for i := range lines {
		lines[i] = fmt.Sprintf("INFO|svc%d|%d|line %d", i, i*10, i)
	}

	b := newBarrier(3)
	errs := ParseAll(lines, 4, b)

	if len(errs) != 0 {
		t.Fatalf("ParseAll returned %v, want no errors", errs)
	}
	if b.entries != len(lines) {
		t.Fatalf("sink received %d entries, want %d", b.entries, len(lines))
	}
	if b.timedOut {
		t.Errorf("no three Adds were ever in flight at once with workers = 4: parse the lines on worker goroutines rather than one after another")
	}
}

func TestSummarize(t *testing.T) {
	got, errs := Summarize(mixedLines, 4)

	if len(errs) != 3 {
		t.Fatalf("Summarize returned %d errors, want 3: %v", len(errs), errs)
	}
	if got.Total != 4 {
		t.Errorf("Total = %d, want 4", got.Total)
	}
	wantLevels := map[string]int{"INFO": 2, "WARN": 1, "ERROR": 1}
	if !reflect.DeepEqual(got.Levels, wantLevels) {
		t.Errorf("Levels = %v, want %v", got.Levels, wantLevels)
	}
	wantServices := []ServiceStat{
		{Service: "search", Count: 2, TotalMillis: 338, SlowestMillis: 310, SlowestMessage: "d slow"},
		{Service: "auth", Count: 2, TotalMillis: 157, SlowestMillis: 145, SlowestMessage: "b token"},
	}
	if !reflect.DeepEqual(got.Services, wantServices) {
		t.Errorf("Services = %+v, want %+v", got.Services, wantServices)
	}
	if !strings.Contains(errs[0].Error(), "line 2") {
		t.Errorf("first error = %q, want it to name line 2", errs[0])
	}
}
