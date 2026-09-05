package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestInternalTimeoutSitsBelowTheOuterKill(t *testing.T) {
	tests := []struct {
		outerMs int
		want    time.Duration
	}{
		{outerMs: DefaultTimeoutMs, want: 9250 * time.Millisecond},
		{outerMs: MaxTimeoutMs, want: 29250 * time.Millisecond},
		// Short budgets keep half of themselves rather than being consumed by
		// the margin - a 500ms budget must not become a 0ms one.
		{outerMs: 1000, want: 500 * time.Millisecond},
		{outerMs: MinTimeoutMs, want: 250 * time.Millisecond},
	}

	for _, tt := range tests {
		if got := InternalTimeout(tt.outerMs); got != tt.want {
			t.Errorf("InternalTimeout(%d) = %v, want %v", tt.outerMs, got, tt.want)
		}
		if InternalTimeout(tt.outerMs) >= time.Duration(tt.outerMs)*time.Millisecond {
			t.Errorf("InternalTimeout(%d) does not leave room before the outer kill", tt.outerMs)
		}
	}
}

func TestRunCommandCapturesStreamsSeparately(t *testing.T) {
	outcome := RunCommand(t.TempDir(), []string{"sh", "-c", "echo out; echo err >&2"}, time.Second)

	if outcome.StartErr != nil {
		t.Fatalf("StartErr = %v", outcome.StartErr)
	}
	if outcome.ExitCode != 0 {
		t.Errorf("exitCode = %d, want 0", outcome.ExitCode)
	}
	if strings.TrimSpace(outcome.Stdout) != "out" {
		t.Errorf("stdout = %q, want %q", outcome.Stdout, "out")
	}
	if strings.TrimSpace(outcome.Stderr) != "err" {
		t.Errorf("stderr = %q, want %q - the streams must not be merged", outcome.Stderr, "err")
	}
}

func TestRunCommandReportsTheExitCode(t *testing.T) {
	outcome := RunCommand(t.TempDir(), []string{"sh", "-c", "exit 3"}, time.Second)

	if outcome.ExitCode != 3 {
		t.Errorf("exitCode = %d, want 3", outcome.ExitCode)
	}
	if outcome.TimedOut {
		t.Error("TimedOut = true on a command that exited on its own")
	}
}

// The whole point of the process group: `go test` execs a test binary, so a
// runaway loop is a grandchild of the process we started.
func TestRunCommandKillsTheWholeProcessGroupOnTimeout(t *testing.T) {
	started := time.Now()
	outcome := RunCommand(t.TempDir(), []string{"sh", "-c", "sleep 30 & sleep 30"}, 300*time.Millisecond)
	elapsed := time.Since(started)

	if !outcome.TimedOut {
		t.Fatal("TimedOut = false on a command that outlived its budget")
	}
	if outcome.ExitCode == 0 {
		t.Error("exitCode = 0 on a killed command")
	}
	// Without the group kill, the backgrounded sleep would hold the pipes open
	// and this would take the full WaitDelay on top.
	if elapsed > 5*time.Second {
		t.Errorf("the kill took %v - a grandchild is probably still holding stdout", elapsed)
	}
}

func TestRunCommandReportsAMissingBinaryAsAStartFailure(t *testing.T) {
	outcome := RunCommand(t.TempDir(), []string{"definitely-not-a-real-binary"}, time.Second)

	if outcome.StartErr == nil {
		t.Fatal("StartErr = nil for a binary that does not exist")
	}
	if outcome.ExitCode != -1 {
		t.Errorf("exitCode = %d, want -1", outcome.ExitCode)
	}
}

func TestChildEnvAppliesDefaultsWithoutOverridingTheEnvironment(t *testing.T) {
	t.Setenv("GOPROXY", "https://example.invalid")

	env := childEnv()

	if !hasEnv(env, "GOPROXY=https://example.invalid") {
		t.Error("childEnv overrode an explicitly set GOPROXY")
	}
	for _, want := range []string{"GOTOOLCHAIN=local", "CGO_ENABLED=0", "GOFLAGS=-mod=mod"} {
		if !hasEnv(env, want) && !hasEnvKey(want) {
			t.Errorf("childEnv did not apply the default %q", want)
		}
	}
}

func hasEnv(env []string, kv string) bool {
	for _, entry := range env {
		if entry == kv {
			return true
		}
	}
	return false
}

// The default only applies when the ambient environment has not already set the
// key - which it may have, since `go test` itself runs with GOFLAGS set.
func hasEnvKey(kv string) bool {
	key, _, _ := strings.Cut(kv, "=")
	_, ok := os.LookupEnv(key)
	return ok
}
