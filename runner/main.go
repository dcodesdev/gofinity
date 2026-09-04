package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

// Environment the entrypoint reads. Everything else it needs (GOCACHE, GOFLAGS,
// …) is plain Go toolchain configuration set in the Dockerfile.
const (
	// EnvPayload holds the base64 JSON payload. When it is unset the payload is
	// read from stdin instead, which is what makes the image usable by hand.
	EnvPayload = "GOFINITY_PAYLOAD"
	// EnvWorkdir is the scratch directory the workspace is written into. It
	// must be on the container's writable tmpfs.
	EnvWorkdir = "GOFINITY_WORKDIR"
	// EnvPrewarmedCache is the read-only build cache baked into the image.
	EnvPrewarmedCache = "GOFINITY_PREWARMED_CACHE"
)

const defaultWorkdir = "/tmp/work"

func main() {
	result := execute()
	emit(os.Stdout, result)
	if result.Error != "" {
		os.Exit(1)
	}
	// A failing test is a successful run of the runner. The caller reads `ok`,
	// not the exit code, to decide whether the submission passed.
	os.Exit(0)
}

// execute always returns a Result. Every failure path produces one too, so the
// caller can rely on there being exactly one JSON object on stdout no matter
// what went wrong.
func execute() *Result {
	payload, err := readPayload()
	if err != nil {
		return errorResult(err)
	}

	if err := WarmCache(os.Getenv(EnvPrewarmedCache), os.Getenv("GOCACHE")); err != nil {
		// Non-fatal: a cold cache is slow, not broken. Say so on stderr, where
		// it reaches the container log without polluting the result contract.
		fmt.Fprintf(os.Stderr, "gofinity-runner: could not warm the build cache: %v\n", err)
	}

	workdir := os.Getenv(EnvWorkdir)
	if workdir == "" {
		workdir = defaultWorkdir
	}
	if err := Materialize(workdir, payload.Files); err != nil {
		return errorResult(err)
	}

	outcome := RunCommand(workdir, payload.Command, InternalTimeout(payload.TimeoutMs))
	if outcome.StartErr != nil {
		return errorResult(fmt.Errorf("could not start %q: %w", payload.Command[0], outcome.StartErr))
	}
	return buildResult(outcome)
}

// buildResult merges what the parser knows (which tests ran) with what only the
// process knows (exit code, whether it was killed).
func buildResult(outcome *CommandOutcome) *Result {
	run := ParseGoTestJSON(outcome.Stdout)

	return &Result{
		OK:         outcome.ExitCode == 0 && !outcome.TimedOut && run.Failed == 0 && !run.BuildFailed,
		Passed:     run.Passed,
		Failed:     run.Failed,
		Skipped:    run.Skipped,
		Tests:      run.Tests,
		Stdout:     run.Stdout,
		Stderr:     outcome.Stderr,
		ExitCode:   outcome.ExitCode,
		TimedOut:   outcome.TimedOut,
		DurationMs: outcome.DurationMs,
	}
}

func readPayload() (*Payload, error) {
	if encoded, ok := os.LookupEnv(EnvPayload); ok {
		return DecodePayload(encoded)
	}
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("could not read the payload from stdin: %w", err)
	}
	payload, err := DecodePayload(string(raw))
	if errors.Is(err, ErrEmptyPayload) {
		return nil, fmt.Errorf("no payload: set %s or pipe one to stdin", EnvPayload)
	}
	return payload, err
}

// emit writes the framed result. Marshalling cannot fail for this struct, but
// if it somehow did, printing a hand-built error object beats printing nothing
// between the sentinels.
func emit(w io.Writer, result *Result) {
	encoded, err := json.Marshal(result)
	if err != nil {
		encoded = []byte(`{"ok":false,"error":"could not encode the result"}`)
	}
	fmt.Fprintf(w, "%s\n%s\n%s\n", BeginSentinel, encoded, EndSentinel)
}
