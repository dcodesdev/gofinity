package main

// Sentinels framing the result object on stdout. The runner captures the
// subprocess's own stdout, so in practice nothing else is printed - but a Go
// panic, a `go` toolchain warning, or a future change could still land on our
// stdout, and the consumer must not have to guess which line is the result.
//
// The contract: scan for the LAST BeginSentinel line, take every line up to
// the next EndSentinel line, and parse that as JSON.
const (
	BeginSentinel = "<<<GOFINITY:RESULT>>>"
	EndSentinel   = "<<<GOFINITY:END>>>"
)

// Statuses a single test can end in.
const (
	StatusPassed  = "passed"
	StatusFailed  = "failed"
	StatusSkipped = "skipped"
)

// TestResult is one test (or subtest) from the run. Subtests appear as their
// own entries, named the way `go test` names them: `TestGreet/a_name`.
type TestResult struct {
	Name      string `json:"name"`
	Package   string `json:"package"`
	Status    string `json:"status"`
	Output    string `json:"output"`
	ElapsedMs int64  `json:"elapsedMs"`
}

// Result is the single JSON object the runner prints. It is the entire
// contract between the container and the API; see README.md.
type Result struct {
	// OK is true only when the command exited zero, nothing timed out, and no
	// test failed. It is the one field a caller needs for a pass/fail decision.
	OK      bool `json:"ok"`
	Passed  int  `json:"passed"`
	Failed  int  `json:"failed"`
	Skipped int  `json:"skipped"`

	Tests []TestResult `json:"tests"`

	// Stdout is the human-readable reconstruction of what `go test` printed:
	// the `Output` of every JSON event in order, plus any line that was not
	// JSON at all (which is where older toolchains put build errors).
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`

	ExitCode   int   `json:"exitCode"`
	TimedOut   bool  `json:"timedOut"`
	DurationMs int64 `json:"durationMs"`

	// Error is set only for runner-level failures - a malformed payload, a
	// workspace that could not be written, a `go` binary that is not there.
	// A failing test is not an error; it is Failed > 0.
	Error string `json:"error,omitempty"`
}

// errorResult builds the result printed when the runner itself could not do
// its job. Tests is non-nil so the consumer never has to null-check it.
func errorResult(err error) *Result {
	return &Result{
		OK:       false,
		Tests:    []TestResult{},
		ExitCode: -1,
		Error:    err.Error(),
	}
}
