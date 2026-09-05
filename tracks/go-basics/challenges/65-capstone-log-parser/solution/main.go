package main

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
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
	return fmt.Sprintf("line %d: %v: %q", e.Line, e.Err, e.Raw)
}

// Unwrap exposes the sentinel, which is what makes errors.Is(err, ErrBadLevel)
// true for a *ParseError carrying it.
func (e *ParseError) Unwrap() error {
	return e.Err
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
	fail := func(err error) (Entry, error) {
		return Entry{}, &ParseError{Line: line, Raw: raw, Err: err}
	}

	fields := strings.SplitN(strings.TrimSpace(raw), "|", 4)
	if len(fields) < 4 {
		return fail(ErrMalformed)
	}

	level := strings.TrimSpace(fields[0])
	switch level {
	case "INFO", "WARN", "ERROR":
	default:
		return fail(ErrBadLevel)
	}

	service := strings.TrimSpace(fields[1])
	if service == "" {
		return fail(ErrMalformed)
	}

	millis, err := strconv.Atoi(strings.TrimSpace(fields[2]))
	if err != nil || millis < 0 {
		return fail(ErrBadDuration)
	}

	return Entry{
		Level:   level,
		Service: service,
		Millis:  millis,
		Message: strings.TrimSpace(fields[3]),
	}, nil
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
	mu       sync.Mutex
	total    int
	levels   map[string]int
	services map[string]*ServiceStat
}

// NewStats returns an empty Stats, ready to use.
func NewStats() *Stats {
	return &Stats{
		levels:   make(map[string]int),
		services: make(map[string]*ServiceStat),
	}
}

// Add records one entry. It is called concurrently.
func (s *Stats) Add(e Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.total++
	s.levels[e.Level]++

	stat := s.services[e.Service]
	if stat == nil {
		stat = &ServiceStat{Service: e.Service}
		s.services[e.Service] = stat
	}
	stat.Count++
	stat.TotalMillis += e.Millis
	if e.Millis > stat.SlowestMillis || (e.Millis == stat.SlowestMillis && (stat.Count == 1 || e.Message < stat.SlowestMessage)) {
		stat.SlowestMillis = e.Millis
		stat.SlowestMessage = e.Message
	}
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
	s.mu.Lock()
	defer s.mu.Unlock()

	sum := Summary{
		Total:    s.total,
		Levels:   make(map[string]int, len(s.levels)),
		Services: make([]ServiceStat, 0, len(s.services)),
	}
	for level, n := range s.levels {
		sum.Levels[level] = n
	}
	for _, stat := range s.services {
		sum.Services = append(sum.Services, *stat)
	}
	sort.Slice(sum.Services, func(i, j int) bool {
		if sum.Services[i].TotalMillis != sum.Services[j].TotalMillis {
			return sum.Services[i].TotalMillis > sum.Services[j].TotalMillis
		}
		return sum.Services[i].Service < sum.Services[j].Service
	})
	return sum
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
	if workers < 1 {
		workers = 1
	}

	type job struct {
		line int
		raw  string
	}

	jobs := make(chan job)
	errs := []error{}
	var mu sync.Mutex
	var wg sync.WaitGroup

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				entry, err := ParseLine(j.line, j.raw)
				if err != nil {
					mu.Lock()
					errs = append(errs, err)
					mu.Unlock()
					continue
				}
				sink.Add(entry)
			}
		}()
	}

	for i, raw := range lines {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		jobs <- job{line: i + 1, raw: raw}
	}
	close(jobs)
	wg.Wait()

	lineOf := func(err error) int {
		var pe *ParseError
		if errors.As(err, &pe) {
			return pe.Line
		}
		return 0
	}
	sort.Slice(errs, func(i, j int) bool { return lineOf(errs[i]) < lineOf(errs[j]) })
	return errs
}

// Summarize is the whole program: parse the lines with workers goroutines into
// a fresh Stats, and return its Summary alongside the ordered failures.
func Summarize(lines []string, workers int) (Summary, []error) {
	stats := NewStats()
	errs := ParseAll(lines, workers, stats)
	return stats.Snapshot(), errs
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
