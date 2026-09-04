package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidatePath enforces the same rules as `filePathSchema` in
// `packages/gofinity/src/schema.ts`. Both ends check, because either one could
// be the only one running: content is validated at seed time, submissions are
// validated at API time, and this is the last gate before a write.
func ValidatePath(p string) error {
	if p == "" {
		return errors.New("file path must not be empty")
	}
	if len(p) > MaxPathLen {
		return fmt.Errorf("%q: file path exceeds %d characters", p, MaxPathLen)
	}
	if strings.ContainsRune(p, 0) {
		return fmt.Errorf("%q: file path must not contain a NUL byte", p)
	}
	if strings.HasPrefix(p, "/") {
		return fmt.Errorf("%q: file path must be relative, not absolute", p)
	}
	if isWindowsDrivePath(p) {
		return fmt.Errorf("%q: file path must be relative, not absolute", p)
	}
	if strings.Contains(p, `\`) {
		return fmt.Errorf("%q: file path must use `/` as the separator", p)
	}
	for _, segment := range strings.Split(p, "/") {
		switch segment {
		case "":
			return fmt.Errorf("%q: file path must not contain an empty segment", p)
		case ".", "..":
			return fmt.Errorf("%q: file path must not contain a `.` or `..` segment", p)
		}
	}
	return nil
}

func isWindowsDrivePath(p string) bool {
	if len(p) < 3 {
		return false
	}
	c := p[0]
	isLetter := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
	return isLetter && p[1] == ':' && (p[2] == '/' || p[2] == '\\')
}

// ResolveInRoot turns a validated relative path into an absolute one and proves
// it did not escape. ValidatePath already makes traversal impossible; this is
// the belt to that pair of braces, and it is what would catch a future change
// that loosened the syntax rules.
func ResolveInRoot(root, p string) (string, error) {
	if err := ValidatePath(p); err != nil {
		return "", err
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	full := filepath.Join(absRoot, filepath.FromSlash(p))
	if full != absRoot && !strings.HasPrefix(full, absRoot+string(os.PathSeparator)) {
		return "", fmt.Errorf("%q: file path escapes the workspace", p)
	}
	return full, nil
}

// Materialize writes every payload file into root, creating parent directories
// as needed. Files are written 0o644 and directories 0o755: the runner user
// owns them and nothing else in the container needs to write them.
func Materialize(root string, files []PayloadFile) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("could not create the workspace: %w", err)
	}
	for _, f := range files {
		full, err := ResolveInRoot(root, f.Path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return fmt.Errorf("%q: could not create its directory: %w", f.Path, err)
		}
		if err := os.WriteFile(full, []byte(f.Content), 0o644); err != nil {
			return fmt.Errorf("%q: could not write it: %w", f.Path, err)
		}
	}
	return nil
}
