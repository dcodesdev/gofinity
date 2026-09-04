package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWarmCacheCopiesIntoAnEmptyTarget(t *testing.T) {
	source := t.TempDir()
	target := filepath.Join(t.TempDir(), "gocache")

	writeFile(t, filepath.Join(source, "ab", "abcdef-d"), "cached")

	if err := WarmCache(source, target); err != nil {
		t.Fatalf("WarmCache: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(target, "ab", "abcdef-d"))
	if err != nil {
		t.Fatalf("the cache entry was not copied: %v", err)
	}
	if string(got) != "cached" {
		t.Fatalf("content = %q", got)
	}
}

// The pre-warmed cache is a starting point, never an overwrite: a target that
// already has entries is a cache from earlier in this container's life.
func TestWarmCacheLeavesAPopulatedTargetAlone(t *testing.T) {
	source := t.TempDir()
	target := t.TempDir()
	writeFile(t, filepath.Join(source, "entry"), "from the image")
	writeFile(t, filepath.Join(target, "entry"), "already here")

	if err := WarmCache(source, target); err != nil {
		t.Fatalf("WarmCache: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(target, "entry"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "already here" {
		t.Fatalf("the existing cache was overwritten: %q", got)
	}
}

func TestWarmCacheIsANoOpWithoutASource(t *testing.T) {
	target := filepath.Join(t.TempDir(), "gocache")

	for _, source := range []string{"", filepath.Join(t.TempDir(), "missing")} {
		if err := WarmCache(source, target); err != nil {
			t.Fatalf("WarmCache(%q) = %v, want nil", source, err)
		}
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatal("WarmCache created the target with nothing to put in it")
	}
}

func TestWarmCacheIsANoOpWithoutATarget(t *testing.T) {
	if err := WarmCache(t.TempDir(), ""); err != nil {
		t.Fatalf("WarmCache with no GOCACHE = %v, want nil", err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
