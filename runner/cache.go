package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// WarmCache seeds a writable GOCACHE from the read-only copy baked into the
// image.
//
// Why this exists: the container runs with a read-only root filesystem and a
// tmpfs scratch, so GOCACHE has to live on the tmpfs and starts empty on every
// run. Compiling the standard library from cold costs far more than the whole
// 10s budget, so the image pre-warms a cache at /opt/gocache and each run
// copies it in. The copy is a few hundred milliseconds against a tmpfs; a cold
// compile is tens of seconds.
//
// It is best-effort by design: a failure here makes the run slow, not wrong,
// and turning a slow run into no run at all would be the worse trade.
func WarmCache(source, target string) error {
	if source == "" || target == "" {
		return nil
	}
	if info, err := os.Stat(source); err != nil || !info.IsDir() {
		return nil
	}
	if !isEmptyDir(target) {
		return nil
	}
	return copyTree(source, target)
}

// isEmptyDir reports whether target is absent or an empty directory. Anything
// else (a populated cache, a file in the way) means "leave it alone".
func isEmptyDir(target string) bool {
	entries, err := os.ReadDir(target)
	if err != nil {
		return os.IsNotExist(err)
	}
	return len(entries) == 0
}

func copyTree(source, target string) error {
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(target, rel)

		if entry.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		// Cache entries are regular files. Anything else (a symlink pointing
		// out of the tree, say) is skipped rather than followed.
		if !entry.Type().IsRegular() {
			return nil
		}
		return copyFile(path, dest)
	})
}

func copyFile(source, dest string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return fmt.Errorf("copying %s: %w", source, err)
	}
	return out.Close()
}
