package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

// These sources mirror `01-hello-gofinity` in `packages/gofinity`, but are
// inlined deliberately: the runner is its own Go module and must not reach into
// another workspace to be testable.
const (
	goMod = "module gofinity/hello\n\ngo 1.24\n"

	testFile = `package main

import "testing"

func TestGreet(t *testing.T) {
	if got := Greet("Ada"); got != "Hello, Ada!" {
		t.Errorf("Greet(%q) = %q", "Ada", got)
	}
}
`

	solution = `package main

import "fmt"

func Greet(name string) string { return fmt.Sprintf("Hello, %s!", name) }

func main() { fmt.Println(Greet("Gofinity")) }
`

	starter = `package main

func Greet(name string) string { return "" }

func main() {}
`

	doesNotCompile = `package main

func Greet(name string) string { return 1 }

func main() {}
`

	infiniteLoop = `package main

func Greet(name string) string {
	for {
	}
}

func main() {}
`
)

// runEntrypoint drives the real entrypoint the way the container does: payload
// in the environment, workspace on disk, `go test` actually executed.
func runEntrypoint(t *testing.T, main string, timeoutMs int) *Result {
	t.Helper()

	payload := Payload{
		Files: []PayloadFile{
			{Path: "go.mod", Content: goMod},
			{Path: "main.go", Content: main},
			{Path: "main_test.go", Content: testFile},
		},
		TimeoutMs: timeoutMs,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv(EnvPayload, base64.StdEncoding.EncodeToString(raw))
	t.Setenv(EnvWorkdir, t.TempDir())
	return execute()
}

func TestEntrypointOnAPassingSolution(t *testing.T) {
	result := runEntrypoint(t, solution, DefaultTimeoutMs)

	if result.Error != "" {
		t.Fatalf("runner error: %s\n%s", result.Error, result.Stderr)
	}
	if !result.OK {
		t.Fatalf("ok = false for a correct solution: %+v", result)
	}
	if result.Passed == 0 || result.Failed != 0 {
		t.Errorf("passed/failed = %d/%d", result.Passed, result.Failed)
	}
	if result.ExitCode != 0 || result.TimedOut {
		t.Errorf("exitCode = %d, timedOut = %v", result.ExitCode, result.TimedOut)
	}
}

func TestEntrypointOnAFailingSolution(t *testing.T) {
	result := runEntrypoint(t, starter, DefaultTimeoutMs)

	if result.Error != "" {
		t.Fatalf("runner error: %s", result.Error)
	}
	if result.OK {
		t.Fatal("ok = true for code that fails its tests")
	}
	if result.Failed == 0 {
		t.Errorf("failed = 0 but the tests do not pass: %+v", result.Tests)
	}
	if !strings.Contains(result.Stdout, "Greet") {
		t.Errorf("stdout does not carry the assertion message: %q", result.Stdout)
	}
}

func TestEntrypointOnACompileError(t *testing.T) {
	result := runEntrypoint(t, doesNotCompile, DefaultTimeoutMs)

	if result.Error != "" {
		t.Fatalf("a compile error is the user's problem, not a runner error: %s", result.Error)
	}
	if result.OK {
		t.Fatal("ok = true for code that does not compile")
	}
	if !strings.Contains(result.Stdout, "cannot use 1") {
		t.Errorf("stdout does not carry the compiler error: %q", result.Stdout)
	}
}

func TestEntrypointKillsAnInfiniteLoop(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles and then waits out a timeout")
	}
	result := runEntrypoint(t, infiniteLoop, 4000)

	if !result.TimedOut {
		t.Fatalf("timedOut = false for a program that never returns: %+v", result)
	}
	if result.OK {
		t.Error("ok = true for a run that was killed")
	}
}

func TestEntrypointReportsABadPayloadAsARunnerError(t *testing.T) {
	t.Setenv(EnvPayload, base64.StdEncoding.EncodeToString([]byte(`{"files":[]}`)))

	result := execute()

	if result.Error == "" {
		t.Fatal("a payload with no files should be a runner error")
	}
	if result.OK {
		t.Error("ok = true on a runner error")
	}
	if result.Tests == nil {
		t.Error("tests is null - the consumer should never have to null-check it")
	}
}

// The stdout contract, exercised the way `apps/api` will read it in Phase 8.
func TestEmitFramesExactlyOneResult(t *testing.T) {
	var buf bytes.Buffer
	emit(&buf, &Result{OK: true, Passed: 1, Tests: []TestResult{}})

	out := buf.String()
	if strings.Count(out, BeginSentinel) != 1 || strings.Count(out, EndSentinel) != 1 {
		t.Fatalf("expected exactly one framed result, got:\n%s", out)
	}

	body, ok := extractResult(out)
	if !ok {
		t.Fatalf("could not extract the result from:\n%s", out)
	}
	var parsed Result
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("the framed body is not JSON: %v", err)
	}
	if !parsed.OK || parsed.Passed != 1 {
		t.Fatalf("result did not round-trip: %+v", parsed)
	}
}

// Stray output around the sentinels must not break extraction - that is the
// entire reason the framing exists.
func TestExtractResultIgnoresStrayOutput(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("panic: something printed before\n")
	emit(&buf, &Result{OK: false, Error: "boom"})
	buf.WriteString("and after\n")

	body, ok := extractResult(buf.String())
	if !ok {
		t.Fatal("extraction failed with stray output around the frame")
	}
	if !strings.Contains(body, `"boom"`) {
		t.Fatalf("extracted the wrong body: %q", body)
	}
}

// extractResult is the reference implementation of the documented contract:
// take the LAST begin sentinel, read to the next end sentinel.
func extractResult(out string) (string, bool) {
	start := strings.LastIndex(out, BeginSentinel)
	if start < 0 {
		return "", false
	}
	rest := out[start+len(BeginSentinel):]
	end := strings.Index(rest, EndSentinel)
	if end < 0 {
		return "", false
	}
	return strings.TrimSpace(rest[:end]), true
}
