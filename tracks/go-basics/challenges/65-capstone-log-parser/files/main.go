package main

import (
	"errors"
	"fmt"
)

// The log format. One record per line, four fields, pipe separated:
//
//	LEVEL|service|duration_ms|message
//	ERROR|auth|145|token expired for user 91
//
// LEVEL is exactly one of INFO, WARN or ERROR, in capitals. service is a
// non-empty name. duration_ms is a whole number of milliseconds, zero or
// more. message is the rest of the line and may itself contain pipes.
//
// Every field is trimmed of surrounding spaces. A blank line is not a record
// and not an error: it is skipped.

// The three sentinels a caller can react to. They say what was wrong with a
// line, and they are compared with errors.Is, never by their text.
var (
	ErrMalformed   = errors.New("malformed line")
	ErrBadLevel    = errors.New("unknown level")
	ErrBadDuration = errors.New("bad duration")
)

// Entry is one parsed record.
type Entry struct {
	Level   string
	Service string
	Millis  int
	Message string
}

// ParseError says which line failed and why. It carries the raw line so the
// message can quote it, and it wraps the sentinel so errors.Is still reaches
// through it.
type ParseError struct {
	Line int    // 1-based index into the input
	Raw  string // the line exactly as it arrived
	Err  error  // one of the sentinels above
}

// Error renders the failure in exactly this form:
//
//	fmt.Sprintf("line %d: %v: %q", e.Line, e.Err, e.Raw)
//
// giving, for example:
//
//	line 4: unknown level: "DEBUG|auth|10|hi"
func (e *ParseError) Error() string {
	// TODO
	return ""
}

// Unwrap exposes the sentinel, which is what makes errors.Is(err, ErrBadLevel)
// true for a *ParseError carrying it.
func (e *ParseError) Unwrap() error {
	// TODO
	return nil
}

// ParseLine parses one record. line is its 1-based number in the input, used
// only for reporting.
//
// Every failure is a *ParseError with the right sentinel inside:
//
//	fewer than four fields, or an empty service  ->  ErrMalformed
//	a level that is not INFO, WARN or ERROR      ->  ErrBadLevel
//	a duration that is not a number, or negative ->  ErrBadDuration
//
// The message is the fourth field and everything after it, so a message
// containing a pipe survives intact.
func ParseLine(line int, raw string) (Entry, error) {
	// TODO
	return Entry{}, nil
}

// Sink is what a parser hands finished entries to. Anything that aggregates
// entries can be one, which is how ParseAll stays ignorant of what is being
// counted.
//
// ParseAll calls Add from several goroutines at once, so an implementation
// must be safe for concurrent use. That requirement belongs in the interface's
// documentation, because it cannot be expressed in its signature.
type Sink interface {
	Add(e Entry)
}

// ServiceStat is what one service did across the whole log.
type ServiceStat struct {
	Service        string
	Count          int
	TotalMillis    int
	SlowestMillis  int
	SlowestMessage string
}

// Summary is the aggregate, as a plain value: nothing in it is shared with the
// Stats it came from, so a caller can read it without holding a lock.
type Summary struct {
	Total    int            // records aggregated
	Levels   map[string]int // count per level, only the levels that appeared
	Services []ServiceStat  // slowest first: TotalMillis descending, then Service ascending
}

// Stats is the Sink this program uses. Add the fields it needs - the counters,
// and whatever makes it safe to call Add from several goroutines at once.
type Stats struct {
	// TODO
}

// NewStats returns an empty Stats, ready to use.
func NewStats() *Stats {
	// TODO
	return &Stats{}
}

// Add records one entry. It is called concurrently.
func (s *Stats) Add(e Entry) {
	// TODO
}

// Snapshot returns the totals so far as a Summary.
//
// Services is sorted by TotalMillis descending, then by Service ascending, so
// two runs over the same log produce the same report. SlowestMillis is the
// largest Millis seen for that service and SlowestMessage its message; when
// two entries tie on Millis the lexicographically smaller message wins, so the
// answer does not depend on which goroutine got there first.
//
// The returned map and slice are the caller's: later Adds must not change
// them.
func (s *Stats) Snapshot() Summary {
	// TODO
	return Summary{}
}

// ParseAll parses every line with workers goroutines and feeds the entries to
// sink. workers below 1 counts as 1.
//
// Blank lines - empty, or only spaces - are skipped without an entry and
// without an error, but they still count towards the line numbers.
//
// The returned errors are ordered by line number, ascending, whatever order
// the goroutines happened to finish in: a parser whose report reshuffles
// itself between runs is unusable. The line number lives inside the
// *ParseError, so getting it back out is errors.As.
//
// It returns an empty slice, not nil, when every line parsed.
func ParseAll(lines []string, workers int, sink Sink) []error {
	// TODO
	return nil
}

// Summarize is the whole program: parse the lines with workers goroutines into
// a fresh Stats, and return its Summary alongside the ordered failures.
func Summarize(lines []string, workers int) (Summary, []error) {
	// TODO
	return Summary{}, nil
}

func main() {
	lines := []string{
		"INFO|auth|12|login ok",
		"ERROR|auth|145|token expired",
		"",
		"WARN|search|310|slow query: SELECT | FROM docs",
		"DEBUG|auth|10|not a level",
		"INFO|search|28|query ok",
		"INFO|billing|no|charge",
	}

	summary, errs := Summarize(lines, 4)
	fmt.Printf("%d records, %d bad lines\n", summary.Total, len(errs))
	for _, level := range []string{"INFO", "WARN", "ERROR"} {
		if n := summary.Levels[level]; n > 0 {
			fmt.Printf("  %-5s %d\n", level, n)
		}
	}
	for _, s := range summary.Services {
		fmt.Printf("  %-8s %3d records %6dms total, slowest %dms %q\n",
			s.Service, s.Count, s.TotalMillis, s.SlowestMillis, s.SlowestMessage)
	}
	for _, err := range errs {
		fmt.Println("  ", err)
	}
}
