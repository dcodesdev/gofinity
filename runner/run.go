package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// TimeoutMarginMs is how far the runner's own deadline sits below the timeout
// it was given. The API kills the container at `timeoutMs`; killing the process
// slightly earlier means a timed-out run still returns a real result — the
// partial output, the tests that did finish — instead of the API only knowing
// that the container disappeared.
const TimeoutMarginMs = 750

// WaitDelay is how long the process gets to die after SIGKILL before we stop
// waiting on its pipes. A process that ignores SIGKILL does not exist, but a
// grandchild holding stdout open does.
const WaitDelay = 2 * time.Second

// InternalTimeout converts the payload's outer budget into the runner's own,
// keeping the margin from swallowing a very short budget entirely.
func InternalTimeout(outerMs int) time.Duration {
	margin := TimeoutMarginMs
	if half := outerMs / 2; margin > half {
		margin = half
	}
	return time.Duration(outerMs-margin) * time.Millisecond
}

// CommandOutcome is the raw result of running the command, before any parsing.
type CommandOutcome struct {
	Stdout     string
	Stderr     string
	ExitCode   int
	TimedOut   bool
	DurationMs int64
	// StartErr is set when the command could not be started at all — a missing
	// `go` binary, an unexecutable workspace. It is a runner error, not a user
	// error.
	StartErr error
}

// RunCommand executes command in dir with the given budget, capturing stdout
// and stderr separately.
//
// The child is put in its own process group and the group is signalled on
// timeout: `go test` compiles and then execs a test binary, so killing only the
// direct child would leave an infinite loop running inside the container.
func RunCommand(dir string, command []string, timeout time.Duration) *CommandOutcome {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = nil
	cmd.Env = childEnv()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = WaitDelay
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		// Negative pid: the whole process group.
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}

	started := time.Now()
	err := cmd.Run()
	outcome := &CommandOutcome{
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		DurationMs: time.Since(started).Milliseconds(),
		TimedOut:   errors.Is(ctx.Err(), context.DeadlineExceeded),
	}

	switch {
	case err == nil:
		outcome.ExitCode = 0
	case isStartFailure(err):
		outcome.StartErr = err
		outcome.ExitCode = -1
	default:
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			outcome.ExitCode = exitErr.ExitCode()
		} else {
			outcome.ExitCode = -1
		}
	}
	return outcome
}

// isStartFailure separates "the command never ran" from "the command ran and
// failed", which are different things to report to a user.
func isStartFailure(err error) bool {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return false
	}
	return errors.Is(err, exec.ErrNotFound) || errors.Is(err, os.ErrNotExist)
}

// childEnvDefaults are applied only when the variable is not already set, so
// the image's Dockerfile stays the place these are configured and this stays
// the safety net for running the binary outside the image.
var childEnvDefaults = map[string]string{
	// No network exists in the container; `off` fails immediately with a clear
	// message instead of retrying a proxy that will never answer.
	"GOPROXY": "off",
	// Never try to download a different toolchain — see above.
	"GOTOOLCHAIN": "local",
	"GOFLAGS":     "-mod=mod",
	"CGO_ENABLED": "0",
}

func childEnv() []string {
	env := os.Environ()
	present := make(map[string]struct{}, len(env))
	for _, kv := range env {
		if k, _, ok := strings.Cut(kv, "="); ok {
			present[k] = struct{}{}
		}
	}
	for k, v := range childEnvDefaults {
		if _, ok := present[k]; !ok {
			env = append(env, k+"="+v)
		}
	}
	return env
}
