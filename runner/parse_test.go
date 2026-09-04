package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The fixtures in testdata/ are real `go test -json` output, recorded by
// running the `01-hello-gofinity` challenge against its solution, its starter,
// a version that does not compile, and a skipped test. Hand-written fixtures
// would only ever prove the parser agrees with my memory of the format.
func fixture(t *testing.T, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("reading the fixture: %v", err)
	}
	return string(raw)
}

func TestParseGoTestJSONOnAPassingRun(t *testing.T) {
	run := ParseGoTestJSON(fixture(t, "gotest-pass.jsonl"))

	if run.Failed != 0 {
		t.Errorf("failed = %d, want 0", run.Failed)
	}
	if run.BuildFailed {
		t.Error("BuildFailed = true on a run that compiled")
	}
	// TestGreet plus its three subtests.
	if run.Passed != 4 {
		t.Errorf("passed = %d, want 4 (the test and its three subtests)", run.Passed)
	}

	names := testNames(run)
	for _, want := range []string{"TestGreet", "TestGreet/a_name", "TestGreet/empty_falls_back_to_World"} {
		if !contains(names, want) {
			t.Errorf("expected %q among the tests, got %v", want, names)
		}
	}
	if !strings.Contains(run.Stdout, "PASS") {
		t.Errorf("stdout does not look like test output: %q", run.Stdout)
	}
}

func TestParseGoTestJSONOnAFailingRun(t *testing.T) {
	run := ParseGoTestJSON(fixture(t, "gotest-fail.jsonl"))

	if run.Failed == 0 {
		t.Fatal("failed = 0 on a run where the starter code is wrong")
	}
	if run.BuildFailed {
		t.Error("BuildFailed = true on a run that compiled but failed its assertions")
	}

	var failing *TestResult
	for i := range run.Tests {
		if run.Tests[i].Status == StatusFailed && strings.Contains(run.Tests[i].Name, "/") {
			failing = &run.Tests[i]
			break
		}
	}
	if failing == nil {
		t.Fatal("expected at least one failing subtest")
	}
	// The assertion message is the whole point of showing a failure to a user.
	if !strings.Contains(failing.Output, "want") {
		t.Errorf("failing test output does not carry the assertion: %q", failing.Output)
	}
}

func TestParseGoTestJSONOnACompileError(t *testing.T) {
	run := ParseGoTestJSON(fixture(t, "gotest-build-fail.jsonl"))

	if !run.BuildFailed {
		t.Error("BuildFailed = false on output that contains a build-fail event")
	}
	if len(run.Tests) != 0 {
		t.Errorf("tests = %v, want none — nothing ran", testNames(run))
	}
	// The compiler's message must survive into stdout or the user is told
	// nothing at all about why their code did not build.
	if !strings.Contains(run.Stdout, "cannot use 1") {
		t.Errorf("stdout does not carry the compiler error: %q", run.Stdout)
	}
}

func TestParseGoTestJSONOnASkippedTest(t *testing.T) {
	run := ParseGoTestJSON(fixture(t, "gotest-skip.jsonl"))

	if run.Skipped != 1 {
		t.Errorf("skipped = %d, want 1", run.Skipped)
	}
	if run.Failed != 0 || run.Passed != 0 {
		t.Errorf("passed/failed = %d/%d, want 0/0", run.Passed, run.Failed)
	}
}

func TestParseGoTestJSONKeepsNonJSONLines(t *testing.T) {
	stream := "go: some toolchain notice\n" +
		`{"Action":"run","Package":"p","Test":"TestA"}` + "\n" +
		`{"Action":"pass","Package":"p","Test":"TestA","Elapsed":0.25}` + "\n" +
		"trailing garbage\n"

	run := ParseGoTestJSON(stream)

	if run.Passed != 1 {
		t.Fatalf("passed = %d, want 1", run.Passed)
	}
	if run.Tests[0].ElapsedMs != 250 {
		t.Errorf("elapsedMs = %d, want 250", run.Tests[0].ElapsedMs)
	}
	for _, want := range []string{"some toolchain notice", "trailing garbage"} {
		if !strings.Contains(run.Stdout, want) {
			t.Errorf("stdout dropped the non-JSON line %q: %q", want, run.Stdout)
		}
	}
}

// A run killed mid-test leaves a `run` event with no terminal action. Counting
// that as anything but a failure would report a timed-out submission as green.
func TestParseGoTestJSONCountsUnfinishedTestsAsFailed(t *testing.T) {
	run := ParseGoTestJSON(`{"Action":"run","Package":"p","Test":"TestSlow"}` + "\n")

	if run.Failed != 1 {
		t.Fatalf("failed = %d, want 1 for a test that never finished", run.Failed)
	}
	if run.Tests[0].Status != StatusFailed {
		t.Errorf("status = %q, want %q", run.Tests[0].Status, StatusFailed)
	}
}

func TestParseGoTestJSONOnEmptyInput(t *testing.T) {
	run := ParseGoTestJSON("")

	if len(run.Tests) != 0 || run.Passed != 0 || run.Failed != 0 || run.Stdout != "" {
		t.Fatalf("empty input produced %+v", run)
	}
}

// Two packages can each declare a test of the same name; keying only on the
// name would merge them into one entry.
func TestParseGoTestJSONKeepsSameNamedTestsInDifferentPackagesApart(t *testing.T) {
	stream := `{"Action":"pass","Package":"a","Test":"TestX"}` + "\n" +
		`{"Action":"fail","Package":"b","Test":"TestX"}` + "\n"

	run := ParseGoTestJSON(stream)

	if len(run.Tests) != 2 {
		t.Fatalf("tests = %d, want 2", len(run.Tests))
	}
	if run.Passed != 1 || run.Failed != 1 {
		t.Errorf("passed/failed = %d/%d, want 1/1", run.Passed, run.Failed)
	}
}

func testNames(run *ParsedRun) []string {
	names := make([]string, 0, len(run.Tests))
	for _, t := range run.Tests {
		names = append(names, t.Name)
	}
	return names
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
