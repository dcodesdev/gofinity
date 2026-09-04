package main

import (
	"bufio"
	"encoding/json"
	"math"
	"strings"
)

// testEvent is one line of `go test -json`. Only the fields the runner uses are
// declared; the stream carries more and is allowed to grow.
type testEvent struct {
	Action  string   `json:"Action"`
	Package string   `json:"Package"`
	Test    string   `json:"Test"`
	Output  string   `json:"Output"`
	Elapsed *float64 `json:"Elapsed"`
}

// ParsedRun is everything the parser can tell from the stream alone. Whether
// the run is `ok` also depends on the exit code and the timeout, which the
// parser does not see.
type ParsedRun struct {
	Tests   []TestResult
	Passed  int
	Failed  int
	Skipped int
	// Stdout is the reconstructed human-readable output.
	Stdout string
	// BuildFailed is true when the toolchain reported a build failure, which is
	// a compile error in the user's code — a distinct thing from a failing test
	// and worth telling them apart in the UI.
	BuildFailed bool
}

// maxScanLine is generous: a single `go test -json` event carries one line of
// output, but a Go panic's stack trace can arrive as one very long line.
const maxScanLine = 4 * 1024 * 1024

// ParseGoTestJSON turns a `go test -json` stream into a structured run.
//
// Lines that are not JSON are not an error: a compile error from an older
// toolchain, a `go: downloading …` notice, or anything the toolchain prints
// before the stream starts all arrive as plain text. They are preserved in
// Stdout in the order they appeared, so nothing the user needs to read is lost.
func ParseGoTestJSON(stream string) *ParsedRun {
	run := &ParsedRun{Tests: []TestResult{}}

	var out strings.Builder
	// index into run.Tests, keyed by package+test, so entries keep the order
	// they first appeared in rather than a map's random one.
	index := make(map[string]int)
	sawBuildFailed := false

	scanner := bufio.NewScanner(strings.NewReader(stream))
	scanner.Buffer(make([]byte, 0, 64*1024), maxScanLine)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var ev testEvent
		if !strings.HasPrefix(line, "{") || json.Unmarshal([]byte(line), &ev) != nil {
			// Not an event: plain toolchain output. Keep it verbatim.
			out.WriteString(line)
			out.WriteString("\n")
			continue
		}

		if ev.Output != "" {
			out.WriteString(ev.Output)
			// Toolchains that do not emit `build-fail` mark it in the text.
			if strings.Contains(ev.Output, "[build failed]") {
				sawBuildFailed = true
			}
		}

		if ev.Action == "build-fail" {
			run.BuildFailed = true
			continue
		}
		// Package-level events carry no test name. `fail` at package level with
		// no failing test inside it is what a build failure looks like on
		// toolchains that do not emit `build-fail`.
		if ev.Test == "" {
			if ev.Action == "fail" && sawBuildFailed {
				run.BuildFailed = true
			}
			continue
		}

		key := ev.Package + "\x00" + ev.Test
		i, ok := index[key]
		if !ok {
			run.Tests = append(run.Tests, TestResult{
				Name:    ev.Test,
				Package: ev.Package,
				Status:  StatusFailed,
			})
			i = len(run.Tests) - 1
			index[key] = i
		}
		t := &run.Tests[i]

		switch ev.Action {
		case "output":
			t.Output += ev.Output
		case "pass":
			t.Status = StatusPassed
		case "fail":
			t.Status = StatusFailed
		case "skip":
			t.Status = StatusSkipped
		}
		if ev.Elapsed != nil {
			t.ElapsedMs = int64(math.Round(*ev.Elapsed * 1000))
		}
	}

	// A truncated stream (the process was killed mid-test) leaves entries that
	// never got a terminal action. They stay `failed`, which is the honest
	// reading: the test did not pass.
	for _, t := range run.Tests {
		switch t.Status {
		case StatusPassed:
			run.Passed++
		case StatusSkipped:
			run.Skipped++
		default:
			run.Failed++
		}
	}

	run.Stdout = out.String()
	return run
}
