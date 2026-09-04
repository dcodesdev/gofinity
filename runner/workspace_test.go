package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatePathAcceptsWorkspaceRelativePaths(t *testing.T) {
	for _, p := range []string{"main.go", "go.mod", "pkg/thing.go", "a/b/c/d.go", "weird name.go", "..hidden.go"} {
		if err := ValidatePath(p); err != nil {
			t.Errorf("ValidatePath(%q) = %v, want nil", p, err)
		}
	}
}

func TestValidatePathRejectsAnythingThatCouldEscape(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"", "must not be empty"},
		{"/etc/passwd", "must be relative"},
		{"/", "must be relative"},
		{`C:\Windows\system32`, "must be relative"},
		{`C:/Windows`, "must be relative"},
		{`dir\file.go`, "`/` as the separator"},
		{"../escape.go", "`..` segment"},
		{"a/../../escape.go", "`..` segment"},
		{"a/./b.go", "`..` segment"},
		{"./main.go", "`..` segment"},
		{"a//b.go", "empty segment"},
		{"main.go/", "empty segment"},
		{"nul\x00.go", "NUL byte"},
		{strings.Repeat("a", MaxPathLen+1), "exceeds"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			err := ValidatePath(tt.path)
			if err == nil {
				t.Fatalf("ValidatePath(%q) = nil, want a rejection mentioning %q", tt.path, tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error %q does not mention %q", err, tt.want)
			}
		})
	}
}

func TestResolveInRootStaysInsideTheRoot(t *testing.T) {
	root := t.TempDir()

	full, err := ResolveInRoot(root, "pkg/thing.go")
	if err != nil {
		t.Fatalf("expected a valid path to resolve, got %v", err)
	}
	if !strings.HasPrefix(full, root+string(os.PathSeparator)) {
		t.Fatalf("resolved %q, which is not inside %q", full, root)
	}

	if _, err := ResolveInRoot(root, "../outside.go"); err == nil {
		t.Fatal("expected traversal to be rejected")
	}
}

// A sibling directory sharing the root's name prefix (`/tmp/workspace-evil`
// next to `/tmp/workspace`) is the classic way a prefix check goes wrong.
func TestResolveInRootIsNotFooledByASiblingPrefix(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "work")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}

	full, err := ResolveInRoot(root, "sub/file.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(full, filepath.Join(parent, "work-evil")) {
		t.Fatalf("resolved into a sibling directory: %q", full)
	}
}

func TestMaterializeWritesTheWorkspace(t *testing.T) {
	root := filepath.Join(t.TempDir(), "work")

	err := Materialize(root, []PayloadFile{
		{Path: "go.mod", Content: "module gofinity/hello\n\ngo 1.24\n"},
		{Path: "pkg/deep/thing.go", Content: "package deep\n"},
	})
	if err != nil {
		t.Fatalf("Materialize: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(root, "pkg", "deep", "thing.go"))
	if err != nil {
		t.Fatalf("nested file was not written: %v", err)
	}
	if string(got) != "package deep\n" {
		t.Fatalf("content = %q", got)
	}
}

func TestMaterializeRefusesToWriteOutsideTheRoot(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "work")

	err := Materialize(root, []PayloadFile{{Path: "../pwned.go", Content: "package main"}})
	if err == nil {
		t.Fatal("expected traversal to be rejected")
	}
	if _, statErr := os.Stat(filepath.Join(parent, "pwned.go")); statErr == nil {
		t.Fatal("a file was written outside the workspace root")
	}
}
